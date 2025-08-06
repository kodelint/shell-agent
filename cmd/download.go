package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/kodelint/shell-agent/internal/ai"
	"github.com/kodelint/shell-agent/internal/logger"
	"github.com/kodelint/shell-agent/internal/output"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download AI models locally",
	Long: `Download and configure AI models for local use.

This command allows you to download various AI models and configure
them for use with shell-agent.

Examples:
  shell-agent download                    # Interactive model selection
  shell-agent download --model llama2     # Download specific model
  shell-agent download --list             # List available models`,
	Run: func(cmd *cobra.Command, args []string) {
		runDownload(cmd, args)
	},
}

var (
	modelName  string
	listModels bool
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVarP(&modelName, "model", "m", "", "Specific model to download")
	downloadCmd.Flags().BoolVarP(&listModels, "list", "l", false, "List available models")
}

func runDownload(cmd *cobra.Command, args []string) {
	log := logger.GetLogger()
	log.Info("Starting model download")

	modelManager := ai.NewModelManager()

	if listModels {
		models := modelManager.ListAvailableModels()
		output.PrintAvailableModels(models)
		return
	}

	if modelName == "" {
		// Interactive model selection
		fmt.Println()
		output.PrintInfo("🔍 Select an AI model to download:")

		selectedModel, err := output.PromptModelSelection(modelManager.ListAvailableModels())
		if err != nil {
			if err.Error() == "^C" {
				output.PrintInfo("❌ Download cancelled")
				return
			}
			output.PrintError(fmt.Sprintf("Model selection failed: %v", err))
			return
		}
		modelName = selectedModel
	}

	// Check if model is already downloaded
	models := modelManager.ListAvailableModels()
	var selectedModelInfo *ai.ModelInfo
	for _, model := range models {
		if model.Name == modelName {
			selectedModelInfo = &model
			break
		}
	}

	if selectedModelInfo == nil {
		output.PrintError(fmt.Sprintf("Unknown model: %s", modelName))
		output.PrintInfo("💡 Run 'shell-agent download --list' to see available models")
		return
	}

	if selectedModelInfo.Downloaded {
		output.PrintWarning(fmt.Sprintf("Model %s is already downloaded", modelName))
		output.PrintInfo(fmt.Sprintf("📁 Location: %s", modelManager.GetModelPath()))
		return
	}

	// Show what we're about to download
	fmt.Println()
	output.PrintInfo(fmt.Sprintf("📦 Downloading: %s", selectedModelInfo.Name))
	output.PrintInfo(fmt.Sprintf("📝 Description: %s", selectedModelInfo.Description))
	output.PrintInfo(fmt.Sprintf("💾 Size: %s", selectedModelInfo.Size))
	output.PrintInfo(fmt.Sprintf("📁 Destination: %s", modelManager.GetModelPath()))
	fmt.Println()

	// Download the model
	err := modelManager.DownloadModel(modelName)
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to download model %s: %v", modelName, err))
		return
	}

	output.PrintSuccess(fmt.Sprintf("Successfully downloaded model: %s", modelName))
	output.PrintInfo(fmt.Sprintf("📁 Installed to: %s", filepath.Join(modelManager.GetModelPath(), modelName)))
	output.PrintInfo("🚀 You can now start using shell-agent!")
}
