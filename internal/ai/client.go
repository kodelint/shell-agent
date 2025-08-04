package ai

import (
	"context"
	"fmt"
	"github.com/kodelint/shell-agent/internal/config"
	"github.com/kodelint/shell-agent/internal/logger"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Client struct {
	modelManager *ModelManager
	config       *config.Config
	logger       *logrus.Entry
}

type CommandResponse struct {
	Command     string
	Explanation string
	Warning     string
	Confidence  float64
}

func NewClient() (*Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &Client{
		modelManager: NewModelManager(),
		config:       cfg,
		logger:       logger.GetLogger().WithField("component", "ai-client"),
	}, nil
}

func (c *Client) GenerateCommand(input string) (*CommandResponse, error) {
	c.logger.WithField("input", input).Debug("Generating command")

	// Check if model is available
	currentModel := c.modelManager.GetCurrentModel()
	if currentModel == nil || !currentModel.Downloaded {
		return nil, fmt.Errorf("no AI model available. Please run 'shell-agent download' first")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.config.AI.Timeout)*time.Second)
	defer cancel()

	// This is where you would integrate with your local AI model
	// The context can be used for cancellation in actual AI model calls
	// For now, we'll return a mock response based on the input
	response := c.generateMockResponse(ctx, input)

	c.logger.WithField("command", response.Command).Debug("Generated command")

	return response, nil
}

func (c *Client) generateMockResponse(ctx context.Context, input string) *CommandResponse {
	// Mock AI response generation - replace with actual AI model integration
	// The context parameter would be used for actual AI model calls for cancellation
	command, explanation, warning := c.parseInputToCommand(input)

	return &CommandResponse{
		Command:     command,
		Explanation: explanation,
		Warning:     warning,
		Confidence:  0.85,
	}
}

func (c *Client) parseInputToCommand(input string) (string, string, string) {
	// Simple pattern matching for demo - replace with AI model
	patterns := map[string]struct {
		command     string
		explanation string
		warning     string
	}{
		"list files": {
			command:     "ls -la",
			explanation: "Lists all files and directories with detailed information including permissions, size, and modification time.",
			warning:     "",
		},
		"list all files": {
			command:     "ls -la",
			explanation: "Lists all files and directories with detailed information including permissions, size, and modification time.",
			warning:     "",
		},
		"find python files": {
			command:     "find . -name '*.py' -type f",
			explanation: "Recursively searches for all Python files (.py extension) in the current directory and subdirectories.",
			warning:     "",
		},
		"disk usage": {
			command:     "du -sh .",
			explanation: "Shows the disk usage of the current directory in human-readable format.",
			warning:     "",
		},
		"compress folder": {
			command:     "tar -czf archive.tar.gz .",
			explanation: "Creates a compressed tar.gz archive of the current directory.",
			warning:     "This will create an archive of the entire current directory. Make sure you're in the right location.",
		},
	}

	// Find the best match
	for pattern, response := range patterns {
		if contains(input, pattern) {
			return response.command, response.explanation, response.warning
		}
	}

	// Default response
	return "echo 'Command not recognized. Please try a different request.'",
		"I couldn't understand your request. Please try rephrasing it or use more specific terms.",
		"This is a fallback response. Consider using more specific language."
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			strings.Contains(strings.ToLower(s), strings.ToLower(substr)))))
}
