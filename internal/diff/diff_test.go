package diff

import (
	"strings"
	"testing"

	"github.com/aleks/switchnix/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

func init() {
	// Disable styles in tests for predictable output
	ui.Green = lipgloss.NewStyle()
	ui.Red = lipgloss.NewStyle()
	ui.Yellow = lipgloss.NewStyle()
	ui.Cyan = lipgloss.NewStyle()
	ui.Bold = lipgloss.NewStyle()
	ui.Faint = lipgloss.NewStyle()
	ui.Success = lipgloss.NewStyle()
	ui.Info = lipgloss.NewStyle()
	ui.Warn = lipgloss.NewStyle()
}

func TestComputeFileDiff_IdenticalContent(t *testing.T) {
	content := "line1\nline2\nline3\n"
	result := ComputeFileDiff("test.nix", content, content)
	if result != "" {
		t.Errorf("expected empty diff for identical content, got %q", result)
	}
}

func TestComputeFileDiff_ModifiedContent(t *testing.T) {
	old := "line1\nline2\nline3\n"
	new := "line1\nmodified\nline3\n"
	result := ComputeFileDiff("test.nix", old, new)
	if result == "" {
		t.Error("expected non-empty diff for modified content")
	}
	if !strings.Contains(result, "--- a/test.nix") {
		t.Error("expected diff to contain old file header")
	}
	if !strings.Contains(result, "+++ b/test.nix") {
		t.Error("expected diff to contain new file header")
	}
	if !strings.Contains(result, "-line2") {
		t.Error("expected diff to contain removed line")
	}
	if !strings.Contains(result, "+modified") {
		t.Error("expected diff to contain added line")
	}
}

func TestComputeFileDiff_NewFile(t *testing.T) {
	result := ComputeFileDiff("new.nix", "", "new content\n")
	if result == "" {
		t.Error("expected non-empty diff for new file")
	}
	if !strings.Contains(result, "+new content") {
		t.Error("expected diff to show added content")
	}
}

func TestComputeFileDiff_DeletedFile(t *testing.T) {
	result := ComputeFileDiff("old.nix", "old content\n", "")
	if result == "" {
		t.Error("expected non-empty diff for deleted file")
	}
	if !strings.Contains(result, "-old content") {
		t.Error("expected diff to show removed content")
	}
}

func TestComputeFileDiff_NoTrailingNewline(t *testing.T) {
	old := "line1\nline2"
	new := "line1\nline2\nline3"
	result := ComputeFileDiff("test.nix", old, new)
	if result == "" {
		t.Error("expected non-empty diff")
	}
	if !strings.Contains(result, "+line3") {
		t.Error("expected diff to show added line")
	}
}

func TestFormatDiff_EmptyInput(t *testing.T) {
	result := FormatDiff("")
	if result != "" {
		t.Errorf("expected empty string for empty input, got %q", result)
	}
}

func TestFormatDiff_PreservesAllLines(t *testing.T) {
	input := "--- a/test.nix\n+++ b/test.nix\n@@ -1,3 +1,3 @@\n line1\n-line2\n+modified\n line3\n"
	result := FormatDiff(input)

	// With styles disabled, output should contain all the original content
	if !strings.Contains(result, "--- a/test.nix") {
		t.Error("expected formatted diff to contain old file header")
	}
	if !strings.Contains(result, "+++ b/test.nix") {
		t.Error("expected formatted diff to contain new file header")
	}
	if !strings.Contains(result, "@@ -1,3 +1,3 @@") {
		t.Error("expected formatted diff to contain hunk header")
	}
	if !strings.Contains(result, "-line2") {
		t.Error("expected formatted diff to contain removed line")
	}
	if !strings.Contains(result, "+modified") {
		t.Error("expected formatted diff to contain added line")
	}
	if !strings.Contains(result, " line1") {
		t.Error("expected formatted diff to contain context line")
	}
}

func TestComputeChangeSet_NoChanges(t *testing.T) {
	local := map[string]string{
		"configuration.nix": "content1",
	}
	remote := map[string]string{
		"configuration.nix": "content1",
	}

	cs := ComputeChangeSet(local, remote)
	if len(cs.Added) != 0 || len(cs.Removed) != 0 || len(cs.Modified) != 0 {
		t.Error("expected no changes")
	}
	if cs.HasChanges() {
		t.Error("expected HasChanges to be false")
	}
}

func TestComputeChangeSet_AddedFile(t *testing.T) {
	local := map[string]string{
		"configuration.nix": "content1",
		"hardware.nix":      "content2",
	}
	remote := map[string]string{
		"configuration.nix": "content1",
	}

	cs := ComputeChangeSet(local, remote)
	if len(cs.Added) != 1 || cs.Added[0] != "hardware.nix" {
		t.Errorf("expected 1 added file 'hardware.nix', got %v", cs.Added)
	}
	if !cs.HasChanges() {
		t.Error("expected HasChanges to be true")
	}
}

func TestComputeChangeSet_RemovedFile(t *testing.T) {
	local := map[string]string{
		"configuration.nix": "content1",
	}
	remote := map[string]string{
		"configuration.nix": "content1",
		"hardware.nix":      "content2",
	}

	cs := ComputeChangeSet(local, remote)
	if len(cs.Removed) != 1 || cs.Removed[0] != "hardware.nix" {
		t.Errorf("expected 1 removed file 'hardware.nix', got %v", cs.Removed)
	}
}

func TestComputeChangeSet_ModifiedFile(t *testing.T) {
	local := map[string]string{
		"configuration.nix": "new content",
	}
	remote := map[string]string{
		"configuration.nix": "old content",
	}

	cs := ComputeChangeSet(local, remote)
	if len(cs.Modified) != 1 || cs.Modified[0] != "configuration.nix" {
		t.Errorf("expected 1 modified file 'configuration.nix', got %v", cs.Modified)
	}
}

func TestComputeChangeSet_MixedChanges(t *testing.T) {
	local := map[string]string{
		"configuration.nix": "modified",
		"flake.nix":         "new file",
	}
	remote := map[string]string{
		"configuration.nix": "original",
		"hardware.nix":      "to be removed",
	}

	cs := ComputeChangeSet(local, remote)
	if len(cs.Added) != 1 {
		t.Errorf("expected 1 added, got %d", len(cs.Added))
	}
	if len(cs.Removed) != 1 {
		t.Errorf("expected 1 removed, got %d", len(cs.Removed))
	}
	if len(cs.Modified) != 1 {
		t.Errorf("expected 1 modified, got %d", len(cs.Modified))
	}
}

func TestComputeChangeSet_EmptyMaps(t *testing.T) {
	cs := ComputeChangeSet(map[string]string{}, map[string]string{})
	if cs.HasChanges() {
		t.Error("expected no changes for empty maps")
	}
}

func TestComputeChangeSet_SortedOutput(t *testing.T) {
	local := map[string]string{
		"c.nix": "c",
		"a.nix": "a",
		"b.nix": "b",
	}
	remote := map[string]string{}

	cs := ComputeChangeSet(local, remote)
	if len(cs.Added) != 3 {
		t.Fatalf("expected 3 added, got %d", len(cs.Added))
	}
	if cs.Added[0] != "a.nix" || cs.Added[1] != "b.nix" || cs.Added[2] != "c.nix" {
		t.Errorf("expected sorted output, got %v", cs.Added)
	}
}
