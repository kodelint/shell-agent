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
		green.Printf("   %s\n", response.Explanation)
		fmt.Println()
	}

	// Print command in bold white
	white.Println("🚀 Generated Command:")
	white.Printf("   %s\n", response.Command)

	// Print warning if exists
	if response.Warning != "" {
		fmt.Println()
		yellow.Println("⚠️  Warning:")
		yellow.Printf("   %s\n", response.Warning)
	}

	// Print confidence if available
	if response.Confidence > 0 {
		fmt.Println()
		if response.Confidence >= 0.8 {
			green.Printf("✅ Confidence: %.0f%%\n", response.Confidence*100)
		} else if response.Confidence >= 0.6 {
			yellow.Printf("⚠️  Confidence: %.0f%%\n", response.Confidence*100)
		} else {
			red.Printf("❌ Confidence: %.0f%% (Low)\n", response.Confidence*100)
		}
	}

	fmt.Println()
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

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", command)
	default:
		cmd = exec.Command("sh", "-c", command)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
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
