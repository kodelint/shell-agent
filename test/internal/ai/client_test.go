package ai

import (
	"github.com/kodelint/shell-agent/internal/ai"
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := ai.NewClient()
	if err != nil {
		t.Fatalf("Failed to create AI client: %v", err)
	}

	if client == nil {
		t.Error("Client should not be nil")
	}
}

func TestGenerateCommand(t *testing.T) {
	client, err := ai.NewClient()
	if err != nil {
		t.Fatalf("Failed to create AI client: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"list files", "ls -la"},
		{"find python files", "find . -name '*.py' -type f"},
		{"disk usage", "du -sh ."},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			response, err := client.GenerateCommand(test.input)
			if err != nil {
				// Skip test if no model is available
				t.Skipf("Skipping test - no model available: %v", err)
				return
			}

			if response.Command != test.expected {
				t.Errorf("Expected command '%s', got '%s'", test.expected, response.Command)
			}
		})
	}
}
