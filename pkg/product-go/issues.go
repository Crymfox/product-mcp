package product

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kidkuddy/product-go/parse"
)

// Issue represents a single entry from ISSUES.md.
type Issue struct {
	ID       string // "ISSUE-001"
	Title    string
	Type     string
	Severity string
	Status   string
	Source   string
	Effort   string
	Location string
	AuditRef string
	Domain   string
	Feature  string
	Body     string
	Fix      string

	// internal: which file this issue came from (for multi-file save)
	sourceFile string
}

// issuesFrontmatter is the YAML frontmatter for ISSUES.md.
type issuesFrontmatter struct {
	Schema  string `yaml:"schema"`
	Project string `yaml:"project"`
}

// ---- LoadIssues ------------------------------------------------------------

// LoadIssues loads ISSUES.md from dir. If an issues/ sub-directory exists its
// *.md files are also loaded and merged. Duplicate IDs across files return an
// error.
func LoadIssues(dir string) ([]Issue, error) {
	seen := make(map[string]string) // id -> source file
	var all []Issue

	// Primary file
	primary := filepath.Join(dir, "ISSUES.md")
	if _, err := os.Stat(primary); err == nil {
		issues, err := loadIssuesFile(primary)
		if err != nil {
			return nil, err
		}
		for _, iss := range issues {
			if prev, ok := seen[iss.ID]; ok {
				return nil, fmt.Errorf("LoadIssues: duplicate issue ID %s (in %s and %s)", iss.ID, prev, primary)
			}
			seen[iss.ID] = primary
			all = append(all, iss)
		}
	}

	// issues/ directory
	glob := filepath.Join(dir, "issues", "*.md")
	matches, _ := filepath.Glob(glob)
	for _, path := range matches {
		issues, err := loadIssuesFile(path)
		if err != nil {
			return nil, err
		}
		for _, iss := range issues {
			if prev, ok := seen[iss.ID]; ok {
				return nil, fmt.Errorf("LoadIssues: duplicate issue ID %s (in %s and %s)", iss.ID, prev, path)
			}
			seen[iss.ID] = path
			all = append(all, iss)
		}
	}

	return all, nil
}

var reIssueHeading = regexp.MustCompile(`^\[?(ISSUE-\d+)\]?\s+(.+)$`)

// loadIssuesFile parses a single issues file (ISSUES.md or issues/*.md).
func loadIssuesFile(path string) ([]Issue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loadIssuesFile %s: %w", path, err)
	}
	text := string(data)

	// Strip optional frontmatter
	var body string
	if strings.HasPrefix(text, "---\n") {
		var fm issuesFrontmatter
		b, err := parse.ParseFrontmatter(text, &fm)
		if err != nil {
			return nil, fmt.Errorf("loadIssuesFile %s: %w", path, err)
		}
		body = b
	} else {
		body = text
	}

	// Split on "---" separators (between issue blocks)
	// We normalise newlines so the separator "^---$" is clean.
	blocks := splitIssueBlocks(body)

	var issues []Issue
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		iss, err := parseIssueBlock(block)
		if err != nil {
			// skip malformed blocks silently
			continue
		}
		iss.sourceFile = path
		issues = append(issues, iss)
	}
	return issues, nil
}

// splitIssueBlocks splits on lines that are exactly "---".
func splitIssueBlocks(body string) []string {
	var blocks []string
	var cur strings.Builder
	for _, line := range strings.Split(body, "\n") {
		if strings.TrimSpace(line) == "---" {
			blocks = append(blocks, cur.String())
			cur.Reset()
		} else {
			cur.WriteString(line)
			cur.WriteByte('\n')
		}
	}
	if s := strings.TrimSpace(cur.String()); s != "" {
		blocks = append(blocks, s)
	}
	return blocks
}

// parseIssueBlock parses a single issue block starting with "## [ISSUE-NNN] Title".
func parseIssueBlock(block string) (Issue, error) {
	lines := strings.Split(block, "\n")
	if len(lines) == 0 {
		return Issue{}, fmt.Errorf("empty block")
	}

	// First line: ## [ISSUE-NNN] Title
	heading := strings.TrimPrefix(strings.TrimSpace(lines[0]), "## ")
	m := reIssueHeading.FindStringSubmatch(heading)
	if m == nil {
		return Issue{}, fmt.Errorf("bad issue heading: %q", lines[0])
	}
	iss := Issue{
		ID:    m[1],
		Title: strings.TrimSpace(m[2]),
	}

	// Remaining lines: field list + body + fix
	rest := strings.Join(lines[1:], "\n")
	iss.Body, iss.Fix = parseIssueBodyAndFix(rest)

	fields := parse.FieldMap(rest)
	iss.Type = fields["type"]
	iss.Severity = fields["severity"]
	iss.Status = fields["status"]
	iss.Source = fields["source"]
	iss.Effort = fields["effort"]
	iss.Location = strings.Trim(fields["location"], "`")
	iss.AuditRef = fields["audit-ref"]
	iss.Domain = fields["domain"]
	iss.Feature = fields["feature"]

	return iss, nil
}

// parseIssueBodyAndFix extracts the description body paragraph and the **fix**
// value from the remaining text after the heading.
func parseIssueBodyAndFix(text string) (body, fix string) {
	reFix := regexp.MustCompile(`(?m)^\*\*fix\*\*:\s*(.+)$`)
	if m := reFix.FindStringSubmatch(text); m != nil {
		fix = strings.TrimSpace(m[1])
	}

	// Body: lines that are not field-list lines ("**key**: value" or "- **key**: value") and not empty-only
	reField := regexp.MustCompile(`^(?:-\s+)?\*\*[\w][\w-]*\*\*:`)
	var bodyLines []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if reField.MatchString(trimmed) {
			continue
		}
		bodyLines = append(bodyLines, trimmed)
	}
	body = strings.Join(bodyLines, " ")
	return
}

// ---- SaveIssues ------------------------------------------------------------

// SaveIssues writes issues back. If issues were loaded from multiple files
// they are written back to each original file, but only if those files are
// under dir. Issues from outside dir (e.g. from a different testdata path)
// are written to dir/ISSUES.md. If all issues have no source file, writes
// dir/ISSUES.md.
func SaveIssues(dir string, issues []Issue) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	primary := filepath.Join(absDir, "ISSUES.md")

	// Group by destination file
	grouped := make(map[string][]Issue)
	for _, iss := range issues {
		sf := iss.sourceFile
		dest := primary
		if sf != "" {
			absSF, _ := filepath.Abs(sf)
			// Only use the original path if it lives under dir
			if isUnder(absSF, absDir) {
				dest = absSF
			}
		}
		grouped[dest] = append(grouped[dest], iss)
	}

	if len(grouped) == 0 {
		return writeIssuesFile(primary, "issues/1.0", "", nil)
	}

	for path, group := range grouped {
		project := ""
		// Try to read existing frontmatter for project name
		if existing, err := os.ReadFile(path); err == nil {
			var fm issuesFrontmatter
			if _, err := parse.ParseFrontmatter(string(existing), &fm); err == nil {
				project = fm.Project
			}
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := writeIssuesFile(path, "issues/1.0", project, group); err != nil {
			return err
		}
	}
	return nil
}

// isUnder reports whether path is under (or equal to) dir.
func isUnder(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

func writeIssuesFile(path, schema, project string, issues []Issue) error {
	var sb strings.Builder

	// Frontmatter
	type fm struct {
		Schema  string `yaml:"schema"`
		Project string `yaml:"project,omitempty"`
	}
	header, err := parse.MarshalFrontmatter(fm{Schema: schema, Project: project})
	if err != nil {
		return err
	}
	sb.WriteString(header)
	sb.WriteString("\n")

	for _, iss := range issues {
		sb.WriteString(fmt.Sprintf("## [%s] %s\n\n", iss.ID, iss.Title))
		if iss.Type != "" {
			sb.WriteString(fmt.Sprintf("- **type**: %s\n", iss.Type))
		}
		if iss.Severity != "" {
			sb.WriteString(fmt.Sprintf("- **severity**: %s\n", iss.Severity))
		}
		if iss.Status != "" {
			sb.WriteString(fmt.Sprintf("- **status**: %s\n", iss.Status))
		}
		if iss.Source != "" {
			sb.WriteString(fmt.Sprintf("- **source**: %s\n", iss.Source))
		}
		if iss.Effort != "" {
			sb.WriteString(fmt.Sprintf("- **effort**: %s\n", iss.Effort))
		}
		if iss.Location != "" {
			sb.WriteString(fmt.Sprintf("- **location**: `%s`\n", iss.Location))
		}
		if iss.AuditRef != "" {
			sb.WriteString(fmt.Sprintf("- **audit-ref**: %s\n", iss.AuditRef))
		}
		if iss.Domain != "" {
			sb.WriteString(fmt.Sprintf("- **domain**: %s\n", iss.Domain))
		}
		if iss.Feature != "" {
			sb.WriteString(fmt.Sprintf("- **feature**: %s\n", iss.Feature))
		}
		sb.WriteString("\n")
		if iss.Body != "" {
			sb.WriteString(iss.Body)
			sb.WriteString("\n\n")
		}
		if iss.Fix != "" {
			sb.WriteString(fmt.Sprintf("**fix**: %s\n\n", iss.Fix))
		}
		sb.WriteString("---\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}
