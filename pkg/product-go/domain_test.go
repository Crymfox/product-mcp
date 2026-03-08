package product_test

import (
	"path/filepath"
	"strings"
	"testing"

	product "github.com/kidkuddy/product-go"
)

func TestLoadDomain_Frontmatter(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "storage.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	if d.Name == "" {
		t.Error("Domain.Name is empty")
	}
	if !strings.EqualFold(d.Name, "storage") {
		t.Errorf("Domain.Name = %q, want %q (case-insensitive)", d.Name, "Storage")
	}
}

func TestLoadDomain_Summary(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "storage.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	if d.Summary == "" {
		t.Error("Domain.Summary is empty")
	}
}

func TestLoadDomain_Files(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "storage.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	if len(d.Files) == 0 {
		t.Error("Domain.Files is empty")
	}
}

func TestLoadDomain_Features(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "storage.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	if len(d.Features) != 2 {
		t.Fatalf("len(Features) = %d, want 2", len(d.Features))
	}

	f := d.Features[0]
	if f.Name != "load-store" {
		t.Errorf("Features[0].Name = %q, want %q", f.Name, "load-store")
	}
	if f.State != "ready" {
		t.Errorf("Features[0].State = %q, want %q", f.State, "ready")
	}
	if f.Why == "" {
		t.Error("Features[0].Why is empty")
	}
	if len(f.Acceptance) != 2 {
		t.Errorf("Features[0].Acceptance count = %d, want 2", len(f.Acceptance))
	}
}

func TestLoadDomain_FeatureDependsOn(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "storage.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	f := d.Features[1]
	if f.Name != "save-store" {
		t.Errorf("Features[1].Name = %q, want %q", f.Name, "save-store")
	}
	if len(f.DependsOn) == 0 {
		t.Error("Features[1].DependsOn is empty")
	}
	if f.DependsOn[0] != "load-store" {
		t.Errorf("DependsOn[0] = %q, want %q", f.DependsOn[0], "load-store")
	}
}

func TestLoadDomain_FeatureIssues(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "storage.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	f := d.Features[1]
	if len(f.Issues) == 0 {
		t.Error("save-store feature Issues is empty, want [ISSUE-003]")
	}
	if f.Issues[0] != "ISSUE-003" {
		t.Errorf("Issues[0] = %q, want %q", f.Issues[0], "ISSUE-003")
	}
}

func TestSaveDomain_RoundTrip(t *testing.T) {
	orig, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "storage.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}

	dir := t.TempDir()
	outPath := filepath.Join(dir, "domains", "storage.md")
	if err := product.SaveDomain(outPath, orig); err != nil {
		t.Fatalf("SaveDomain: %v", err)
	}

	reloaded, err := product.LoadDomain(outPath)
	if err != nil {
		t.Fatalf("re-LoadDomain: %v", err)
	}

	if reloaded.Name != orig.Name {
		t.Errorf("Name: got %q, want %q", reloaded.Name, orig.Name)
	}
	if reloaded.Summary != orig.Summary {
		t.Errorf("Summary: got %q, want %q", reloaded.Summary, orig.Summary)
	}
	if len(reloaded.Features) != len(orig.Features) {
		t.Errorf("Features count: got %d, want %d", len(reloaded.Features), len(orig.Features))
	}
	for i, f := range orig.Features {
		if i >= len(reloaded.Features) {
			break
		}
		rf := reloaded.Features[i]
		if rf.Name != f.Name {
			t.Errorf("Features[%d].Name: got %q, want %q", i, rf.Name, f.Name)
		}
		if rf.State != f.State {
			t.Errorf("Features[%d].State: got %q, want %q", i, rf.State, f.State)
		}
		if len(rf.Acceptance) != len(f.Acceptance) {
			t.Errorf("Features[%d].Acceptance count: got %d, want %d", i, len(rf.Acceptance), len(f.Acceptance))
		}
	}
}

func TestLoadDomain_MissingFile(t *testing.T) {
	_, err := product.LoadDomain("/nonexistent/domain.md")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// ---- H1 domain file format (# Domain with ## Features) --------------------

func TestLoadDomain_H1Format_Name(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "memory.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	if d.Name != "Memory" {
		t.Errorf("Domain.Name = %q, want %q", d.Name, "Memory")
	}
}

func TestLoadDomain_H1Format_Summary(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "memory.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	if d.Summary == "" {
		t.Error("Domain.Summary is empty")
	}
}

func TestLoadDomain_H1Format_Files(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "memory.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	if len(d.Files) != 2 {
		t.Errorf("len(Domain.Files) = %d, want 2", len(d.Files))
	}
}

func TestLoadDomain_H1Format_Features(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "memory.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	if len(d.Features) != 2 {
		t.Fatalf("len(Features) = %d, want 2", len(d.Features))
	}

	f := d.Features[0]
	if f.Name != "sql-query" {
		t.Errorf("Features[0].Name = %q, want %q", f.Name, "sql-query")
	}
	if f.State != "ready" {
		t.Errorf("Features[0].State = %q, want %q", f.State, "ready")
	}
	if len(f.Acceptance) != 2 {
		t.Errorf("Features[0].Acceptance count = %d, want 2", len(f.Acceptance))
	}
	if f.Notes == "" {
		t.Error("Features[0].Notes is empty")
	}
}

func TestLoadDomain_H1Format_FeatureDependsOn(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "memory.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	f := d.Features[1]
	if f.Name != "file-read" {
		t.Errorf("Features[1].Name = %q, want %q", f.Name, "file-read")
	}
	if len(f.DependsOn) == 0 {
		t.Error("Features[1].DependsOn is empty")
	}
	if f.DependsOn[0] != "sql-query" {
		t.Errorf("DependsOn[0] = %q, want %q", f.DependsOn[0], "sql-query")
	}
}

func TestLoadDomain_H1Format_FeatureIssues(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "memory.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	f := d.Features[1]
	if len(f.Issues) == 0 {
		t.Error("file-read feature Issues is empty, want [ISSUE-002]")
	}
	if f.Issues[0] != "ISSUE-002" {
		t.Errorf("Issues[0] = %q, want %q", f.Issues[0], "ISSUE-002")
	}
}

func TestLoadDomain_Notes(t *testing.T) {
	d, err := product.LoadDomain(filepath.Join(testdataDir, "domains", "storage.md"))
	if err != nil {
		t.Fatalf("LoadDomain: %v", err)
	}
	f := d.Features[0] // load-store has notes
	if f.Notes == "" {
		t.Error("load-store feature Notes is empty")
	}
}
