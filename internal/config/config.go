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
		DefaultModel string `mapstructure:"default_model"`
		ModelPath    string `mapstructure:"model_path"`
		Timeout      int    `mapstructure:"timeout"`
	} `mapstructure:"ai"`

	Logging struct {
		Level string `mapstructure:"level"`
		File  string `mapstructure:"file"`
	} `mapstructure:"logging"`

	Interactive struct {
		ConfirmCommands bool `mapstructure:"confirm_commands"`
		ShowExplanation bool `mapstructure:"show_explanation"`
	} `mapstructure:"interactive"`
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
	viper.SetDefault("ai.default_model", "llama2-7b")
	viper.SetDefault("ai.model_path", filepath.Join(home, ".shell-agent", "models"))
	viper.SetDefault("ai.timeout", 30)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "")

	// Interactive defaults
	viper.SetDefault("interactive.confirm_commands", true)
	viper.SetDefault("interactive.show_explanation", true)
}

// expandPath expands ~ to home directory and resolves the path
func expandPath(path string) string {
	log.Debugf("Expanding path: %s", path)
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		log.Debugf("Using home dir: %s", home)
		return filepath.Join(home, path[1:])
	}
	log.Debugf("Using home dir: %s", path)
	return path
}

func GetModelPath() string {
	log.Infof("Getting model path")
	path := viper.GetString("ai.model_path")
	log.Infof("Model path is %s", path)
	log.Infof("Returning Model info: %s", expandPath(path))
	return expandPath(path)
}

func GetDefaultModel() string {
	return viper.GetString("ai.default_model")
}

func GetTimeout() int {
	return viper.GetInt("ai.timeout")
}
