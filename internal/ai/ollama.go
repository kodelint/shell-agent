package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kodelint/shell-agent/internal/config"
	"github.com/kodelint/shell-agent/internal/logger"
	"github.com/sirupsen/logrus"
)

// OllamaClient handles communication with Ollama API
type OllamaClient struct {
	baseURL string
	client  *http.Client
	config  *config.Config
	logger  *logrus.Entry
}

// OllamaRequest represents the request payload for Ollama API
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	System  string                 `json:"system,omitempty"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
	Format  string                 `json:"format,omitempty"`
}

// OllamaResponse represents the response from Ollama API
type OllamaResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
	Error              string    `json:"error,omitempty"`
}

// OllamaListResponse represents the response from Ollama list models API
type OllamaListResponse struct {
	Models []OllamaModel `json:"models"`
}

// OllamaModel represents a model in Ollama
type OllamaModel struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    struct {
		Format            string   `json:"format"`
		Family            string   `json:"family"`
		Families          []string `json:"families"`
		ParameterSize     string   `json:"parameter_size"`
		QuantizationLevel string   `json:"quantization_level"`
	} `json:"details"`
}

// OllamaPullRequest represents a pull request for downloading models
type OllamaPullRequest struct {
	Name   string `json:"name"`
	Stream bool   `json:"stream"`
}

// OllamaPullResponse represents the streaming response during model download
type OllamaPullResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
	Error     string `json:"error,omitempty"`
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(cfg *config.Config) *OllamaClient {
	baseURL := fmt.Sprintf("http://%s:%d", cfg.AI.Ollama.Host, cfg.AI.Ollama.Port)

	return &OllamaClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: time.Duration(cfg.AI.Timeout) * time.Second,
		},
		config: cfg,
		logger: logger.GetLogger().WithField("component", "ollama-client"),
	}
}

// IsAvailable checks if Ollama service is running
func (c *OllamaClient) IsAvailable(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama service is not available at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama service returned status %d", resp.StatusCode)
	}

	return nil
}

// ListModels retrieves all available models from Ollama
func (c *OllamaClient) ListModels(ctx context.Context) ([]OllamaModel, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API returned status %d", resp.StatusCode)
	}

	var listResp OllamaListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return listResp.Models, nil
}

// PullModel downloads a model from Ollama
func (c *OllamaClient) PullModel(ctx context.Context, modelName string, progressCallback func(status string, progress float64)) error {
	pullReq := OllamaPullRequest{
		Name:   modelName,
		Stream: true,
	}

	reqBody, err := json.Marshal(pullReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/pull", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to pull model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama API returned status %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var pullResp OllamaPullResponse
		if err := decoder.Decode(&pullResp); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode response: %w", err)
		}

		if pullResp.Error != "" {
			return fmt.Errorf("ollama error: %s", pullResp.Error)
		}

		// Calculate progress
		var progress float64
		if pullResp.Total > 0 {
			progress = float64(pullResp.Completed) / float64(pullResp.Total)
		}

		if progressCallback != nil {
			progressCallback(pullResp.Status, progress)
		}

		c.logger.WithFields(logrus.Fields{
			"model":     modelName,
			"status":    pullResp.Status,
			"progress":  progress,
			"completed": pullResp.Completed,
			"total":     pullResp.Total,
		}).Debug("Model download progress")
	}

	return nil
}

// Generate sends a prompt to Ollama and returns the response
func (c *OllamaClient) Generate(ctx context.Context, modelName, prompt string) (*CommandResponse, error) {
	// Prepare the request
	ollamaReq := OllamaRequest{
		Model:  modelName,
		Prompt: prompt,
		System: c.config.AI.SystemPrompt,
		Stream: false,
		Format: "json",
		Options: map[string]interface{}{
			"temperature": c.config.AI.Temperature,
			"num_predict": c.config.AI.MaxTokens,
		},
	}

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.logger.WithFields(logrus.Fields{
		"model":  modelName,
		"prompt": prompt[:min(len(prompt), 100)],
	}).Info("Sending request to Ollama")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if ollamaResp.Error != "" {
		return nil, fmt.Errorf("ollama error: %s", ollamaResp.Error)
	}

	c.logger.WithFields(logrus.Fields{
		"model":             modelName,
		"response_length":   len(ollamaResp.Response),
		"total_duration":    ollamaResp.TotalDuration,
		"prompt_eval_count": ollamaResp.PromptEvalCount,
		"eval_count":        ollamaResp.EvalCount,
	}).Info("Received response from Ollama")

	// Parse the JSON response
	return c.parseOllamaResponse(ollamaResp.Response)
}

// parseOllamaResponse parses the JSON response from Ollama into CommandResponse
func (c *OllamaClient) parseOllamaResponse(response string) (*CommandResponse, error) {
	// Clean the response - sometimes Ollama adds extra text
	response = strings.TrimSpace(response)

	// Try to find JSON in the response
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")

	if startIdx == -1 || endIdx == -1 {
		// If no JSON found, try to extract command from text
		return c.fallbackParseResponse(response), nil
	}

	jsonStr := response[startIdx : endIdx+1]

	// Parse JSON response
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		c.logger.WithError(err).Warn("Failed to parse JSON response, using fallback")
		return c.fallbackParseResponse(response), nil
	}

	cmdResp := &CommandResponse{}

	// Extract fields with type assertions
	if cmd, ok := result["command"].(string); ok {
		cmdResp.Command = strings.TrimSpace(cmd)
	}

	if explanation, ok := result["explanation"].(string); ok {
		cmdResp.Explanation = explanation
	}

	if warning, ok := result["warning"].(string); ok {
		cmdResp.Warning = warning
	}

	if confidence, ok := result["confidence"].(float64); ok {
		cmdResp.Confidence = confidence
	} else {
		cmdResp.Confidence = 0.8 // Default confidence
	}

	if alternatives, ok := result["alternatives"].([]interface{}); ok {
		for _, alt := range alternatives {
			if altStr, ok := alt.(string); ok {
				cmdResp.Alternatives = append(cmdResp.Alternatives, altStr)
			}
		}
	}

	// Validate the response
	if cmdResp.Command == "" && cmdResp.Warning == "" {
		return c.fallbackParseResponse(response), nil
	}

	return cmdResp, nil
}

// fallbackParseResponse provides a fallback when JSON parsing fails
func (c *OllamaClient) fallbackParseResponse(response string) *CommandResponse {
	// Simple heuristic to extract command from text
	lines := strings.Split(response, "\n")
	var command string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Look for command-like patterns
		if strings.Contains(line, "$") {
			// Remove shell prompt
			if idx := strings.Index(line, "$"); idx != -1 && idx < len(line)-1 {
				command = strings.TrimSpace(line[idx+1:])
				break
			}
		} else if strings.Contains(line, "`") {
			// Extract from code blocks
			start := strings.Index(line, "`")
			end := strings.LastIndex(line, "`")
			if start != -1 && end != -1 && start != end {
				command = line[start+1 : end]
				break
			}
		} else if len(line) > 0 && !strings.Contains(line, " ") {
			// Single word commands
			command = line
			break
		}
	}

	if command == "" {
		command = "echo 'Could not parse command from AI response'"
	}

	return &CommandResponse{
		Command:     command,
		Explanation: "AI response was not in expected format, extracted command using fallback method",
		Warning:     "Please verify this command before executing",
		Confidence:  0.3,
	}
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
