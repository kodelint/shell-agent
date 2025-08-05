package config

import (
	"github.com/kodelint/shell-agent/internal/logger"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

var log = logger.GetLogger()

type Config struct {
	AI struct {
		Provider     string  `mapstructure:"provider"`
		DefaultModel string  `mapstructure:"default_model"`
		ModelPath    string  `mapstructure:"model_path"`
		Timeout      int     `mapstructure:"timeout"`
		MaxTokens    int     `mapstructure:"max_tokens"`
		Temperature  float64 `mapstructure:"temperature"`
		SystemPrompt string  `mapstructure:"system_prompt"`

		// Ollama specific settings
		Ollama struct {
			Host string `mapstructure:"host"`
			Port int    `mapstructure:"port"`
		} `mapstructure:"ollama"`
	} `mapstructure:"ai"`

	Logging struct {
		Level string `mapstructure:"level"`
		File  string `mapstructure:"file"`
	} `mapstructure:"logging"`

	Interactive struct {
		ConfirmCommands bool `mapstructure:"confirm_commands"`
		ShowExplanation bool `mapstructure:"show_explanation"`
		ShowConfidence  bool `mapstructure:"show_confidence"`
		AutoExecute     bool `mapstructure:"auto_execute"`
	} `mapstructure:"interactive"`

	Safety struct {
		DangerousCommands []string `mapstructure:"dangerous_commands"`
		RequireConfirm    bool     `mapstructure:"require_confirm"`
		BlockDestructive  bool     `mapstructure:"block_destructive"`
	} `mapstructure:"safety"`
}

func Load() (*Config, error) {
	var config Config

	// Set defaults
	setDefaults()

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults() {
	home, _ := os.UserHomeDir()

	// AI defaults
	viper.SetDefault("ai.provider", "ollama")
	viper.SetDefault("ai.default_model", "llama3.2:3b")
	viper.SetDefault("ai.model_path", filepath.Join(home, ".shell-agent", "models"))
	viper.SetDefault("ai.timeout", 120)
	viper.SetDefault("ai.max_tokens", 2048)
	viper.SetDefault("ai.temperature", 0.1)
	viper.SetDefault("ai.system_prompt", getDefaultSystemPrompt())

	// Ollama defaults
	viper.SetDefault("ai.ollama.host", "localhost")
	viper.SetDefault("ai.ollama.port", 11434)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "")

	// Interactive defaults
	viper.SetDefault("interactive.confirm_commands", true)
	viper.SetDefault("interactive.show_explanation", true)
	viper.SetDefault("interactive.show_confidence", true)
	viper.SetDefault("interactive.auto_execute", false)

	// Safety defaults
	viper.SetDefault("safety.dangerous_commands", []string{
		"rm -rf", "dd if=", "mkfs", "fdisk", "shutdown", "reboot",
		"halt", "init 0", "init 6", "killall", "pkill -9",
	})
	viper.SetDefault("safety.require_confirm", true)
	viper.SetDefault("safety.block_destructive", false)
}

func getDefaultSystemPrompt() string {
	return `You are a helpful shell command assistant. Your job is to convert natural language requests into safe, accurate shell commands.

IMPORTANT RULES:
1. Only respond with valid shell commands for the current operating system
2. Include brief explanations when helpful
3. Warn about potentially dangerous operations
4. Prefer safer alternatives when possible
5. If unsure, ask for clarification rather than guessing
6. Focus on commonly used, portable commands
7. Avoid overly complex one-liners unless specifically requested

Response format should be JSON with these fields:
{
  "command": "the actual shell command",
  "explanation": "brief explanation of what the command does",
  "warning": "any safety warnings or considerations (optional)",
  "confidence": 0.95,
  "alternatives": ["alternative commands if applicable"]
}

Examples:
- "list files" → {"command": "ls -la", "explanation": "Lists all files with detailed information", "confidence": 0.95}
- "delete everything" → {"command": "", "warning": "This request is too dangerous. Please specify exactly what you want to delete.", "confidence": 0.0}
`
}

// expandPath expands ~ to home directory and resolves the path
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

func GetModelPath() string {
	path := viper.GetString("ai.model_path")
	return expandPath(path)
}

func GetDefaultModel() string {
	return viper.GetString("ai.default_model")
}

func GetTimeout() int {
	return viper.GetInt("ai.timeout")
}

func GetProvider() string {
	return viper.GetString("ai.provider")
}

func GetOllamaHost() string {
	return viper.GetString("ai.ollama.host")
}

func GetOllamaPort() int {
	return viper.GetInt("ai.ollama.port")
}

func GetMaxTokens() int {
	return viper.GetInt("ai.max_tokens")
}

func GetTemperature() float64 {
	return viper.GetFloat64("ai.temperature")
}

func GetSystemPrompt() string {
	return viper.GetString("ai.system_prompt")
}
