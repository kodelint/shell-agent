package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/kodelint/shell-agent/internal/ai"
	"github.com/kodelint/shell-agent/internal/feedback"
	"github.com/kodelint/shell-agent/internal/logger"
	"github.com/kodelint/shell-agent/internal/output"
	"github.com/manifoldco/promptui"
)

func runInteractiveMode() {
	log := logger.GetLogger()
	log.Info("Starting interactive REPL mode")

	output.PrintWelcome()

	// Initialize AI client
	aiClient, err := ai.NewClient()
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to initialize AI client: %v", err))
		output.PrintInfo("💡 Try running 'shell-agent download' to install an AI model first")
		os.Exit(1)
	}

	// Handle Ctrl+C gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		output.PrintInfo("\n🛑 Received interrupt signal")
		output.PrintGoodbye()
		os.Exit(0)
	}()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		output.PrintPrompt()

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		// Handle special commands
		switch strings.ToLower(input) {
		case "exit", "quit", "q":
			output.PrintGoodbye()
			return
		case "help", "h":
			output.PrintHelp()
			continue
		case "clear", "cls":
			output.ClearScreen()
			continue
		case "status":
			// Quick status check without exiting
			runStatusInline()
			continue
		}

		// Store the user's original prompt for feedback
		userPrompt := input

		// Process the command with AI
		output.PrintThinking()
		response, err := aiClient.GenerateCommand(input)
		if err != nil {
			output.PrintError(fmt.Sprintf("Error generating command: %v", err))
			if strings.Contains(err.Error(), "no AI model available") {
				output.PrintInfo("💡 Run 'shell-agent download' to install an AI model")
			}
			continue
		}

		output.PrintResponse(response)

		// Ask if user wants to execute the command
		if output.PromptExecuteCommand() {
			output.PrintInfo("🚀 Executing command...")
			err := output.ExecuteCommand(response.Command)
			if err != nil {
				output.PrintError(fmt.Sprintf("Command execution failed: %v", err))
			}
			// FEEDBACK LOGIC
			// After execution, prompt the user for feedback
			feedbackStatus, err := output.PromptForFeedback()

			if err != nil {
				log.WithError(err).Warn("Failed to get user feedback")
				continue
			}
			// If the user provided feedback, save it
			if feedbackStatus != "" {
				// Initialize the feedback manager
				feedbackManager, err := feedback.NewManager()
				if err != nil {
					output.PrintError(fmt.Sprintf("Failed to initialize feedback manager: %v", err))
					continue
				}

				// Create a new feedback entry
				entry := feedback.Feedback{
					ID:               uuid.New().String(),
					Timestamp:        time.Now(),
					UserPrompt:       userPrompt,
					GeneratedCommand: response.Command,
					Status:           feedbackStatus,
				}

				// Save the feedback to the local file
				if err := feedbackManager.SaveFeedback(entry); err != nil {
					output.PrintError(fmt.Sprintf("Failed to save feedback: %v", err))
					continue
				}
				output.PrintSuccess("✅ Feedback submitted. Thank you!")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.WithError(err).Error("Error reading input")
	}
}

func PromptForFeedback() (string, error) {
	// Options for the user to choose from
	options := []string{"👍 Worked", "👎 Didn't Work", "❌ Incorrect"}

	// Use `promptui` to create a menu for a rich UI
	prompt := promptui.Select{
		Label: "Did this command work for you?",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . | bold}}",
			Active:   "{{ . | bold | green | underline}}",
			Inactive: "{{ . | white}}",
			Selected: "{{ . | bold | green | underline}}",
		},
		Size: 3,
	}

	// Run the prompt
	_, result, err := prompt.Run()
	if err != nil {
		if err.Error() == "^C" { // Handle Ctrl+C
			return "", nil
		}
		return "", fmt.Errorf("prompt failed: %w", err)
	}

	// Map the result to a simple string for the feedback struct
	switch result {
	case "👍 Worked":
		return "worked", nil
	case "👎 Didn't Work":
		return "failed", nil
	case "❌ Incorrect":
		return "incorrect", nil
	}

	return "", nil
}

func runSingleCommand(args []string) {
	input := strings.Join(args, " ")

	log := logger.GetLogger()
	log.WithField("input", input).Info("Processing single command")

	// Initialize AI client
	aiClient, err := ai.NewClient()
	if err != nil {
		output.PrintError(fmt.Sprintf("Failed to initialize AI client: %v", err))
		if strings.Contains(err.Error(), "no AI model available") {
			output.PrintInfo("💡 Run 'shell-agent download' to install an AI model first")
		}
		os.Exit(1)
	}

	response, err := aiClient.GenerateCommand(input)
	if err != nil {
		output.PrintError(fmt.Sprintf("Error generating command: %v", err))
		os.Exit(1)
	}

	output.PrintResponse(response)
}

func runStatusInline() {
	modelManager := ai.NewModelManager()
	currentModel := modelManager.GetCurrentModel()

	if currentModel != nil {
		if currentModel.Downloaded {
			output.PrintSuccess(fmt.Sprintf("Model: %s (Ready)", currentModel.Name))
		} else {
			output.PrintWarning(fmt.Sprintf("Model: %s (Not Downloaded)", currentModel.Name))
		}
	} else {
		output.PrintError("No model configured")
	}
}
