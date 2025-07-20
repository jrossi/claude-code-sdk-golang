package claudecode

import (
	"testing"
)

func TestPermissionModeConstants(t *testing.T) {
	tests := []struct {
		mode     PermissionMode
		expected string
	}{
		{PermissionModeDefault, "default"},
		{PermissionModeAcceptEdits, "acceptEdits"},
		{PermissionModeBypassPermissions, "bypassPermissions"},
	}

	for _, test := range tests {
		if string(test.mode) != test.expected {
			t.Errorf("Expected permission mode %v to equal '%s', got '%s'", test.mode, test.expected, string(test.mode))
		}
	}
}
