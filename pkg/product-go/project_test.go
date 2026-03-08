package product_test

import (
	"strings"
	"testing"
	"time"

	product "github.com/kidkuddy/product-go"
)

func TestOpen(t *testing.T) {
	proj, err := product.Open(testdataDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if proj.Product == nil {
		t.Fatal("Project.Product is nil")
	}
	if proj.Product.Name != "taskr" {
		t.Errorf("Product.Name = %q, want %q", proj.Product.Name, "taskr")
	}
	if len(proj.Issues) != 4 {
		t.Errorf("Issues count = %d, want 4", len(proj.Issues))
	}
}

// ---- Query methods ---------------------------------------------------------

func TestIssuesByStatus(t *testing.T) {
	proj := mustOpen(t)
	open := proj.IssuesByStatus("open")
	if len(open) != 3 {
		t.Errorf("IssuesByStatus(open) = %d, want 3", len(open))
	}
	closed := proj.IssuesByStatus("closed")
	if len(closed) != 1 {
		t.Errorf("IssuesByStatus(closed) = %d, want 1", len(closed))
	}
}

func TestIssuesByType(t *testing.T) {
	proj := mustOpen(t)
	bugs := proj.IssuesByType("bug")
	if len(bugs) != 1 {
		t.Errorf("IssuesByType(bug) = %d, want 1", len(bugs))
	}
	tasks := proj.IssuesByType("task")
	if len(tasks) != 1 {
		t.Errorf("IssuesByType(task) = %d, want 1", len(tasks))
	}
}

func TestIssuesByDomain(t *testing.T) {
	proj := mustOpen(t)
	cli := proj.IssuesByDomain("cli")
	if len(cli) != 2 {
		t.Errorf("IssuesByDomain(cli) = %d, want 2", len(cli))
	}
	storage := proj.IssuesByDomain("storage")
	if len(storage) != 1 {
		t.Errorf("IssuesByDomain(storage) = %d, want 1", len(storage))
	}
}

func TestIssuesBySeverity(t *testing.T) {
	proj := mustOpen(t)
	high := proj.IssuesBySeverity("high")
	if len(high) != 2 {
		t.Errorf("IssuesBySeverity(high) = %d, want 2", len(high))
	}
	low := proj.IssuesBySeverity("low")
	if len(low) != 1 {
		t.Errorf("IssuesBySeverity(low) = %d, want 1", len(low))
	}
}

func TestOpenIssues(t *testing.T) {
	proj := mustOpen(t)
	open := proj.OpenIssues()
	if len(open) != 3 {
		t.Errorf("OpenIssues() = %d, want 3", len(open))
	}
	for _, iss := range open {
		s := strings.ToLower(iss.Status)
		if s != "open" && s != "in-progress" {
			t.Errorf("OpenIssues() returned issue with status %q", iss.Status)
		}
	}
}

func TestGetIssue(t *testing.T) {
	proj := mustOpen(t)
	iss := proj.GetIssue("ISSUE-002")
	if iss == nil {
		t.Fatal("GetIssue(ISSUE-002) = nil")
	}
	if iss.ID != "ISSUE-002" {
		t.Errorf("ID = %q, want %q", iss.ID, "ISSUE-002")
	}

	missing := proj.GetIssue("ISSUE-999")
	if missing != nil {
		t.Error("GetIssue(ISSUE-999) should return nil")
	}
}

func TestDomain(t *testing.T) {
	proj := mustOpen(t)
	d := proj.Domain("Storage")
	if d == nil {
		t.Fatal("Domain(Storage) = nil")
	}
	if len(d.Features) != 2 {
		t.Errorf("Storage features = %d, want 2", len(d.Features))
	}

	missing := proj.Domain("nonexistent")
	if missing != nil {
		t.Error("Domain(nonexistent) should return nil")
	}
}

func TestFeature(t *testing.T) {
	proj := mustOpen(t)
	f := proj.Feature("CLI", "add-task")
	if f == nil {
		t.Fatal("Feature(CLI, add-task) = nil")
	}
	if f.State != "building" {
		t.Errorf("State = %q, want %q", f.State, "building")
	}

	missing := proj.Feature("CLI", "nonexistent")
	if missing != nil {
		t.Error("Feature(CLI, nonexistent) should return nil")
	}

	missingDomain := proj.Feature("nonexistent", "add-task")
	if missingDomain != nil {
		t.Error("Feature(nonexistent, add-task) should return nil")
	}
}

// ---- Mutate methods --------------------------------------------------------

func TestAddIssue(t *testing.T) {
	proj := mustOpen(t)
	newIss := product.Issue{
		ID:     "ISSUE-005",
		Title:  "Test issue",
		Type:   "task",
		Status: "open",
	}
	if err := proj.AddIssue(newIss); err != nil {
		t.Fatalf("AddIssue: %v", err)
	}
	if len(proj.Issues) != 5 {
		t.Errorf("Issues count = %d, want 5", len(proj.Issues))
	}
	found := proj.GetIssue("ISSUE-005")
	if found == nil {
		t.Error("Added issue not found via GetIssue")
	}
}

func TestAddIssue_DuplicateID(t *testing.T) {
	proj := mustOpen(t)
	dup := product.Issue{ID: "ISSUE-001", Title: "Duplicate"}
	err := proj.AddIssue(dup)
	if err == nil {
		t.Error("AddIssue should reject duplicate ID")
	}
}

func TestUpdateIssue(t *testing.T) {
	proj := mustOpen(t)
	err := proj.UpdateIssue("ISSUE-001", func(iss *product.Issue) {
		iss.Status = "in-progress"
	})
	if err != nil {
		t.Fatalf("UpdateIssue: %v", err)
	}
	iss := proj.GetIssue("ISSUE-001")
	if iss.Status != "in-progress" {
		t.Errorf("Status = %q, want %q", iss.Status, "in-progress")
	}
}

func TestUpdateIssue_NotFound(t *testing.T) {
	proj := mustOpen(t)
	err := proj.UpdateIssue("ISSUE-999", func(iss *product.Issue) {})
	if err == nil {
		t.Error("UpdateIssue(ISSUE-999) should return error")
	}
}

func TestCloseIssue(t *testing.T) {
	proj := mustOpen(t)
	if err := proj.CloseIssue("ISSUE-004"); err != nil {
		t.Fatalf("CloseIssue: %v", err)
	}
	iss := proj.GetIssue("ISSUE-004")
	if iss.Status != "closed" {
		t.Errorf("Status = %q, want %q", iss.Status, "closed")
	}
}

func TestCloseIssue_NotFound(t *testing.T) {
	proj := mustOpen(t)
	err := proj.CloseIssue("ISSUE-999")
	if err == nil {
		t.Error("CloseIssue(ISSUE-999) should return error")
	}
}

func TestUpdateFeatureState(t *testing.T) {
	proj := mustOpen(t)
	if err := proj.UpdateFeatureState("CLI", "add-task", "ready"); err != nil {
		t.Fatalf("UpdateFeatureState: %v", err)
	}
	f := proj.Feature("CLI", "add-task")
	if f.State != "ready" {
		t.Errorf("State = %q, want %q", f.State, "ready")
	}
}

func TestUpdateFeatureState_NotFound(t *testing.T) {
	proj := mustOpen(t)
	err := proj.UpdateFeatureState("CLI", "nonexistent", "ready")
	if err == nil {
		t.Error("UpdateFeatureState(CLI, nonexistent) should return error")
	}
}

func TestSetGoalDone(t *testing.T) {
	proj := mustOpen(t)
	if err := proj.SetGoalDone("add-task", true); err != nil {
		t.Fatalf("SetGoalDone: %v", err)
	}
	for _, g := range proj.Product.Goals {
		if g.Slug == "add-task" {
			if !g.Done {
				t.Error("Goal add-task should be done")
			}
			return
		}
	}
	t.Error("goal add-task not found")
}

func TestSetGoalDone_NotFound(t *testing.T) {
	proj := mustOpen(t)
	err := proj.SetGoalDone("nonexistent", true)
	if err == nil {
		t.Error("SetGoalDone(nonexistent) should return error")
	}
}

func TestSetGoalDone_MarkUndone(t *testing.T) {
	proj := mustOpen(t)
	// persistence is already done — mark it undone
	if err := proj.SetGoalDone("persistence", false); err != nil {
		t.Fatalf("SetGoalDone: %v", err)
	}
	for _, g := range proj.Product.Goals {
		if g.Slug == "persistence" {
			if g.Done {
				t.Error("Goal persistence should be marked undone")
			}
			return
		}
	}
	t.Error("goal persistence not found")
}

func TestProjectSave_UpdatesLastUpdated(t *testing.T) {
	proj, err := product.Open(testdataDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	proj = withTempDir(t, proj)
	if err := proj.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	today := time.Now().Format("2006-01-02")
	if proj.Product.LastUpdated != today {
		t.Errorf("LastUpdated = %q, want %q", proj.Product.LastUpdated, today)
	}
}

// ---- helpers ---------------------------------------------------------------

func mustOpen(t *testing.T) *product.Project {
	t.Helper()
	proj, err := product.Open(testdataDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return proj
}

// withTempDir copies the project's data into a fresh temp directory and
// returns a new Project pointing at it. This prevents test side-effects on
// testdata.
func withTempDir(t *testing.T, orig *product.Project) *product.Project {
	t.Helper()
	dir := t.TempDir()
	// Save current state into temp dir first, then Open from there.
	if err := orig.Product.Save(dir); err != nil {
		t.Fatalf("withTempDir Save: %v", err)
	}
	if err := product.SaveIssues(dir, orig.Issues); err != nil {
		t.Fatalf("withTempDir SaveIssues: %v", err)
	}
	// Reset LastUpdated to original so the save-updates-last-updated test is meaningful
	orig.Product.LastUpdated = "2025-01-15"
	proj, err := product.Open(dir)
	if err != nil {
		t.Fatalf("withTempDir Open: %v", err)
	}
	return proj
}
