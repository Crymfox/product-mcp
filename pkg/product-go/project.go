package product

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Project aggregates a loaded PRODUCT.md with all its issues and domains.
type Project struct {
	Dir     string
	Product *Product
	Issues  []Issue
}

// Open loads a full project: PRODUCT.md + all issues + all domains.
// Returns an error if PRODUCT.md does not exist in dir.
func Open(dir string) (*Project, error) {
	productPath := filepath.Join(dir, "PRODUCT.md")
	if _, err := os.Stat(productPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Open: no PRODUCT.md found in %s", dir)
	}
	p, err := Load(dir)
	if err != nil {
		return nil, fmt.Errorf("Open: %w", err)
	}
	issues, err := LoadIssues(dir)
	if err != nil {
		return nil, fmt.Errorf("Open: %w", err)
	}
	return &Project{
		Dir:     dir,
		Product: p,
		Issues:  issues,
	}, nil
}

// Save writes all state back to disk, auto-updating last_updated.
func (p *Project) Save() error {
	if err := p.Product.Save(p.Dir); err != nil {
		return err
	}
	return SaveIssues(p.Dir, p.Issues)
}

// ---- Query API -------------------------------------------------------------

// IssuesByStatus returns issues matching the given status.
func (p *Project) IssuesByStatus(status string) []Issue {
	var out []Issue
	for _, iss := range p.Issues {
		if strings.EqualFold(iss.Status, status) {
			out = append(out, iss)
		}
	}
	return out
}

// IssuesByType returns issues matching the given type.
func (p *Project) IssuesByType(t string) []Issue {
	var out []Issue
	for _, iss := range p.Issues {
		if strings.EqualFold(iss.Type, t) {
			out = append(out, iss)
		}
	}
	return out
}

// IssuesByDomain returns issues matching the given domain.
func (p *Project) IssuesByDomain(domain string) []Issue {
	var out []Issue
	for _, iss := range p.Issues {
		if strings.EqualFold(iss.Domain, domain) {
			out = append(out, iss)
		}
	}
	return out
}

// IssuesBySeverity returns issues matching the given severity.
func (p *Project) IssuesBySeverity(severity string) []Issue {
	var out []Issue
	for _, iss := range p.Issues {
		if strings.EqualFold(iss.Severity, severity) {
			out = append(out, iss)
		}
	}
	return out
}

// OpenIssues returns all issues with status "open" or "in-progress".
func (p *Project) OpenIssues() []Issue {
	var out []Issue
	for _, iss := range p.Issues {
		s := strings.ToLower(iss.Status)
		if s == "open" || s == "in-progress" {
			out = append(out, iss)
		}
	}
	return out
}

// GetIssue returns the issue with the given ID, or nil if not found.
func (p *Project) GetIssue(id string) *Issue {
	for i := range p.Issues {
		if p.Issues[i].ID == id {
			return &p.Issues[i]
		}
	}
	return nil
}

// Domain returns the domain with the given name, or nil.
func (p *Project) Domain(name string) *Domain {
	for i := range p.Product.Domains {
		if strings.EqualFold(p.Product.Domains[i].Name, name) {
			return &p.Product.Domains[i]
		}
	}
	return nil
}

// Feature returns the feature with the given name in the given domain, or nil.
func (p *Project) Feature(domain, feature string) *Feature {
	d := p.Domain(domain)
	if d == nil {
		return nil
	}
	for i := range d.Features {
		if strings.EqualFold(d.Features[i].Name, feature) {
			return &d.Features[i]
		}
	}
	return nil
}

// ---- Mutate API ------------------------------------------------------------

// AddIssue appends an issue. Returns error if ID already exists.
func (p *Project) AddIssue(issue Issue) error {
	for _, existing := range p.Issues {
		if existing.ID == issue.ID {
			return fmt.Errorf("AddIssue: ID %s already exists", issue.ID)
		}
	}
	p.Issues = append(p.Issues, issue)
	return nil
}

// UpdateIssue finds the issue by ID and calls fn on it. Returns error if not found.
func (p *Project) UpdateIssue(id string, fn func(*Issue)) error {
	for i := range p.Issues {
		if p.Issues[i].ID == id {
			fn(&p.Issues[i])
			return nil
		}
	}
	return fmt.Errorf("UpdateIssue: issue %s not found", id)
}

// CloseIssue sets status to "closed" on the given issue ID.
func (p *Project) CloseIssue(id string) error {
	return p.UpdateIssue(id, func(iss *Issue) {
		iss.Status = "closed"
	})
}

// UpdateFeatureState sets the state field on the named feature in the named domain.
func (p *Project) UpdateFeatureState(domainName, featureName, state string) error {
	f := p.Feature(domainName, featureName)
	if f == nil {
		return fmt.Errorf("UpdateFeatureState: feature %q in domain %q not found", featureName, domainName)
	}
	f.State = state
	return nil
}

// SetGoalDone marks a goal as done (or not done) by slug.
func (p *Project) SetGoalDone(slug string, done bool) error {
	for i := range p.Product.Goals {
		if p.Product.Goals[i].Slug == slug {
			p.Product.Goals[i].Done = done
			return nil
		}
	}
	return fmt.Errorf("SetGoalDone: goal %q not found", slug)
}
