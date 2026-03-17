package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aleks/switchnix/internal/ui"
	"github.com/pmezard/go-difflib/difflib"
)

// ChangeSet describes the differences between local and remote file sets.
type ChangeSet struct {
	Added    []string // files in local but not remote
	Removed  []string // files in remote but not local
	Modified []string // files present in both but with different content
}

func (cs *ChangeSet) HasChanges() bool {
	return len(cs.Added) > 0 || len(cs.Removed) > 0 || len(cs.Modified) > 0
}

// ComputeChangeSet compares local and remote file maps and returns a ChangeSet.
func ComputeChangeSet(local, remote map[string]string) ChangeSet {
	var cs ChangeSet

	for name := range local {
		if _, ok := remote[name]; !ok {
			cs.Added = append(cs.Added, name)
		} else if local[name] != remote[name] {
			cs.Modified = append(cs.Modified, name)
		}
	}

	for name := range remote {
		if _, ok := local[name]; !ok {
			cs.Removed = append(cs.Removed, name)
		}
	}

	sort.Strings(cs.Added)
	sort.Strings(cs.Removed)
	sort.Strings(cs.Modified)

	return cs
}

// ComputeFileDiff returns a unified diff string for a single file.
// Returns an empty string if there are no differences.
func ComputeFileDiff(filename, oldContent, newContent string) string {
	if oldContent == newContent {
		return ""
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldContent),
		B:        difflib.SplitLines(newContent),
		FromFile: "a/" + filename,
		ToFile:   "b/" + filename,
		Context:  3,
	}

	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return fmt.Sprintf("error computing diff for %s: %v", filename, err)
	}
	return text
}

// FormatDiff adds color to a unified diff string for terminal display.
func FormatDiff(diffText string) string {
	if diffText == "" {
		return ""
	}

	var out strings.Builder
	for _, line := range strings.Split(diffText, "\n") {
		switch {
		case strings.HasPrefix(line, "---"):
			out.WriteString(ui.Bold.Render(line))
		case strings.HasPrefix(line, "+++"):
			out.WriteString(ui.Bold.Render(line))
		case strings.HasPrefix(line, "@@"):
			out.WriteString(ui.Cyan.Render(line))
		case strings.HasPrefix(line, "-"):
			out.WriteString(ui.Red.Render(line))
		case strings.HasPrefix(line, "+"):
			out.WriteString(ui.Green.Render(line))
		default:
			out.WriteString(line)
		}
		out.WriteString("\n")
	}
	return out.String()
}

// PrintChangeSet prints a formatted summary and diffs for a ChangeSet.
// The target parameter describes where changes will be applied (e.g. "remote", "locally").
func PrintChangeSet(cs ChangeSet, local, remote map[string]string, target string) {
	if !cs.HasChanges() {
		fmt.Println("No changes detected.")
		return
	}

	if len(cs.Added) > 0 {
		fmt.Println(ui.Green.Render(fmt.Sprintf("Files to be added %s:", target)))
		for _, f := range cs.Added {
			fmt.Printf("  + %s\n", f)
		}
		fmt.Println()
	}

	if len(cs.Removed) > 0 {
		fmt.Println(ui.Red.Render(fmt.Sprintf("Files to be removed %s:", target)))
		for _, f := range cs.Removed {
			fmt.Printf("  - %s\n", f)
		}
		fmt.Println()
	}

	if len(cs.Modified) > 0 {
		fmt.Println(ui.Yellow.Render(fmt.Sprintf("Files to be modified %s:", target)))
		for _, f := range cs.Modified {
			fmt.Printf("  ~ %s\n", f)
		}
		fmt.Println()
	}

	// Print diffs for modified files
	for _, f := range cs.Modified {
		d := ComputeFileDiff(f, remote[f], local[f])
		if d != "" {
			fmt.Print(FormatDiff(d))
		}
	}

	// Print diffs for new files
	for _, f := range cs.Added {
		d := ComputeFileDiff(f, "", local[f])
		if d != "" {
			fmt.Print(FormatDiff(d))
		}
	}

	// Print diffs for removed files
	for _, f := range cs.Removed {
		d := ComputeFileDiff(f, remote[f], "")
		if d != "" {
			fmt.Print(FormatDiff(d))
		}
	}
}
