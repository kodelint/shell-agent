package output

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/kodelint/shell-agent/internal/ai"
	"github.com/kodelint/shell-agent/internal/system"
	"github.com/manifoldco/promptui"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var (
	green     = color.New(color.FgGreen)
	white     = color.New(color.FgWhite, color.Bold)
	red       = color.New(color.FgRed)
	yellow    = color.New(color.FgYellow)
	blue      = color.New(color.FgBlue)
	cyan      = color.New(color.FgCyan)
	magenta   = color.New(color.FgMagenta)
	boldGreen = color.New(color.FgGreen, color.Bold)
)

type StatusInfo struct {
	ModelManager *ai.ModelManager
	SystemInfo   *system.SystemInfo
	ModelOnly    bool
}

func PrintWelcome() {
	fmt.Println()
	cyan.Println("🤖 Shell Agent - AI-Powered Command Generator")
	cyan.Println("==========================================")
	green.Println("✨ Welcome to your intelligent shell assistant!")
	green.Println("💬 Type your requests in natural language")
	green.Println("📝 Available commands: help, status, clear, exit")
	fmt.Println()
}

func PrintPrompt() {
	boldGreen.Print("🤖 shell-agent ➤ ")
}

func PrintThinking() {
	magenta.Print("🧠 Thinking... ")
	fmt.Println()
}

func PrintResponse(response *ai.CommandResponse) {
	fmt.Println()

	// Print explanation in green
	if response.Explanation != "" {
		green.Println("💡 Explanation:")
		streamString("   "+response.Explanation+"\n", green, 20*time.Millisecond)
		fmt.Println()
	}

	// Print command in bold white
	white.Println("🚀 Generated Command:")
	streamString("   "+response.Command+"\n", white, 10*time.Millisecond)

	// Print warning if exists
	if response.Warning != "" {
		fmt.Println()
		yellow.Println("⚠️  Warning:")
		streamString("   "+response.Warning+"\n", yellow, 20*time.Millisecond)
	}

	// Print confidence if available
	if response.Confidence > 0 {
		fmt.Println()
		prefix := "✅ Confidence: "
		var confidenceColor *color.Color
		if response.Confidence >= 0.8 {
			confidenceColor = green
		} else if response.Confidence >= 0.6 {
			confidenceColor = yellow
			prefix = "⚠️  Confidence: "
		} else {
			confidenceColor = red
			prefix = "❌ Confidence: "
		}

		// We'll print the prefix first without streaming.
		confidenceColor.Print(prefix)
		// Now we'll stream the rest of the message.
		confidenceMessage := fmt.Sprintf("%.0f%%\n", response.Confidence*100)
		streamString(confidenceMessage, confidenceColor, 20*time.Millisecond)
	}

	fmt.Println()
}

// PromptForFeedback asks the user to rate the last command.
func PromptForFeedback() (string, error) {
	// Options for the user to choose from
	options := []string{"👍 Worked", "👎 Didn't Work", "❌ Incorrect"}

	// Use promptui to create a menu for a rich UI
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

func PrintError(message string) {
	red.Printf("❌ Error: %s\n", message)
}

func PrintSuccess(message string) {
	green.Printf("✅ %s\n", message)
}

func PrintWarning(message string) {
	yellow.Printf("⚠️  %s\n", message)
}

func PrintInfo(message string) {
	blue.Printf("ℹ️  %s\n", message)
}

func PrintGoodbye() {
	fmt.Println()
	cyan.Println("👋 Thank you for using Shell Agent!")
	cyan.Println("🚀 May your commands be swift and your deployments bug-free!")
	fmt.Println()
}

func PrintStatus(status *StatusInfo) {
	fmt.Println()
	cyan.Println("📊 Shell Agent Status")
	cyan.Println("=====================")
	fmt.Println()

	// Model information
	currentModel := status.ModelManager.GetCurrentModel()
	if currentModel != nil {
		boldGreen.Printf("🤖 Current Model: %s\n", currentModel.Name)
		green.Printf("📁 Model Path: %s\n", status.ModelManager.GetModelPath())

		if currentModel.Downloaded {
			green.Println("✅ Model Status: Ready")
		} else {
			yellow.Println("⚠️  Model Status: Not Downloaded")
			PrintInfo("💡 Run 'shell-agent download' to install this model")
		}
	} else {
		red.Println("❌ No model configured")
		PrintInfo("💡 Run 'shell-agent download' to install a model")
	}

	if !status.ModelOnly {
		fmt.Println()

		// System information
		sysInfo := status.SystemInfo.GetInfo()
		boldGreen.Println("💻 System Information:")
		fmt.Printf("   🖥️  OS: %s\n", sysInfo.OS)
		fmt.Printf("   🏗️  Architecture: %s\n", sysInfo.Arch)
		fmt.Printf("   🐹 Go Version: %s\n", sysInfo.GoVersion)

		fmt.Println()

		// Configuration
		boldGreen.Println("⚙️  Configuration:")
		if sysInfo.ConfigFile != "" {
			fmt.Printf("   📋 Config File: %s\n", sysInfo.ConfigFile)
		} else {
			fmt.Printf("   📋 Config File: Not found (using defaults)\n")
		}
		fmt.Printf("   🐛 Debug Mode: %v\n", sysInfo.Debug)
		fmt.Printf("   📝 Verbose Mode: %v\n", sysInfo.Verbose)
	}

	fmt.Println()
}

func PrintHelp() {
	fmt.Println()
	cyan.Println("🆘 Shell Agent Help & Commands")
	cyan.Println("===============================")
	fmt.Println()

	boldGreen.Println("📝 Built-in Commands:")
	green.Println("  help, h     - Show this help message")
	green.Println("  status      - Show current model status")
	green.Println("  clear, cls  - Clear the screen")
	green.Println("  exit, quit, q - Exit shell agent")
	fmt.Println()

	boldGreen.Println("💬 Example Natural Language Requests:")
	green.Println("  • 'list all files in current directory'")
	green.Println("  • 'find all .py files modified in the last 7 days'")
	green.Println("  • 'create a backup of my documents folder'")
	green.Println("  • 'show disk usage of current directory'")
	green.Println("  • 'compress this folder into a tar.gz file'")
	green.Println("  • 'find files larger than 100MB'")
	green.Println("  • 'show running processes using port 8080'")
	fmt.Println()

	boldGreen.Println("🎯 Tips for Better Results:")
	green.Println("  • Be specific about what you want to accomplish")
	green.Println("  • Mention file types, directories, or specific criteria")
	green.Println("  • Ask for explanations if you're unsure about a command")
	fmt.Println()
}

func ClearScreen() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func PromptExecuteCommand() bool {
	prompt := promptui.Prompt{
		Label:     "Execute this command",
		IsConfirm: true,
		Default:   "n",
	}

	result, err := prompt.Run()
	if err != nil {
		return false
	}

	return strings.ToLower(result) == "y"
}

func ExecuteCommand(command string) error {
	var cmd *exec.Cmd

	// Split the command string into the command name and its arguments.
	// This helps prevent command injection vulnerabilities.
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("command string is empty")
	}
	name := parts[0]
	args := parts[1:]

	switch runtime.GOOS {
	case "windows":
		// On Windows, use "cmd /C" to execute the command.
		// Note: We use "cmd" as the name and pass the rest as arguments.
		cmd = exec.Command("cmd", append([]string{"/C", name}, args...)...)
	default:
		// On other systems, like Linux and macOS, we can execute the command directly.
		cmd = exec.Command(name, args...)
	}

	// Connect the command's standard input, output, and error streams
	// to the current process's streams so you can see the output in real-time.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run the command and return any error that occurs.
	// This includes cases where the command exits with a non-zero status.
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute command '%s': %w", command, err)
	}

	return nil
}

func PrintAvailableModels(models []ai.ModelInfo) {
	fmt.Println()
	cyan.Println("📋 Available AI Models")
	cyan.Println("======================")
	fmt.Println()

	for _, model := range models {
		if model.Downloaded {
			green.Printf("✅ %s - %s (Downloaded)\n", model.Name, model.Description)
		} else {
			fmt.Printf("⬜ %s - %s (Available for download)\n", model.Name, model.Description)
		}
		fmt.Printf("   📦 Size: %s | 🏷️  Type: %s\n", model.Size, model.Type)
		fmt.Println()
	}
}

func PromptModelSelection(models []ai.ModelInfo) (string, error) {
	items := make([]string, len(models))
	for i, model := range models {
		status := "Available"
		if model.Downloaded {
			status = "Downloaded"
		}
		items[i] = fmt.Sprintf("%s (%s) - %s", model.Name, status, model.Description)
	}

	prompt := promptui.Select{
		Label: "Select a model to download",
		Items: items,
		Size:  len(items),
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✅ Selected: {{ . | green }}",
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return models[index].Name, nil
}

func PrintSetupWelcome() {
	fmt.Println()
	cyan.Println("🚀 Shell Agent Setup")
	cyan.Println("====================")
	fmt.Println()
	green.Println("This setup will:")
	green.Println("  ✅ Install Ollama (if needed)")
	green.Println("  ✅ Start Ollama service")
	green.Println("  ✅ Download a recommended AI model")
	green.Println("  ✅ Create default configuration")
	fmt.Println()
}

func PrintSetupComplete() {
	fmt.Println()
	boldGreen.Println("🎉 Setup Complete!")
	boldGreen.Println("==================")
	fmt.Println()
	green.Println("Shell Agent is now ready to use!")
	fmt.Println()
	boldGreen.Println("Quick Start:")
	green.Println("  shell-agent                    # Start interactive mode")
	green.Println("  shell-agent status            # Check system status")
	green.Println("  shell-agent \"list files\"      # Generate a command")
	fmt.Println()
	green.Println("💡 Tips:")
	green.Println("  • Be specific in your requests")
	green.Println("  • Always review commands before executing")
	green.Println("  • Use 'help' for assistance in interactive mode")
	fmt.Println()
}

func PromptSetupConfirm() bool {
	prompt := promptui.Prompt{
		Label:     "Continue with setup",
		IsConfirm: true,
		Default:   "y",
	}

	result, err := prompt.Run()
	if err != nil {
		return false
	}

	return strings.ToLower(result) == "y"
}

// streamString prints a string character by character with a delay.
func streamString(text string, c *color.Color, delay time.Duration) {
	for _, char := range text {
		c.Fprint(os.Stdout, string(char))
		time.Sleep(delay)
	}
}
