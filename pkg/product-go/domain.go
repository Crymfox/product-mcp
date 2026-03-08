package product

import (
	"fmt"
	"os"
	"strings"

	"github.com/kidkuddy/product-go/parse"
)

// domainFrontmatter is the YAML frontmatter for domains/*.md.
type domainFrontmatter struct {
	Schema  string `yaml:"schema"`
	Name    string `yaml:"name"`
	Product string `yaml:"product"`
}

// LoadDomain parses a single domains/{name}.md file.
func LoadDomain(path string) (*Domain, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("LoadDomain %s: %w", path, err)
	}

	var fm domainFrontmatter
	body, err := parse.ParseFrontmatter(string(data), &fm)
	if err != nil {
		return nil, fmt.Errorf("LoadDomain %s: %w", path, err)
	}

	d := &Domain{}

	// Domain files may use H1 (# Domain) with H2 features, or H2 (## Domain)
	// with H3 features. Try H1 first since it's the natural standalone format.
	h1blocks := parse.SplitH1Blocks(body)
	if len(h1blocks) > 0 {
		blk := h1blocks[0]
		d.Name = blk.Heading
		// Features are at H2 level; parseDomainBody with fileLevel=true uses H3,
		// so we split H2 blocks ourselves.
		featureBlocks := parse.SplitH2Blocks(blk.Body)

		// Extract preamble (text before first H2 feature)
		preamble := blk.Body
		if len(featureBlocks) > 0 {
			idx := strings.Index("\n"+blk.Body, "\n## ")
			if idx >= 0 {
				preamble = blk.Body[:idx]
			}
		}

		// Parse summary and files from preamble
		lines := strings.Split(strings.TrimSpace(preamble), "\n")
		var summaryLines []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			lower := strings.ToLower(line)
			if strings.HasPrefix(lower, "**files**:") {
				filesText := line[len("**files**:"):]
				d.Files = parse.ParseFileList(filesText)
			} else if line != "" {
				summaryLines = append(summaryLines, line)
			}
		}
		d.Summary = strings.Join(summaryLines, " ")

		for _, fb := range featureBlocks {
			f := parseFeature(fb.Heading, fb.Body)
			d.Features = append(d.Features, f)
		}
	} else {
		// Fallback: H2 domain heading with H3 features.
		h2blocks := parse.SplitH2Blocks(body)
		if len(h2blocks) == 0 {
			d.Name = fm.Name
			return d, nil
		}
		blk := h2blocks[0]
		d.Name = blk.Heading
		d.Summary, d.Files, d.Features = parseDomainBody(blk.Body, true)
	}

	// If no name from heading, use frontmatter.
	if d.Name == "" {
		d.Name = fm.Name
	}

	return d, nil
}

// SaveDomain writes a Domain to the given path.
func SaveDomain(path string, d *Domain) error {
	fm := domainFrontmatter{
		Schema: "domain/1.0",
		Name:   strings.ToLower(strings.ReplaceAll(d.Name, " ", "-")),
	}
	header, err := parse.MarshalFrontmatter(fm)
	if err != nil {
		return fmt.Errorf("SaveDomain %s: %w", path, err)
	}

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("# %s\n\n", d.Name))

	if d.Summary != "" {
		sb.WriteString(d.Summary)
		sb.WriteString("\n")
	}
	if len(d.Files) > 0 {
		var fl []string
		for _, f := range d.Files {
			fl = append(fl, "`"+f+"`")
		}
		sb.WriteString("**files**: ")
		sb.WriteString(strings.Join(fl, ", "))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	for _, feat := range d.Features {
		sb.WriteString(serializeFeature(feat, "## "))
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(dirOf(path), 0755); err != nil {
		return fmt.Errorf("SaveDomain %s: %w", path, err)
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func dirOf(path string) string {
	idx := strings.LastIndexByte(path, '/')
	if idx == -1 {
		return "."
	}
	return path[:idx]
}
