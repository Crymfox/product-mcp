// Package parse provides low-level markdown and YAML frontmatter parsing
// utilities used by the product-go library.
package parse

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFrontmatter splits a markdown document into its YAML frontmatter and
// body. The document must begin with "---\n". Returns an error if the opening
// or closing delimiter is missing.
func ParseFrontmatter(src string, out any) (body string, err error) {
	if !strings.HasPrefix(src, "---\n") {
		return "", fmt.Errorf("parse: document does not start with '---'")
	}
	rest := src[4:] // skip opening "---\n"
	idx := strings.Index(rest, "\n---\n")
	if idx == -1 {
		// try end-of-string terminator
		if strings.HasSuffix(strings.TrimRight(rest, "\n"), "\n---") {
			idx = strings.LastIndex(rest, "\n---")
		}
		if idx == -1 {
			return "", fmt.Errorf("parse: closing '---' not found")
		}
	}
	fm := rest[:idx]
	body = rest[idx+5:] // skip "\n---\n"
	if err := yaml.Unmarshal([]byte(fm), out); err != nil {
		return "", fmt.Errorf("parse: frontmatter YAML: %w", err)
	}
	return body, nil
}

// MarshalFrontmatter serialises v as YAML wrapped in "---\n...\n---\n".
func MarshalFrontmatter(v any) (string, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return "---\n" + string(b) + "---\n", nil
}
