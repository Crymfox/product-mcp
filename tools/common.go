package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	product "github.com/kidkuddy/product-go"
)

// discoverProjectRoot searches upwards from the start path for a PRODUCT.md file.
func discoverProjectRoot(startPath string) (string, error) {
	curr, err := filepath.Abs(startPath)
	if err != nil {
		curr = startPath
	}

	for {
		if _, err := os.Stat(filepath.Join(curr, "PRODUCT.md")); err == nil {
			return curr, nil
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return "", fmt.Errorf("no PRODUCT.md found in %s or any parent directory", startPath)
}

// getProject loads the product-go project. If providedPath is empty, it attempts
// to discover the project root from the current working directory.
func getProject(providedPath string) (*product.Project, error) {
	root := providedPath
	if root == "" || root == "." {
		cwd, _ := os.Getwd()
		discovered, err := discoverProjectRoot(cwd)
		if err != nil {
			return nil, err
		}
		root = discovered
	}
	
	return product.Open(root)
}

// findDomainFile looks for a domain file in domains/ matching the name.
// Returns the absolute path if found, or empty string.
func findDomainFile(dir, name string) string {
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	expected := filepath.Join(dir, "domains", slug+".md")

	// Check exact match first
	if _, err := os.Stat(expected); err == nil {
		return expected
	}

	// Fallback: iterate over all files to check for case-insensitive match
	glob := filepath.Join(dir, "domains", "*.md")
	matches, _ := filepath.Glob(glob)
	for _, m := range matches {
		base := filepath.Base(m)
		ext := filepath.Ext(base)
		namePart := base[:len(base)-len(ext)]
		// Check if namePart matches slug (ignoring case)
		if strings.EqualFold(namePart, slug) {
			return m
		}
	}

	return ""
}

// SmartSave saves the project state, automatically determining whether a domain
// should be saved inline in PRODUCT.md or to a separate domains/*.md file based
// on the current file system structure.
func SmartSave(p *product.Project, projectPath string) error {
	// Save issues
	if err := product.SaveIssues(projectPath, p.Issues); err != nil {
		return fmt.Errorf("failed to save issues: %w", err)
	}

	// Save domains and PRODUCT.md
	// We need to modify p.Product.Domains temporarily for PRODUCT.md generation
	// But p.Product is a pointer? No, p.Product is *Product.
	// p.Product.Domains is []Domain.
	originalDomains := make([]product.Domain, len(p.Product.Domains))
	copy(originalDomains, p.Product.Domains)

	// We'll modify p.Product.Domains in place
	// But we need to be careful not to mutate the original structs if we want to restore them exactly.
	// Actually, we are modifying the slice elements.
	// Let's iterate and replace.

	for i := range p.Product.Domains {
		d := &p.Product.Domains[i]
		path := findDomainFile(projectPath, d.Name)
		if path != "" {
			// It is an external domain. Save it to the file.
			// d is a pointer to the domain in the slice.
			// We save the full domain content to the file.
			if err := product.SaveDomain(path, d); err != nil {
				// Restore before returning error
				p.Product.Domains = originalDomains
				return fmt.Errorf("failed to save domain %s: %w", d.Name, err)
			}

			// Replace in PRODUCT.md with a link/reference
			// We clear features and files so they don't get inlined.
			slug := filepath.Base(path)
			// Create a lightweight copy for the inline version
			inlineDomain := *d
			inlineDomain.Features = nil
			inlineDomain.Files = nil
			inlineDomain.Summary = fmt.Sprintf("See [%s](domains/%s).", slug, slug)
			
			// Update the slice to use the inline version
			p.Product.Domains[i] = inlineDomain
		}
	}

	// Save PRODUCT.md
	err := p.Product.Save(projectPath)

	// Restore domains
	p.Product.Domains = originalDomains

	if err != nil {
		return fmt.Errorf("failed to save PRODUCT.md: %w", err)
	}

	return nil
}
