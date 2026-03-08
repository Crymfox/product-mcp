![Go Version](https://img.shields.io/badge/go-1.23+-blue)
![License](https://img.shields.io/badge/license-MIT-green)

# product-go

Go library for reading, querying, and writing [PRODUCT.md](https://github.com/kidkuddy/PRODUCT.md) projects. Reference implementation of the spec.

Handles `PRODUCT.md`, `ISSUES.md`, `issues/*.md`, and `domains/*.md`.

## Install

```
go get github.com/kidkuddy/product-go
```

## Quick start

```go
import product "github.com/kidkuddy/product-go"

proj, err := product.Open(".")
if err != nil {
    panic(err)
}

// Query
fmt.Println(proj.Product.Name)
for _, iss := range proj.OpenIssues() {
    fmt.Printf("[%s] %s\n", iss.ID, iss.Title)
}

// Mutate
proj.UpdateFeatureState("storage", "write-task", "ready")
proj.CloseIssue("ISSUE-003")
proj.SetGoalDone("mvp", true)

// Write back — auto-updates last_updated
proj.Save()
```

## API

### Load

| Function | Description |
|----------|-------------|
| `Open(dir string) (*Project, error)` | Load PRODUCT.md + issues + domains. Errors if PRODUCT.md missing. |
| `Load(dir string) (*Product, error)` | Parse PRODUCT.md and merge domains/*.md. |
| `LoadIssues(dir string) ([]Issue, error)` | Load ISSUES.md and issues/*.md. |
| `LoadDomain(path string) (*Domain, error)` | Parse a single domains/{name}.md. |

### Save

| Function | Description |
|----------|-------------|
| `(*Product).Save(dir string) error` | Write PRODUCT.md, auto-bumping last_updated. |
| `SaveIssues(dir string, issues []Issue) error` | Write issues back to their source files. |
| `SaveDomain(path string, d *Domain) error` | Write a domain file. |
| `(*Project).Save() error` | Write all changes back to disk. |

### Query

| Method | Description |
|--------|-------------|
| `OpenIssues() []Issue` | Issues with status open or in-progress. |
| `IssuesByStatus(status string) []Issue` | Filter by status. |
| `IssuesByType(t string) []Issue` | Filter by type. |
| `IssuesByDomain(domain string) []Issue` | Filter by domain. |
| `IssuesBySeverity(severity string) []Issue` | Filter by severity. |
| `GetIssue(id string) *Issue` | Lookup by ID. |
| `Domain(name string) *Domain` | Lookup domain by name. |
| `Feature(domain, feature string) *Feature` | Lookup feature by domain and name. |

### Mutate

| Method | Description |
|--------|-------------|
| `AddIssue(issue Issue) error` | Append issue; error if ID exists. |
| `UpdateIssue(id string, fn func(*Issue)) error` | Update issue in-place. |
| `CloseIssue(id string) error` | Set status to closed. |
| `UpdateFeatureState(domain, feature, state string) error` | Update a feature's state. |
| `SetGoalDone(slug string, done bool) error` | Mark a goal done or undone. |

## Spec

[github.com/kidkuddy/PRODUCT.md](https://github.com/kidkuddy/PRODUCT.md)

## License

MIT
