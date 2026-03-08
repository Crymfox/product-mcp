package product_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	product "github.com/kidkuddy/product-go"
)

func TestLoadIssues_Count(t *testing.T) {
	issues, err := product.LoadIssues(testdataDir)
	if err != nil {
		t.Fatalf("LoadIssues: %v", err)
	}
	if len(issues) != 4 {
		t.Fatalf("len(issues) = %d, want 4", len(issues))
	}
}

func TestLoadIssues_Fields(t *testing.T) {
	issues, err := product.LoadIssues(testdataDir)
	if err != nil {
		t.Fatalf("LoadIssues: %v", err)
	}

	iss := issues[0]
	if iss.ID != "ISSUE-001" {
		t.Errorf("ID = %q, want %q", iss.ID, "ISSUE-001")
	}
	if iss.Title == "" {
		t.Error("Title is empty")
	}
	if iss.Type != "bug" {
		t.Errorf("Type = %q, want %q", iss.Type, "bug")
	}
	if iss.Severity != "high" {
		t.Errorf("Severity = %q, want %q", iss.Severity, "high")
	}
	if iss.Status != "open" {
		t.Errorf("Status = %q, want %q", iss.Status, "open")
	}
	if iss.Source != "manual" {
		t.Errorf("Source = %q, want %q", iss.Source, "manual")
	}
	if iss.Effort != "S" {
		t.Errorf("Effort = %q, want %q", iss.Effort, "S")
	}
	if iss.Domain != "cli" {
		t.Errorf("Domain = %q, want %q", iss.Domain, "cli")
	}
	if iss.Feature != "add-task" {
		t.Errorf("Feature = %q, want %q", iss.Feature, "add-task")
	}
	if iss.Body == "" {
		t.Error("Body is empty")
	}
	if iss.Fix == "" {
		t.Error("Fix is empty")
	}
}

func TestLoadIssues_OptionalFields(t *testing.T) {
	issues, err := product.LoadIssues(testdataDir)
	if err != nil {
		t.Fatalf("LoadIssues: %v", err)
	}

	// ISSUE-001 has no location or audit-ref
	iss001 := findIssue(issues, "ISSUE-001")
	if iss001 == nil {
		t.Fatal("ISSUE-001 not found")
	}
	if iss001.Location != "" {
		t.Errorf("ISSUE-001 Location = %q, want empty", iss001.Location)
	}
	if iss001.AuditRef != "" {
		t.Errorf("ISSUE-001 AuditRef = %q, want empty", iss001.AuditRef)
	}

	// ISSUE-003 has location and audit-ref
	iss003 := findIssue(issues, "ISSUE-003")
	if iss003 == nil {
		t.Fatal("ISSUE-003 not found")
	}
	if iss003.Location == "" {
		t.Error("ISSUE-003 Location is empty, want non-empty")
	}
	if strings.Contains(iss003.Location, "`") {
		t.Errorf("ISSUE-003 Location contains backticks: %q", iss003.Location)
	}
	if iss003.AuditRef == "" {
		t.Error("ISSUE-003 AuditRef is empty, want non-empty")
	}
	if iss003.Status != "closed" {
		t.Errorf("ISSUE-003 Status = %q, want %q", iss003.Status, "closed")
	}
}

func TestLoadIssues_DuplicateID(t *testing.T) {
	dir := t.TempDir()
	content := `---
schema: issues/1.0
project: test
---

## [ISSUE-001] First

- **type**: bug
- **severity**: high
- **status**: open
- **source**: manual
- **effort**: S

Some body.

**fix**: Fix it.

---

## [ISSUE-001] Duplicate

- **type**: bug
- **severity**: low
- **status**: open
- **source**: manual
- **effort**: S

Another body.

**fix**: Fix that too.

---
`
	if err := os.WriteFile(filepath.Join(dir, "ISSUES.md"), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, err := product.LoadIssues(dir)
	if err == nil {
		t.Error("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error should mention 'duplicate', got: %v", err)
	}
}

func TestLoadIssues_MergesIssuesDir(t *testing.T) {
	dir := t.TempDir()

	// Write primary ISSUES.md
	primary := `---
schema: issues/1.0
project: test
---

## [ISSUE-001] First

- **type**: bug
- **severity**: high
- **status**: open
- **source**: manual
- **effort**: S

Body one.

**fix**: Fix one.

---
`
	if err := os.WriteFile(filepath.Join(dir, "ISSUES.md"), []byte(primary), 0644); err != nil {
		t.Fatalf("WriteFile primary: %v", err)
	}

	// Write issues/ sub-file
	issuesDir := filepath.Join(dir, "issues")
	if err := os.MkdirAll(issuesDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	secondary := `---
schema: issues/1.0
project: test
---

## [ISSUE-002] Second

- **type**: feature
- **severity**: medium
- **status**: open
- **source**: manual
- **effort**: M

Body two.

**fix**: Fix two.

---
`
	if err := os.WriteFile(filepath.Join(issuesDir, "more.md"), []byte(secondary), 0644); err != nil {
		t.Fatalf("WriteFile secondary: %v", err)
	}

	issues, err := product.LoadIssues(dir)
	if err != nil {
		t.Fatalf("LoadIssues: %v", err)
	}
	if len(issues) != 2 {
		t.Errorf("len(issues) = %d, want 2", len(issues))
	}
}

func TestSaveIssues_RoundTrip(t *testing.T) {
	issues, err := product.LoadIssues(testdataDir)
	if err != nil {
		t.Fatalf("LoadIssues: %v", err)
	}

	dir := t.TempDir()
	// Clear sourceFile so all are written to ISSUES.md
	plain := make([]product.Issue, len(issues))
	copy(plain, issues)

	if err := product.SaveIssues(dir, plain); err != nil {
		t.Fatalf("SaveIssues: %v", err)
	}

	reloaded, err := product.LoadIssues(dir)
	if err != nil {
		t.Fatalf("re-LoadIssues: %v", err)
	}
	if len(reloaded) != len(issues) {
		t.Errorf("reloaded count = %d, want %d", len(reloaded), len(issues))
	}

	for _, orig := range issues {
		found := findIssue(reloaded, orig.ID)
		if found == nil {
			t.Errorf("issue %s not found after round-trip", orig.ID)
			continue
		}
		if found.Title != orig.Title {
			t.Errorf("%s Title: got %q, want %q", orig.ID, found.Title, orig.Title)
		}
		if found.Type != orig.Type {
			t.Errorf("%s Type: got %q, want %q", orig.ID, found.Type, orig.Type)
		}
		if found.Status != orig.Status {
			t.Errorf("%s Status: got %q, want %q", orig.ID, found.Status, orig.Status)
		}
	}
}

func TestLoadIssues_NoFile(t *testing.T) {
	// A directory with no ISSUES.md and no issues/ dir should return empty slice.
	dir := t.TempDir()
	issues, err := product.LoadIssues(dir)
	if err != nil {
		t.Fatalf("LoadIssues on empty dir: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

// findIssue is a test helper.
func findIssue(issues []product.Issue, id string) *product.Issue {
	for i := range issues {
		if issues[i].ID == id {
			return &issues[i]
		}
	}
	return nil
}
