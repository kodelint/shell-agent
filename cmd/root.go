package cmd

import (
	"fmt"
	"os"

	"github.com/kodelint/shell-agent/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	debug   bool
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "shell-agent",
	Short: "AI-powered shell command generator",
	Long: `Shell Agent is an AI-powered CLI tool that generates shell commands
based on natural language input using local AI models.

Examples:
  shell-agent                                    # Start interactive mode
  shell-agent "list all files in current directory"
  shell-agent "find all python files modified in last 7 days"
  shell-agent "compress folder into tar.gz"`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			runInteractiveMode()
		} else {
			runSingleCommand(args)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.shell-agent.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Bind flags to viper
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".shell-agent")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("debug") {
			fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
		}
	}

	// Initialize logger
	logger.InitLogger(viper.GetBool("debug"), viper.GetBool("verbose"))
}
