package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kodelint/shell-agent/internal/ai"
	"github.com/kodelint/shell-agent/internal/config"
	"github.com/kodelint/shell-agent/internal/output"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup shell-agent and install Ollama",
	Long: `Setup command helps you install and configure all dependencies needed for shell-agent.

This includes:
- Installing Ollama (if not present)
- Starting Ollama service
- Downloading a recommended AI model
- Configuring shell-agent

Examples:
  shell-agent setup              # Full automated setup
  shell-agent setup --ollama     # Only install Ollama
  shell-agent setup --model      # Only download model`,
	Run: func(cmd *cobra.Command, args []string) {
		runSetup(cmd, args)
	},
}

var (
	setupOllamaOnly bool
	setupModelOnly  bool
	skipConfirm     bool
)

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().BoolVar(&setupOllamaOnly, "ollama", false, "Only install Ollama")
	setupCmd.Flags().BoolVar(&setupModelOnly, "model", false, "Only download recommended model")
	setupCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompts")
}

func runSetup(cmd *cobra.Command, args []string) {
	output.PrintSetupWelcome()

	if !skipConfirm {
		if !output.PromptSetupConfirm() {
			output.PrintInfo("Setup cancelled")
			return
		}
	}

	if !setupModelOnly {
		// Step 1: Check and install Ollama
		if err := setupOllama(); err != nil {
			output.PrintError(fmt.Sprintf("Failed to setup Ollama: %v", err))
			return
		}
	}

	if !setupOllamaOnly {
		// Step 2: Download recommended model
		if err := setupModel(); err != nil {
			output.PrintError(fmt.Sprintf("Failed to setup model: %v", err))
			return
		}
	}

	// Step 3: Create default config if it doesn't exist
	if err := setupConfig(); err != nil {
		output.PrintWarning(fmt.Sprintf("Failed to create config: %v", err))
	}

	output.PrintSetupComplete()
}

func setupOllama() error {
	output.PrintInfo("🔍 Checking Ollama installation...")

	// Check if Ollama is already installed
	if isOllamaInstalled() {
		output.PrintSuccess("✅ Ollama is already installed")

		// Check if Ollama service is running
		if isOllamaRunning() {
			output.PrintSuccess("✅ Ollama service is running")
			return nil
		} else {
			output.PrintInfo("🚀 Starting Ollama service...")
			return startOllama()
		}
	}

	output.PrintInfo("📦 Ollama not found. Installing Ollama...")
	return installOllama()
}

func setupModel() error {
	output.PrintInfo("🤖 Setting up AI model...")

	// Use the download command to get a recommended model
	modelManager := ai.NewModelManager()
	recommendedModel := modelManager.GetRecommendedModel()

	if recommendedModel == nil {
		return fmt.Errorf("no recommended model found")
	}

	if recommendedModel.Downloaded {
		output.PrintSuccess(fmt.Sprintf("✅ Recommended model '%s' is already available", recommendedModel.Name))
		return nil
	}

	output.PrintInfo(fmt.Sprintf("📥 Downloading recommended model: %s", recommendedModel.Name))
	return modelManager.DownloadModel(recommendedModel.Name)
}

func setupConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, ".shell-agent.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		output.PrintSuccess("✅ Configuration file already exists")
		return nil
	}

	output.PrintInfo("📝 Creating default configuration...")

	// Create default config content
	defaultConfig := `# Shell Agent Configuration
ai:
  provider: "ollama"
  default_model: "llama3.2:3b"
  timeout: 120
  max_tokens: 2048
  temperature: 0.1
  
  ollama:
    host: "localhost"
    port: 11434

interactive:
  confirm_commands: true
  show_explanation: true
  show_confidence: true
  auto_execute: false

safety:
  require_confirm: true
  block_destructive: false

logging:
  level: "info"
`

	err = os.WriteFile(configPath, []byte(defaultConfig), 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	output.PrintSuccess(fmt.Sprintf("✅ Created configuration at %s", configPath))
	return nil
}

func isOllamaInstalled() bool {
	_, err := exec.LookPath("ollama")
	return err == nil
}

func isOllamaRunning() bool {
	// Try to connect to Ollama service
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cfg, _ := config.Load()
	client := ai.NewOllamaClient(cfg)

	return client.IsAvailable(ctx) == nil
}

func startOllama() error {
	output.PrintInfo("🚀 Starting Ollama service...")

	cmd := exec.Command("ollama", "serve")

	// Start Ollama in background
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start Ollama: %w", err)
	}

	// Wait a moment for service to start
	time.Sleep(3 * time.Second)

	// Check if it's running
	if isOllamaRunning() {
		output.PrintSuccess("✅ Ollama service started successfully")
		return nil
	}

	return fmt.Errorf("ollama service failed to start properly")
}

func installOllama() error {
	switch runtime.GOOS {
	case "darwin":
		return installOllamaMacOS()
	case "linux":
		return installOllamaLinux()
	case "windows":
		return installOllamaWindows()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func installOllamaMacOS() error {
	output.PrintInfo("🍎 Installing Ollama on macOS...")

	// Check if Homebrew is available
	if _, err := exec.LookPath("brew"); err == nil {
		output.PrintInfo("🍺 Using Homebrew to install Ollama...")
		cmd := exec.Command("brew", "install", "ollama")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install Ollama via Homebrew: %w", err)
		}

		output.PrintSuccess("✅ Ollama installed via Homebrew")
		return startOllama()
	}

	// Fallback to manual installation
	output.PrintInfo("📥 Downloading Ollama installer...")
	return fmt.Errorf("please install Ollama manually from https://ollama.ai/download")
}

func installOllamaLinux() error {
	output.PrintInfo("🐧 Installing Ollama on Linux...")

	// Use the official install script
	cmd := exec.Command("curl", "-fsSL", "https://ollama.ai/install.sh")
	installScript := exec.Command("sh")

	// Pipe curl output to sh
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}

	installScript.Stdin = pipe
	installScript.Stdout = os.Stdout
	installScript.Stderr = os.Stderr

	if err := installScript.Start(); err != nil {
		return fmt.Errorf("failed to start install script: %w", err)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download install script: %w", err)
	}

	if err := installScript.Wait(); err != nil {
		return fmt.Errorf("failed to install Ollama: %w", err)
	}

	output.PrintSuccess("✅ Ollama installed successfully")
	return startOllama()
}

func installOllamaWindows() error {
	output.PrintInfo("🪟 Windows installation detected...")
	output.PrintWarning("Please install Ollama manually:")
	output.PrintInfo("1. Download from: https://ollama.ai/download")
	output.PrintInfo("2. Run the installer")
	output.PrintInfo("3. Start Ollama from Start Menu")
	output.PrintInfo("4. Run 'shell-agent setup --model' to download a model")

	return fmt.Errorf("manual installation required on Windows")
}
