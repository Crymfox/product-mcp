// Package product provides parsing, querying, and writing of PRODUCT.md,
// ISSUES.md, and domains/*.md files following the PRODUCT.md spec.
package product

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kidkuddy/product-go/parse"
)

// ---- Types -----------------------------------------------------------------

// Product represents a parsed PRODUCT.md file.
type Product struct {
	Schema        string
	Name          string
	Description   string
	Version       string
	LastUpdated   string
	Vision        string
	Goals         []Goal
	TechStack     []TechRow
	Architecture  string
	Scopes        []Scope
	Domains       []Domain // inline domains merged with loaded domains/*.md
	OpenQuestions []string
	References    []string
}

// Goal is a single checkbox item in the ## Goals section.
type Goal struct {
	Slug        string
	Description string
	Done        bool
}

// TechRow is a row in the ## Tech Stack table.
type TechRow struct {
	Layer      string
	Technology string
}

// Scope is a row in the ## Scopes table.
type Scope struct {
	Name  string
	Path  string
	Type  string
	State string
}

// Domain is an inline or file-based domain block.
type Domain struct {
	Name     string
	Summary  string
	Files    []string
	Features []Feature
}

// Feature is a feature block inside a domain.
type Feature struct {
	Name       string
	State      string
	Why        string
	Acceptance []string
	DependsOn  []string
	Files      []string
	Notes      string
	Issues     []string // e.g. ["ISSUE-001", "ISSUE-002"]
}

// ---- frontmatter struct for PRODUCT.md -------------------------------------

type productFrontmatter struct {
	Schema      string `yaml:"schema"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	LastUpdated string `yaml:"last_updated"`
}

// ---- Load ------------------------------------------------------------------

// Load parses PRODUCT.md from the given directory.
// It also discovers and merges domains/*.md automatically.
func Load(dir string) (*Product, error) {
	data, err := os.ReadFile(filepath.Join(dir, "PRODUCT.md"))
	if err != nil {
		return nil, fmt.Errorf("product.Load: %w", err)
	}

	var fm productFrontmatter
	body, err := parse.ParseFrontmatter(string(data), &fm)
	if err != nil {
		return nil, fmt.Errorf("product.Load: %w", err)
	}

	p := &Product{
		Schema:      fm.Schema,
		Name:        fm.Name,
		Description: fm.Description,
		Version:     fm.Version,
		LastUpdated: fm.LastUpdated,
	}

	sections := parse.SplitSections(body)

	// Vision
	p.Vision = strings.TrimSpace(sections["vision"])

	// Goals
	if gs, ok := sections["goals"]; ok {
		items := parse.ParseCheckboxList(gs)
		for _, it := range items {
			slug, desc := parse.ParseGoalText(it.Text)
			p.Goals = append(p.Goals, Goal{
				Slug:        slug,
				Description: desc,
				Done:        it.Done,
			})
		}
	}

	// Tech Stack
	if ts, ok := sections["tech stack"]; ok {
		rows := parse.ParseTable(ts)
		for i, row := range rows {
			if i == 0 {
				continue // skip header
			}
			if len(row) >= 2 {
				p.TechStack = append(p.TechStack, TechRow{
					Layer:      row[0],
					Technology: row[1],
				})
			}
		}
	}

	// Architecture
	p.Architecture = strings.TrimSpace(sections["architecture"])

	// Scopes
	if sc, ok := sections["scopes"]; ok {
		rows := parse.ParseTable(sc)
		for i, row := range rows {
			if i == 0 {
				continue // skip header
			}
			if len(row) >= 4 {
				p.Scopes = append(p.Scopes, Scope{
					Name:  row[0],
					Path:  row[1],
					Type:  row[2],
					State: row[3],
				})
			}
		}
	}

	// Domains (inline)
	if ds, ok := sections["domains"]; ok {
		p.Domains = parseInlineDomains(ds)
	}

	// Open Questions
	if oq, ok := sections["open questions"]; ok {
		p.OpenQuestions = parse.ParseNumberedList(oq)
	}

	// References
	if refs, ok := sections["references"]; ok {
		p.References = parse.ParseBulletList(refs)
	}

	// Merge domains/*.md
	domainGlob := filepath.Join(dir, "domains", "*.md")
	matches, _ := filepath.Glob(domainGlob)
	for _, path := range matches {
		d, err := LoadDomain(path)
		if err != nil {
			return nil, fmt.Errorf("product.Load: domain %s: %w", path, err)
		}
		mergeDomain(p, d)
	}

	return p, nil
}

// parseInlineDomains parses the ## Domains section body which may contain
// ### Domain / #### Feature blocks.
func parseInlineDomains(text string) []Domain {
	h3blocks := parse.SplitH3Blocks(text)
	var domains []Domain
	for _, blk := range h3blocks {
		d := Domain{Name: blk.Heading}
		d.Summary, d.Files, d.Features = parseDomainBody(blk.Body, false)
		domains = append(domains, d)
	}
	return domains
}

// parseDomainBody parses the body of a domain block (either inline H3/H4 or
// file-based H2/H3). When fileLevel is true, features are H3 headings;
// otherwise they are H4.
func parseDomainBody(body string, fileLevel bool) (summary string, files []string, features []Feature) {
	var featureBlocks []parse.HeadingBlock
	if fileLevel {
		featureBlocks = parse.SplitH3Blocks(body)
	} else {
		featureBlocks = parse.SplitH4Blocks(body)
	}

	// Extract summary and domain-level files from the text before the first
	// feature heading.
	var preamble string
	if len(featureBlocks) == 0 {
		preamble = body
	} else {
		// Everything before the first feature heading
		marker := "\n#### "
		if fileLevel {
			marker = "\n### "
		}
		idx := strings.Index("\n"+body, marker)
		if idx == -1 {
			preamble = body
		} else {
			preamble = body[:idx]
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
			files = parse.ParseFileList(filesText)
		} else if line != "" {
			summaryLines = append(summaryLines, line)
		}
	}
	summary = strings.Join(summaryLines, " ")

	for _, fb := range featureBlocks {
		f := parseFeature(fb.Heading, fb.Body)
		features = append(features, f)
	}
	return
}

// parseFeature parses a feature block body into a Feature struct.
func parseFeature(name, body string) Feature {
	f := Feature{Name: name}

	// We need to carefully extract the acceptance list before FieldMap
	// flattens everything, because acceptance is a sub-list.
	f.Acceptance = parseAcceptanceFromBody(body)

	fields := parse.FieldMap(body)
	f.State = fields["state"]
	f.Why = fields["why"]
	f.Notes = fields["notes"]

	if dep, ok := fields["depends-on"]; ok {
		f.DependsOn = parse.ParseDependsList(dep)
	}
	if filesStr, ok := fields["files"]; ok {
		f.Files = parse.ParseFileList(filesStr)
	}
	if issStr, ok := fields["issues"]; ok {
		f.Issues = parse.ParseIssueRefs(issStr)
	}
	return f
}

// parseAcceptanceFromBody extracts the acceptance sub-list from a feature body.
// The acceptance block looks like:
//
//	- **acceptance**:
//	  - [ ] when X, then Y
func parseAcceptanceFromBody(body string) []string {
	lines := strings.Split(body, "\n")
	var items []string
	inAcceptance := false
	reAccItem := strings.NewReplacer()
	_ = reAccItem
	for _, line := range lines {
		if strings.TrimSpace(line) == "- **acceptance**:" {
			inAcceptance = true
			continue
		}
		if inAcceptance {
			trimmed := strings.TrimRight(line, " \t")
			// sub-list items are indented with 2+ spaces
			if strings.HasPrefix(trimmed, "  - [") {
				// strip the checkbox marker and extract text
				rest := strings.TrimLeft(trimmed, " ")
				// "- [ ] text" or "- [x] text"
				if len(rest) >= 6 {
					items = append(items, strings.TrimSpace(rest[6:]))
				}
			} else if strings.HasPrefix(line, "- **") || (strings.HasPrefix(strings.TrimSpace(line), "-") && !strings.HasPrefix(strings.TrimSpace(line), "- [")) {
				// New top-level field — stop acceptance parsing
				inAcceptance = false
			}
		}
	}
	return items
}

// normalizeName lowercases and replaces hyphens with spaces so that
// "rogue-memory" matches "Rogue Memory".
func normalizeName(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "-", " "))
}

// mergeDomain merges a domain loaded from a file into Product.Domains.
// Domain files take precedence over inline domains with the same name.
// Names are compared after normalizing hyphens to spaces (case-insensitive).
func mergeDomain(p *Product, d *Domain) {
	norm := normalizeName(d.Name)
	for i, existing := range p.Domains {
		if normalizeName(existing.Name) == norm {
			p.Domains[i] = *d
			return
		}
	}
	p.Domains = append(p.Domains, *d)
}

// ---- Save ------------------------------------------------------------------

// Save writes Product back to dir/PRODUCT.md, auto-updating LastUpdated to today.
func (p *Product) Save(dir string) error {
	p.LastUpdated = time.Now().Format("2006-01-02")

	fm := productFrontmatter{
		Schema:      p.Schema,
		Name:        p.Name,
		Description: p.Description,
		Version:     p.Version,
		LastUpdated: p.LastUpdated,
	}
	header, err := parse.MarshalFrontmatter(fm)
	if err != nil {
		return fmt.Errorf("product.Save: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")

	// Vision
	sb.WriteString("## Vision\n\n")
	sb.WriteString(strings.TrimSpace(p.Vision))
	sb.WriteString("\n\n")

	// Goals
	sb.WriteString("## Goals\n\n")
	for _, g := range p.Goals {
		mark := " "
		if g.Done {
			mark = "x"
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", mark, g.Slug, g.Description))
	}
	sb.WriteString("\n")

	// Tech Stack
	sb.WriteString("## Tech Stack\n\n")
	sb.WriteString("| Layer | Technology |\n")
	sb.WriteString("|-------|------------|\n")
	for _, r := range p.TechStack {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", r.Layer, r.Technology))
	}
	sb.WriteString("\n")

	// Architecture
	sb.WriteString("## Architecture\n\n")
	sb.WriteString(strings.TrimSpace(p.Architecture))
	sb.WriteString("\n\n")

	// Scopes
	sb.WriteString("## Scopes\n\n")
	sb.WriteString("| name | path | type | state |\n")
	sb.WriteString("|------|------|------|-------|\n")
	for _, s := range p.Scopes {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", s.Name, s.Path, s.Type, s.State))
	}
	sb.WriteString("\n")

	// Domains
	sb.WriteString("## Domains\n\n")
	for _, d := range p.Domains {
		sb.WriteString(fmt.Sprintf("### %s\n", d.Name))
		if d.Summary != "" {
			sb.WriteString(d.Summary)
			sb.WriteString("\n")
		}
		if len(d.Files) > 0 {
			sb.WriteString("**files**: ")
			var fl []string
			for _, f := range d.Files {
				fl = append(fl, "`"+f+"`")
			}
			sb.WriteString(strings.Join(fl, ", "))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
		for _, feat := range d.Features {
			sb.WriteString(serializeFeature(feat, "#### "))
		}
	}

	// Open Questions
	sb.WriteString("## Open Questions\n\n")
	for i, q := range p.OpenQuestions {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, q))
	}
	sb.WriteString("\n")

	// References
	sb.WriteString("## References\n\n")
	for _, r := range p.References {
		sb.WriteString(fmt.Sprintf("- %s\n", r))
	}
	sb.WriteString("\n")

	return os.WriteFile(filepath.Join(dir, "PRODUCT.md"), []byte(sb.String()), 0644)
}

// serializeFeature renders a Feature into markdown with the given heading prefix.
func serializeFeature(f Feature, prefix string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s%s\n", prefix, f.Name))
	if f.State != "" {
		sb.WriteString(fmt.Sprintf("- **state**: %s\n", f.State))
	}
	if f.Why != "" {
		sb.WriteString(fmt.Sprintf("- **why**: %s\n", f.Why))
	}
	if len(f.Acceptance) > 0 {
		sb.WriteString("- **acceptance**:\n")
		for _, a := range f.Acceptance {
			sb.WriteString(fmt.Sprintf("  - [ ] %s\n", a))
		}
	}
	if len(f.DependsOn) > 0 {
		sb.WriteString(fmt.Sprintf("- **depends-on**: %s\n", strings.Join(f.DependsOn, ", ")))
	}
	if len(f.Files) > 0 {
		var fl []string
		for _, file := range f.Files {
			fl = append(fl, "`"+file+"`")
		}
		sb.WriteString(fmt.Sprintf("- **files**: %s\n", strings.Join(fl, ", ")))
	}
	if f.Notes != "" {
		sb.WriteString(fmt.Sprintf("- **notes**: %s\n", f.Notes))
	}
	if len(f.Issues) > 0 {
		var il []string
		for _, iss := range f.Issues {
			il = append(il, "["+iss+"]")
		}
		sb.WriteString(fmt.Sprintf("- **issues**: %s\n", strings.Join(il, ", ")))
	}
	sb.WriteString("\n")
	return sb.String()
}
