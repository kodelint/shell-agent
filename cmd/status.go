package cmd

import (
	"github.com/kodelint/shell-agent/internal/ai"
	"github.com/kodelint/shell-agent/internal/output"
	"github.com/kodelint/shell-agent/internal/system"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show shell-agent status",
	Long: `Display the current status of shell-agent including:
- Configured AI model
- Model availability
- System information
- Configuration details

Examples:
  shell-agent status          # Show complete status
  shell-agent status --model  # Show only model status`,
	Run: func(cmd *cobra.Command, args []string) {
		runStatus(cmd, args)
	},
}

var (
	modelOnly bool
)

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolVar(&modelOnly, "model", false, "Show only model status")
}

func runStatus(cmd *cobra.Command, args []string) {
	modelManager := ai.NewModelManager()
	systemInfo := system.NewSystemInfo()

	status := &output.StatusInfo{
		ModelManager: modelManager,
		SystemInfo:   systemInfo,
		ModelOnly:    modelOnly,
	}

	output.PrintStatus(status)
}
