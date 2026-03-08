package parse

import (
	"regexp"
	"strings"
)

// SplitSections splits a markdown body on "## " headings.
// Returns a map of lower-cased heading text -> section body (without the heading line).
func SplitSections(body string) map[string]string {
	sections := make(map[string]string)
	// Ensure we have a leading newline so the split works for a section at the
	// very start of the body.
	normalized := "\n" + body
	parts := strings.Split(normalized, "\n## ")
	// parts[0] is text before the first ## heading (usually empty)
	for _, part := range parts[1:] {
		nl := strings.Index(part, "\n")
		var heading, content string
		if nl == -1 {
			heading = part
			content = ""
		} else {
			heading = part[:nl]
			content = part[nl+1:]
		}
		sections[strings.ToLower(strings.TrimSpace(heading))] = strings.TrimRight(content, "\n")
	}
	return sections
}

// ParseTable parses a GFM-style markdown table. It returns a slice of rows
// where each row is a slice of trimmed cell strings. The header separator row
// (containing only dashes and pipes) is skipped. The header row itself is
// returned as the first row.
func ParseTable(text string) [][]string {
	var rows [][]string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, "|") {
			continue
		}
		// Check if this is a separator row (only -, |, :, space)
		stripped := strings.ReplaceAll(line, "|", "")
		stripped = strings.ReplaceAll(stripped, "-", "")
		stripped = strings.ReplaceAll(stripped, ":", "")
		stripped = strings.ReplaceAll(stripped, " ", "")
		if stripped == "" {
			continue
		}
		// Split on | and trim each cell
		cells := strings.Split(line, "|")
		var row []string
		for _, c := range cells {
			c = strings.TrimSpace(c)
			if c != "" {
				row = append(row, c)
			}
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}
	return rows
}

var (
	reCheckbox = regexp.MustCompile(`^- \[([ xX])\] (.+)$`)
	reGoalSlug = regexp.MustCompile(`^([a-zA-Z0-9_-]+):\s*(.+)$`)
)

// CheckboxItem represents a parsed checkbox list item.
type CheckboxItem struct {
	Done bool
	Text string
}

// ParseCheckboxList parses a markdown checkbox list. Each item is "- [ ] text"
// or "- [x] text".
func ParseCheckboxList(text string) []CheckboxItem {
	var items []CheckboxItem
	for _, line := range strings.Split(text, "\n") {
		m := reCheckbox.FindStringSubmatch(strings.TrimRight(line, " \t"))
		if m == nil {
			continue
		}
		items = append(items, CheckboxItem{
			Done: m[1] == "x" || m[1] == "X",
			Text: strings.TrimSpace(m[2]),
		})
	}
	return items
}

// ParseGoalText splits "slug: description" into its two parts.
// If the format doesn't match it returns ("", text).
func ParseGoalText(text string) (slug, description string) {
	m := reGoalSlug.FindStringSubmatch(text)
	if m == nil {
		return "", text
	}
	return m[1], m[2]
}

// FieldMap parses a flat field list. It handles both:
//   - "**key**: value"  (bare, used in issue blocks)
//   - "- **key**: value" (list item, used in feature blocks)
//
// It returns a map of key -> value.
func FieldMap(text string) map[string]string {
	// Match optional leading "- " then **key**: value
	reField := regexp.MustCompile(`^(?:-\s+)?\*\*([\w][\w-]*)\*\*:\s*(.*)$`)
	m := make(map[string]string)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, " \t")
		match := reField.FindStringSubmatch(line)
		if match != nil {
			m[match[1]] = strings.TrimSpace(match[2])
		}
	}
	return m
}

// ParseNumberedList parses a markdown numbered list "1. text", "2. text", …
func ParseNumberedList(text string) []string {
	re := regexp.MustCompile(`^\d+\.\s+(.+)$`)
	var items []string
	for _, line := range strings.Split(text, "\n") {
		m := re.FindStringSubmatch(strings.TrimSpace(line))
		if m != nil {
			items = append(items, m[1])
		}
	}
	return items
}

// ParseBulletList parses a simple "- item" or "* item" list.
func ParseBulletList(text string) []string {
	re := regexp.MustCompile(`^[-*]\s+(.+)$`)
	var items []string
	for _, line := range strings.Split(text, "\n") {
		m := re.FindStringSubmatch(strings.TrimSpace(line))
		if m != nil {
			items = append(items, m[1])
		}
	}
	return items
}

// SplitH3Blocks splits a section body on "### " headings.
// Returns a slice of (heading, body) pairs.
func SplitH3Blocks(text string) []HeadingBlock {
	return splitHeadingBlocks(text, "\n### ")
}

// SplitH1Blocks splits a section body on "# " headings.
func SplitH1Blocks(text string) []HeadingBlock {
	return splitHeadingBlocks(text, "\n# ")
}

// SplitH2Blocks splits a section body on "## " headings.
func SplitH2Blocks(text string) []HeadingBlock {
	return splitHeadingBlocks(text, "\n## ")
}

// SplitH4Blocks splits a section body on "#### " headings.
func SplitH4Blocks(text string) []HeadingBlock {
	return splitHeadingBlocks(text, "\n#### ")
}

// HeadingBlock holds a heading and its associated body.
type HeadingBlock struct {
	Heading string
	Body    string
}

func splitHeadingBlocks(text, delimiter string) []HeadingBlock {
	normalized := "\n" + text
	parts := strings.Split(normalized, delimiter)
	var blocks []HeadingBlock
	for _, part := range parts[1:] {
		nl := strings.Index(part, "\n")
		var heading, body string
		if nl == -1 {
			heading = part
			body = ""
		} else {
			heading = part[:nl]
			body = part[nl+1:]
		}
		blocks = append(blocks, HeadingBlock{
			Heading: strings.TrimSpace(heading),
			Body:    strings.TrimRight(body, "\n"),
		})
	}
	return blocks
}

// ParseAcceptanceList parses acceptance criteria lines indented under
// "- **acceptance**:". Lines look like "  - [ ] …" or "  - [x] …".
func ParseAcceptanceList(text string) []string {
	re := regexp.MustCompile(`^\s+- \[[ xX]\] (.+)$`)
	var items []string
	for _, line := range strings.Split(text, "\n") {
		m := re.FindStringSubmatch(line)
		if m != nil {
			items = append(items, strings.TrimSpace(m[1]))
		}
	}
	return items
}

// ParseIssueRefs extracts ISSUE-NNN identifiers from a string like
// "[ISSUE-001], [ISSUE-002]".
func ParseIssueRefs(text string) []string {
	re := regexp.MustCompile(`\[(ISSUE-\d+)\]`)
	matches := re.FindAllStringSubmatch(text, -1)
	var ids []string
	for _, m := range matches {
		ids = append(ids, m[1])
	}
	return ids
}

// ParseFileList extracts backtick-quoted file paths from a string like
// "`path/file.go`, `other.go`".
func ParseFileList(text string) []string {
	re := regexp.MustCompile("`([^`]+)`")
	matches := re.FindAllStringSubmatch(text, -1)
	var files []string
	for _, m := range matches {
		files = append(files, m[1])
	}
	return files
}

// ParseDependsList splits a comma-or-newline separated list of dependency names.
func ParseDependsList(text string) []string {
	var deps []string
	for _, part := range strings.Split(text, ",") {
		d := strings.TrimSpace(part)
		if d != "" {
			deps = append(deps, d)
		}
	}
	return deps
}
