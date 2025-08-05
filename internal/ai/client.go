package ai

import (
	"context"
	"fmt"
	"github.com/kodelint/shell-agent/internal/config"
	"github.com/kodelint/shell-agent/internal/logger"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
	"time"
)

type Client struct {
	ollamaClient  *OllamaClient
	modelManager  *ModelManager
	config        *config.Config
	logger        *logrus.Entry
	safetyChecker *SafetyChecker
}

type CommandResponse struct {
	Command      string   `json:"command"`
	Explanation  string   `json:"explanation"`
	Warning      string   `json:"warning"`
	Confidence   float64  `json:"confidence"`
	Alternatives []string `json:"alternatives,omitempty"`
}

// SafetyChecker validates commands for safety
type SafetyChecker struct {
	dangerousPatterns []string
	config            *config.Config
	logger            *logrus.Entry
}

func NewClient() (*Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	client := &Client{
		ollamaClient:  NewOllamaClient(cfg),
		modelManager:  NewModelManager(),
		config:        cfg,
		logger:        logger.GetLogger().WithField("component", "ai-client"),
		safetyChecker: NewSafetyChecker(cfg),
	}

	return client, nil
}

func NewSafetyChecker(cfg *config.Config) *SafetyChecker {
	return &SafetyChecker{
		dangerousPatterns: cfg.Safety.DangerousCommands,
		config:            cfg,
		logger:            logger.GetLogger().WithField("component", "safety-checker"),
	}
}

func (c *Client) GenerateCommand(input string) (*CommandResponse, error) {
	c.logger.WithField("input", input).Info("Generating command")

	// Check if Ollama is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.ollamaClient.IsAvailable(ctx); err != nil {
		return nil, fmt.Errorf("ollama service is not available: %w\n\nPlease ensure Ollama is installed and running:\n- Install: https://ollama.ai/download\n- Start: 'ollama serve'", err)
	}

	// Check if model is available
	currentModel := c.modelManager.GetCurrentModel()
	if currentModel == nil {
		return nil, fmt.Errorf("no AI model configured. Please run 'shell-agent download' first")
	}

	// Verify model exists in Ollama
	if !c.modelManager.IsModelAvailableInOllama(currentModel.Name) {
		return nil, fmt.Errorf("model '%s' is not available in Ollama. Please run 'shell-agent download' to install it", currentModel.Name)
	}

	// Create context with timeout
	ctx, cancel = context.WithTimeout(context.Background(), time.Duration(c.config.AI.Timeout)*time.Second)
	defer cancel()

	// Enhance prompt with system context
	enhancedPrompt := c.enhancePrompt(input)

	// Generate command using Ollama
	response, err := c.ollamaClient.Generate(ctx, currentModel.Name, enhancedPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate command: %w", err)
	}

	// Apply safety checks
	if c.config.Safety.RequireConfirm {
		c.safetyChecker.CheckCommand(response)
	}

	c.logger.WithFields(logrus.Fields{
		"command":    response.Command,
		"confidence": response.Confidence,
		"model":      currentModel.Name,
	}).Info("Generated command successfully")

	return response, nil
}

func (c *Client) enhancePrompt(input string) string {
	// Add system context
	osInfo := runtime.GOOS
	prompt := fmt.Sprintf(`Operating System: %s

User Request: %s

Please provide a shell command that accomplishes this request. Consider:
1. The operating system is %s
2. Use safe, commonly available commands
3. Provide clear explanations
4. Warn about any potential risks
5. Suggest alternatives if helpful

Respond in JSON format as specified in the system prompt.`, osInfo, input, osInfo)

	return prompt
}

func (s *SafetyChecker) CheckCommand(response *CommandResponse) {
	if response.Command == "" {
		return
	}

	command := strings.ToLower(response.Command)

	// Check for dangerous patterns
	for _, pattern := range s.dangerousPatterns {
		if strings.Contains(command, strings.ToLower(pattern)) {
			warning := fmt.Sprintf("⚠️ DANGER: This command contains '%s' which can be destructive", pattern)
			if response.Warning != "" {
				response.Warning = response.Warning + "\n" + warning
			} else {
				response.Warning = warning
			}

			// Lower confidence for dangerous commands
			if response.Confidence > 0.5 {
				response.Confidence = 0.5
			}

			s.logger.WithFields(logrus.Fields{
				"command": response.Command,
				"pattern": pattern,
			}).Warn("Dangerous command pattern detected")
			break
		}
	}

	// Additional safety checks
	if strings.Contains(command, "sudo") && !strings.Contains(response.Warning, "sudo") {
		addWarning := "⚠️ This command requires administrative privileges"
		if response.Warning != "" {
			response.Warning = response.Warning + "\n" + addWarning
		} else {
			response.Warning = addWarning
		}
	}

	// Check for recursive operations
	if strings.Contains(command, "-r") && (strings.Contains(command, "rm") || strings.Contains(command, "chmod") || strings.Contains(command, "chown")) {
		addWarning := "⚠️ This command will operate recursively on directories"
		if response.Warning != "" {
			response.Warning = response.Warning + "\n" + addWarning
		} else {
			response.Warning = addWarning
		}
	}
}
