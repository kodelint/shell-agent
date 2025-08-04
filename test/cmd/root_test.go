package cmd

import (
	"github.com/kodelint/shell-agent/cmd"
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Test that root command can be created
	rootCmd := cmd.NewRootCommand()
	if rootCmd == nil {
		t.Error("Root command should not be nil")
	}

	// Test command name
	if rootCmd.Use != "shell-agent" {
		t.Errorf("Expected command name 'shell-agent', got '%s'", rootCmd.Use)
	}
}
