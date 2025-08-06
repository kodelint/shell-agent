package cmd

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kodelint/shell-agent/internal/feedback"
	"github.com/kodelint/shell-agent/internal/logger"
	"github.com/kodelint/shell-agent/internal/output"
	"github.com/spf13/cobra"
)

// feedbackCmd represents the feedback command
var feedbackCmd = &cobra.Command{
	Use:   "feedback",
	Short: "Submit feedback on generated commands",
	Long: `Submit feedback to help improve the shell-agent.

This command allows you to tell the system whether a generated command was
useful, incorrect, or failed. This data is stored locally and can be
used to improve the AI model.

Examples:
  shell-agent feedback --prompt "list all files" --command "ls -la" --status worked
  shell-agent feedback --prompt "delete all files" --command "rm -rf *" --status failed --reason "This is dangerous"
  shell-agent feedback --status incorrect --correct-command "git status"
`,
	Run: func(cmd *cobra.Command, args []string) {
		runFeedback(cmd)
	},
}

var (
	userPrompt       string
	generatedCommand string
	feedbackStatus   string
	correctCommand   string
	reason           string
)

func init() {
	// Add the feedback command to the root command
	rootCmd.AddCommand(feedbackCmd)

	// Define flags for the feedback command
	feedbackCmd.Flags().StringVarP(&userPrompt, "prompt", "p", "", "The original user prompt")
	feedbackCmd.Flags().StringVarP(&generatedCommand, "command", "c", "", "The generated shell command")
	feedbackCmd.Flags().StringVarP(&feedbackStatus, "status", "s", "", "The feedback status: 'worked', 'failed', 'incorrect'")
	feedbackCmd.Flags().StringVarP(&correctCommand, "correct-command", "r", "", "The correct command, if the generated one was incorrect")
	feedbackCmd.Flags().StringVarP(&reason, "reason", "e", "", "A short explanation or reason for the feedback")

	// Mark required flags
	feedbackCmd.MarkFlagRequired("status")
}

func runFeedback(cmd *cobra.Command) {
	log := logger.GetLogger()
	log.Debug("Running feedback command")

	// Input validation
	if feedbackStatus == "" {
		output.PrintError("The '--status' flag is required.")
		cmd.Help()
		return
	}
	// Check if status is a valid option
	if feedbackStatus != "worked" && feedbackStatus != "failed" && feedbackStatus != "incorrect" {
		output.PrintError("Invalid status. Please use 'worked', 'failed', or 'incorrect'.")
		cmd.Help()
		return
	}

	// Create a new feedback entry
	entry := feedback.Feedback{
		ID:               uuid.New().String(),
		Timestamp:        time.Now(),
		UserPrompt:       userPrompt,
		GeneratedCommand: generatedCommand,
		Status:           feedbackStatus,
		CorrectCommand:   correctCommand,
		Reason:           reason,
	}

	// Initialize the feedback manager
	feedbackManager, err := feedback.NewManager()
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to initialize feedback manager: %v", err))
		return
	}

	// Save the feedback
	if err := feedbackManager.SaveFeedback(entry); err != nil {
		output.PrintError(fmt.Sprintf("Failed to save feedback: %v", err))
		return
	}

	output.PrintSuccess("✅ Feedback submitted successfully!")
	log.WithField("feedback", entry).Info("Feedback saved")
}
