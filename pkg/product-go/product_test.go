package product_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	product "github.com/kidkuddy/product-go"
)

const testdataDir = "example/testdata"

func TestLoad_Frontmatter(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.Schema != "product/1.0" {
		t.Errorf("Schema = %q, want %q", p.Schema, "product/1.0")
	}
	if p.Name != "taskr" {
		t.Errorf("Name = %q, want %q", p.Name, "taskr")
	}
	if p.Description == "" {
		t.Error("Description is empty")
	}
	if p.Version != "0.3.0" {
		t.Errorf("Version = %q, want %q", p.Version, "0.3.0")
	}
	if p.LastUpdated != "2025-01-15" {
		t.Errorf("LastUpdated = %q, want %q", p.LastUpdated, "2025-01-15")
	}
}

func TestLoad_Vision(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !strings.Contains(p.Vision, "terminal") {
		t.Errorf("Vision missing expected content, got: %q", p.Vision)
	}
}

func TestLoad_Goals(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(p.Goals) != 3 {
		t.Fatalf("len(Goals) = %d, want 3", len(p.Goals))
	}

	// First two should be not done
	if p.Goals[0].Done {
		t.Errorf("Goals[0].Done = true, want false")
	}
	if p.Goals[0].Slug != "add-task" {
		t.Errorf("Goals[0].Slug = %q, want %q", p.Goals[0].Slug, "add-task")
	}
	if p.Goals[0].Description == "" {
		t.Error("Goals[0].Description is empty")
	}

	// Third goal should be done
	if !p.Goals[2].Done {
		t.Errorf("Goals[2].Done = false, want true")
	}
	if p.Goals[2].Slug != "persistence" {
		t.Errorf("Goals[2].Slug = %q, want %q", p.Goals[2].Slug, "persistence")
	}
}

func TestLoad_TechStack(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(p.TechStack) != 2 {
		t.Fatalf("len(TechStack) = %d, want 2", len(p.TechStack))
	}
	if p.TechStack[0].Layer != "CLI" {
		t.Errorf("TechStack[0].Layer = %q, want %q", p.TechStack[0].Layer, "CLI")
	}
	if p.TechStack[0].Technology != "cobra" {
		t.Errorf("TechStack[0].Technology = %q, want %q", p.TechStack[0].Technology, "cobra")
	}
}

func TestLoad_Architecture(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !strings.Contains(p.Architecture, "cobra") {
		t.Errorf("Architecture missing expected content, got: %q", p.Architecture)
	}
}

func TestLoad_Scopes(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(p.Scopes) != 2 {
		t.Fatalf("len(Scopes) = %d, want 2", len(p.Scopes))
	}
	if p.Scopes[0].Name != "cli" {
		t.Errorf("Scopes[0].Name = %q, want %q", p.Scopes[0].Name, "cli")
	}
	if p.Scopes[0].Path != "cmd/" {
		t.Errorf("Scopes[0].Path = %q, want %q", p.Scopes[0].Path, "cmd/")
	}
	if p.Scopes[0].Type != "package" {
		t.Errorf("Scopes[0].Type = %q, want %q", p.Scopes[0].Type, "package")
	}
	if p.Scopes[0].State != "active" {
		t.Errorf("Scopes[0].State = %q, want %q", p.Scopes[0].State, "active")
	}
}

func TestLoad_InlineDomains(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// After merging with domains/*.md we should have CLI + Storage
	var cliDomain *product.Domain
	for i := range p.Domains {
		if strings.EqualFold(p.Domains[i].Name, "CLI") {
			cliDomain = &p.Domains[i]
			break
		}
	}
	if cliDomain == nil {
		t.Fatal("CLI domain not found")
	}
	if !strings.Contains(cliDomain.Summary, "command") {
		t.Errorf("CLI summary = %q, expected to contain 'command'", cliDomain.Summary)
	}
	if len(cliDomain.Files) == 0 {
		t.Error("CLI domain has no files")
	}
}

func TestLoad_Features(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	var cliDomain *product.Domain
	for i := range p.Domains {
		if strings.EqualFold(p.Domains[i].Name, "CLI") {
			cliDomain = &p.Domains[i]
			break
		}
	}
	if cliDomain == nil {
		t.Fatal("CLI domain not found")
	}
	if len(cliDomain.Features) != 2 {
		t.Fatalf("CLI features count = %d, want 2", len(cliDomain.Features))
	}

	feat := cliDomain.Features[0]
	if feat.Name != "add-task" {
		t.Errorf("Features[0].Name = %q, want %q", feat.Name, "add-task")
	}
	if feat.State != "building" {
		t.Errorf("Features[0].State = %q, want %q", feat.State, "building")
	}
	if feat.Why == "" {
		t.Error("Features[0].Why is empty")
	}
	if len(feat.Acceptance) != 2 {
		t.Errorf("Features[0].Acceptance count = %d, want 2", len(feat.Acceptance))
	}
	if len(feat.Files) == 0 {
		t.Error("Features[0].Files is empty")
	}
}

func TestLoad_FeatureIssues(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	var feat *product.Feature
	for i := range p.Domains {
		if strings.EqualFold(p.Domains[i].Name, "CLI") {
			for j := range p.Domains[i].Features {
				if p.Domains[i].Features[j].Name == "add-task" {
					feat = &p.Domains[i].Features[j]
				}
			}
		}
	}
	if feat == nil {
		t.Fatal("add-task feature not found")
	}
	if len(feat.Issues) != 2 {
		t.Errorf("Issues count = %d, want 2", len(feat.Issues))
	}
	if feat.Issues[0] != "ISSUE-001" {
		t.Errorf("Issues[0] = %q, want %q", feat.Issues[0], "ISSUE-001")
	}
}

func TestLoad_DomainsFileMerge(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	var storageDomain *product.Domain
	for i := range p.Domains {
		if strings.EqualFold(p.Domains[i].Name, "Storage") {
			storageDomain = &p.Domains[i]
			break
		}
	}
	if storageDomain == nil {
		t.Fatal("Storage domain not found (should be auto-loaded from domains/storage.md)")
	}
	if len(storageDomain.Features) != 2 {
		t.Errorf("Storage features count = %d, want 2", len(storageDomain.Features))
	}
}

func TestLoad_H1DomainFileMerge(t *testing.T) {
	// The memory.md domain file uses H1 format (# Memory with ## features).
	// The inline PRODUCT.md has "### memory" (no features, just a stub).
	// The file domain should replace the inline stub.
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	var memDomain *product.Domain
	memCount := 0
	for i := range p.Domains {
		n := strings.ToLower(p.Domains[i].Name)
		if n == "memory" {
			memDomain = &p.Domains[i]
			memCount++
		}
	}
	if memDomain == nil {
		t.Fatal("Memory domain not found (should be auto-loaded from domains/memory.md)")
	}
	if memCount != 1 {
		t.Errorf("found %d memory domains, want 1 (file should replace inline)", memCount)
	}
	if len(memDomain.Features) != 2 {
		t.Errorf("Memory features count = %d, want 2", len(memDomain.Features))
	}
	if memDomain.Features[0].Name != "sql-query" {
		t.Errorf("Features[0].Name = %q, want %q", memDomain.Features[0].Name, "sql-query")
	}
}

func TestLoad_OpenQuestions(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(p.OpenQuestions) != 2 {
		t.Errorf("OpenQuestions count = %d, want 2", len(p.OpenQuestions))
	}
}

func TestLoad_References(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(p.References) == 0 {
		t.Error("References is empty")
	}
}

func TestSave_UpdatesLastUpdated(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	dir := t.TempDir()
	// Copy domains dir (not needed for this test, but keep it consistent)
	if err := p.Save(dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	today := time.Now().Format("2006-01-02")
	if p.LastUpdated != today {
		t.Errorf("LastUpdated = %q, want today %q", p.LastUpdated, today)
	}

	// Re-load and verify
	p2, err := product.Load(dir)
	if err != nil {
		t.Fatalf("Re-Load: %v", err)
	}
	if p2.LastUpdated != today {
		t.Errorf("Re-loaded LastUpdated = %q, want %q", p2.LastUpdated, today)
	}
}

func TestSave_RoundTrip(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	dir := t.TempDir()
	if err := p.Save(dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	p2, err := product.Load(dir)
	if err != nil {
		t.Fatalf("Re-Load: %v", err)
	}

	if p2.Name != p.Name {
		t.Errorf("Name: got %q, want %q", p2.Name, p.Name)
	}
	if p2.Schema != p.Schema {
		t.Errorf("Schema: got %q, want %q", p2.Schema, p.Schema)
	}
	if len(p2.Goals) != len(p.Goals) {
		t.Errorf("Goals count: got %d, want %d", len(p2.Goals), len(p.Goals))
	}
	if len(p2.TechStack) != len(p.TechStack) {
		t.Errorf("TechStack count: got %d, want %d", len(p2.TechStack), len(p.TechStack))
	}
	if len(p2.Scopes) != len(p.Scopes) {
		t.Errorf("Scopes count: got %d, want %d", len(p2.Scopes), len(p.Scopes))
	}
	if len(p2.OpenQuestions) != len(p.OpenQuestions) {
		t.Errorf("OpenQuestions count: got %d, want %d", len(p2.OpenQuestions), len(p.OpenQuestions))
	}

	// Check goal done-ness round-trips
	for i, g := range p.Goals {
		if i >= len(p2.Goals) {
			break
		}
		if p2.Goals[i].Done != g.Done {
			t.Errorf("Goals[%d].Done: got %v, want %v", i, p2.Goals[i].Done, g.Done)
		}
		if p2.Goals[i].Slug != g.Slug {
			t.Errorf("Goals[%d].Slug: got %q, want %q", i, p2.Goals[i].Slug, g.Slug)
		}
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := product.Load("/nonexistent/dir")
	if err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestSave_CreatesFile(t *testing.T) {
	p, err := product.Load(testdataDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	dir := t.TempDir()
	if err := p.Save(dir); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "PRODUCT.md")); err != nil {
		t.Errorf("PRODUCT.md not created: %v", err)
	}
}
