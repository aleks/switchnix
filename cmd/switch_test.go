package cmd

import "testing"

func TestValidActions(t *testing.T) {
	valid := []string{"switch", "test", "boot"}
	for _, action := range valid {
		if !validActions[action] {
			t.Errorf("expected %q to be a valid action", action)
		}
	}

	invalid := []string{"", "rebuild", "install", "rollback", "; rm -rf /"}
	for _, action := range invalid {
		if validActions[action] {
			t.Errorf("expected %q to be an invalid action", action)
		}
	}
}
