package ai

import (
	"fmt"
	"github.com/kodelint/shell-agent/internal/config"
	"github.com/kodelint/shell-agent/internal/logger"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"time"
)

type ModelInfo struct {
	Name        string
	Description string
	Size        string
	Type        string
	Downloaded  bool
	Path        string
}

type ModelManager struct {
	modelsPath string
	logger     *logrus.Entry
}

func NewModelManager() *ModelManager {
	return &ModelManager{
		modelsPath: config.GetModelPath(),
		logger:     logger.GetLogger().WithField("component", "model-manager"),
	}
}

func (m *ModelManager) ListAvailableModels() []ModelInfo {
	models := []ModelInfo{
		{
			Name:        "llama2-7b",
			Description: "LLaMA 2 7B - Fast and efficient for command generation",
			Size:        "3.8GB",
			Type:        "Language Model",
			Downloaded:  m.isModelDownloaded("llama2-7b"),
		},
		{
			Name:        "codellama-7b",
			Description: "Code Llama 7B - Specialized for code and shell commands",
			Size:        "3.8GB",
			Type:        "Code Model",
			Downloaded:  m.isModelDownloaded("codellama-7b"),
		},
		{
			Name:        "mistral-7b",
			Description: "Mistral 7B - Balanced performance and speed",
			Size:        "4.1GB",
			Type:        "Language Model",
			Downloaded:  m.isModelDownloaded("mistral-7b"),
		},
		{
			Name:        "phi3-mini",
			Description: "Phi-3 Mini - Lightweight and fast",
			Size:        "2.3GB",
			Type:        "Small Model",
			Downloaded:  m.isModelDownloaded("phi3-mini"),
		},
	}

	return models
}

func (m *ModelManager) GetCurrentModel() *ModelInfo {
	defaultModel := config.GetDefaultModel()
	models := m.ListAvailableModels()

	for _, model := range models {
		if model.Name == defaultModel {
			return &model
		}
	}

	return nil
}

func (m *ModelManager) GetModelPath() string {
	return m.modelsPath
}

func (m *ModelManager) DownloadModel(modelName string) error {
	m.logger.WithField("model", modelName).Info("Starting model download")

	// Ensure models directory exists
	if err := os.MkdirAll(m.modelsPath, 0755); err != nil {
		return fmt.Errorf("failed to create models directory %s: %w", m.modelsPath, err)
	}

	modelPath := filepath.Join(m.modelsPath, modelName)

	// Check if already downloaded
	if m.isModelDownloaded(modelName) {
		return fmt.Errorf("model %s is already downloaded", modelName)
	}

	// Get model info for progress bar
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

	// Simulate download with progress bar
	bar := progressbar.NewOptions64(100,
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
		progressbar.OptionShowCount(),
		progressbar.OptionSetRenderBlankState(true),
	)

	// Simulate download progress
	for i := 0; i < 100; i++ {
		bar.Add(1)
		time.Sleep(30 * time.Millisecond) // Simulate download time
	}

	fmt.Println() // New line after progress bar

	// Create model directory and file
	if err := os.MkdirAll(modelPath, 0755); err != nil {
		return fmt.Errorf("failed to create model directory %s: %w", modelPath, err)
	}

	// Create model file (placeholder)
	modelFile := filepath.Join(modelPath, "model.bin")
	file, err := os.Create(modelFile)
	if err != nil {
		return fmt.Errorf("failed to create model file %s: %w", modelFile, err)
	}
	defer file.Close()

	// Write model metadata
	metadata := fmt.Sprintf(`Model: %s
Description: %s
Size: %s
Type: %s
Downloaded: %s
Path: %s
`, modelInfo.Name, modelInfo.Description, modelInfo.Size, modelInfo.Type,
		time.Now().Format(time.RFC3339), modelFile)

	_, err = file.WriteString(metadata)
	if err != nil {
		return fmt.Errorf("failed to write model file: %w", err)
	}

	// Create config file for the model
	configFile := filepath.Join(modelPath, "config.json")
	config, err := os.Create(configFile)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer config.Close()

	configData := fmt.Sprintf(`{
    "name": "%s",
    "description": "%s",
    "size": "%s",
    "type": "%s",
    "downloaded_at": "%s",
    "version": "1.0.0"
}`, modelInfo.Name, modelInfo.Description, modelInfo.Size, modelInfo.Type, time.Now().Format(time.RFC3339))

	_, err = config.WriteString(configData)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"model": modelName,
		"path":  modelPath,
		"size":  modelInfo.Size,
	}).Info("Model download completed successfully")

	return nil
}

func (m *ModelManager) isModelDownloaded(modelName string) bool {
	modelPath := filepath.Join(m.modelsPath, modelName, "model.bin")
	_, err := os.Stat(modelPath)
	return err == nil
}
