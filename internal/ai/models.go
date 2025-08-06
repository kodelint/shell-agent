package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kodelint/shell-agent/internal/config"
	"github.com/kodelint/shell-agent/internal/logger"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
)

type ModelInfo struct {
	Name        string
	Description string
	Size        string
	Type        string
	Downloaded  bool
	Path        string
	OllamaName  string // The actual model name in Ollama
	Recommended bool   // Whether this model is recommended for shell commands
}

type ModelManager struct {
	modelsPath   string
	ollamaClient *OllamaClient
	logger       *logrus.Entry
	config       *config.Config
}

func NewModelManager() *ModelManager {
	cfg, _ := config.Load()
	return &ModelManager{
		modelsPath:   config.GetModelPath(),
		ollamaClient: NewOllamaClient(cfg),
		logger:       logger.GetLogger().WithField("component", "model-manager"),
		config:       cfg,
	}
}

func (m *ModelManager) ListAvailableModels() []ModelInfo {
	models := []ModelInfo{
		{
			Name:        "llama3.2:3b",
			OllamaName:  "llama3.2:3b",
			Description: "Llama 3.2 3B - Fast and efficient for command generation",
			Size:        "2.0GB",
			Type:        "Language Model",
			Recommended: true,
			Downloaded:  m.isModelDownloaded("llama3.2:3b"),
		},
		{
			Name:        "llama3.2:1b",
			OllamaName:  "llama3.2:1b",
			Description: "Llama 3.2 1B - Ultra-fast and lightweight",
			Size:        "1.3GB",
			Type:        "Language Model",
			Recommended: true,
			Downloaded:  m.isModelDownloaded("llama3.2:1b"),
		},
		{
			Name:        "codegemma:7b",
			OllamaName:  "codegemma:7b",
			Description: "CodeGemma 7B - Specialized for code and shell commands",
			Size:        "5.0GB",
			Type:        "Code Model",
			Recommended: true,
			Downloaded:  m.isModelDownloaded("codegemma:7b"),
		},
		{
			Name:        "llama3.1:8b",
			OllamaName:  "llama3.1:8b",
			Description: "Llama 3.1 8B - Balanced performance and accuracy",
			Size:        "4.7GB",
			Type:        "Language Model",
			Recommended: false,
			Downloaded:  m.isModelDownloaded("llama3.1:8b"),
		},
		{
			Name:        "mistral:7b",
			OllamaName:  "mistral:7b",
			Description: "Mistral 7B - Good general purpose model",
			Size:        "4.1GB",
			Type:        "Language Model",
			Recommended: false,
			Downloaded:  m.isModelDownloaded("mistral:7b"),
		},
		{
			Name:        "phi3:mini",
			OllamaName:  "phi3:mini",
			Description: "Phi-3 Mini - Microsoft's compact model",
			Size:        "2.3GB",
			Type:        "Small Model",
			Recommended: false,
			Downloaded:  m.isModelDownloaded("phi3:mini"),
		},
	}

	return models
}

func (m *ModelManager) GetCurrentModel() *ModelInfo {
	defaultModel := config.GetDefaultModel()
	models := m.ListAvailableModels()

	// First, try to find exact match
	for _, model := range models {
		if model.Name == defaultModel || model.OllamaName == defaultModel {
			return &model
		}
	}

	// If no exact match, try to find first downloaded model
	for _, model := range models {
		if model.Downloaded {
			return &model
		}
	}

	// Return first recommended model as fallback
	for _, model := range models {
		if model.Recommended {
			return &model
		}
	}

	// Return first model as last resort
	if len(models) > 0 {
		return &models[0]
	}

	return nil
}

func (m *ModelManager) GetModelPath() string {
	return m.modelsPath
}

func (m *ModelManager) DownloadModel(modelName string) error {
	m.logger.WithField("model", modelName).Info("Starting model download via Ollama")

	// Find model info
	models := m.ListAvailableModels()
	var modelInfo *ModelInfo
	for _, model := range models {
		if model.Name == modelName {
			modelInfo = &model
			break
		}
	}

	if modelInfo == nil {
		return fmt.Errorf("unknown model: %s", modelName)
	}

	// Check if Ollama is available
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.ollamaClient.IsAvailable(ctx); err != nil {
		return fmt.Errorf("ollama service is not available: %w\n\nPlease install and start Ollama:\n1. Download from: https://ollama.ai/download\n2. Start service: 'ollama serve'", err)
	}

	// Check if already downloaded in Ollama
	if m.IsModelAvailableInOllama(modelInfo.OllamaName) {
		return fmt.Errorf("model %s is already downloaded in Ollama", modelName)
	}

	// Create progress bar
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(fmt.Sprintf("Downloading %s (%s)", modelName, modelInfo.Size)),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetRenderBlankState(true),
	)

	// Download via Ollama with progress callback
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute) // Long timeout for model downloads
	defer cancel()

	err := m.ollamaClient.PullModel(ctx, modelInfo.OllamaName, func(status string, progress float64) {
		if progress > 0 {
			bar.Set(int(progress * 100))
		}
		bar.Describe(fmt.Sprintf("Downloading %s: %s", modelName, status))
	})

	bar.Finish()
	fmt.Println() // New line after progress bar

	if err != nil {
		return fmt.Errorf("failed to download model %s: %w", modelName, err)
	}

	// Create local metadata (optional, for our tracking)
	if err := m.createModelMetadata(modelInfo); err != nil {
		m.logger.WithError(err).Warn("Failed to create local metadata, but model is available in Ollama")
	}

	m.logger.WithFields(logrus.Fields{
		"model":       modelName,
		"ollama_name": modelInfo.OllamaName,
		"size":        modelInfo.Size,
	}).Info("Model download completed successfully")

	return nil
}

func (m *ModelManager) createModelMetadata(modelInfo *ModelInfo) error {
	// Ensure models directory exists
	if err := os.MkdirAll(m.modelsPath, 0755); err != nil {
		return fmt.Errorf("failed to create models directory %s: %w", m.modelsPath, err)
	}

	modelPath := filepath.Join(m.modelsPath, modelInfo.Name)
	if err := os.MkdirAll(modelPath, 0755); err != nil {
		return fmt.Errorf("failed to create model directory %s: %w", modelPath, err)
	}

	// Create metadata file
	metadataFile := filepath.Join(modelPath, "metadata.json")
	file, err := os.Create(metadataFile)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()

	metadata := fmt.Sprintf(`{
    "name": "%s",
    "ollama_name": "%s",
    "description": "%s",
    "size": "%s",
    "type": "%s",
    "recommended": %t,
    "downloaded_at": "%s",
    "source": "ollama"
}`, modelInfo.Name, modelInfo.OllamaName, modelInfo.Description,
		modelInfo.Size, modelInfo.Type, modelInfo.Recommended, time.Now().Format(time.RFC3339))

	_, err = file.WriteString(metadata)
	return err
}

func (m *ModelManager) isModelDownloaded(modelName string) bool {
	// Check both local metadata and Ollama availability
	return m.hasLocalMetadata(modelName) && m.IsModelAvailableInOllama(modelName)
}

func (m *ModelManager) hasLocalMetadata(modelName string) bool {
	metadataPath := filepath.Join(m.modelsPath, modelName, "metadata.json")
	_, err := os.Stat(metadataPath)
	return err == nil
}

func (m *ModelManager) IsModelAvailableInOllama(modelName string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.ollamaClient.IsAvailable(ctx); err != nil {
		return false
	}

	models, err := m.ollamaClient.ListModels(ctx)
	if err != nil {
		m.logger.WithError(err).Debug("Failed to list Ollama models")
		return false
	}

	for _, model := range models {
		// Check both exact match and name without tag
		if model.Name == modelName {
			return true
		}
		// Also check if model name matches without the tag (e.g., "llama3.2" matches "llama3.2:3b")
		if strings.Contains(model.Name, strings.Split(modelName, ":")[0]) {
			return true
		}
	}

	return false
}

func (m *ModelManager) GetOllamaModels() ([]OllamaModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.ollamaClient.IsAvailable(ctx); err != nil {
		return nil, fmt.Errorf("ollama service is not available: %w", err)
	}

	return m.ollamaClient.ListModels(ctx)
}

func (m *ModelManager) GetRecommendedModel() *ModelInfo {
	models := m.ListAvailableModels()

	// First try to find a downloaded recommended model
	for _, model := range models {
		if model.Recommended && model.Downloaded {
			return &model
		}
	}

	// Then try to find any recommended model
	for _, model := range models {
		if model.Recommended {
			return &model
		}
	}

	// Fallback to first model
	if len(models) > 0 {
		return &models[0]
	}

	return nil
}
