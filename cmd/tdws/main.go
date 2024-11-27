package main

import (
	"net/url"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/codekuu/tdws/internal/config"
	"github.com/codekuu/tdws/internal/download"
	"github.com/codekuu/tdws/internal/workers"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "tdws",
		Short: "Temporal Dynamic Worker Spawner",
		Long:  "Temporal.io Dynamic Worker Spawner (TDWS) is a temporal worker that downloads workflows and activities from a remote git repository and loads them into the worker.",
		Run: func(cmd *cobra.Command, args []string) {
			// Download the Workflows and Activities
			download.WorkflowsActivities(config.TdwsConfig)

			// Start the workers
			workers.Start(config.TdwsConfig)
		},
	}
)

func main() {
	if config.CfgFile != "" {
		_, err := url.ParseRequestURI(config.CfgFile)
		if err == nil {
			config.CfgFile = download.ConfigFile(config.CfgFile)
		}
	} else {
		tdwsFileName := os.Getenv("TDWS_CONFIG_FILE")
		if tdwsFileName == "" {
			config.CfgFile = "tdws.json"
		}

	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&config.CfgFile, "config", "c", "", "Configuration file or URL")

	// Initialize the configuration
	cobra.OnInitialize(config.Init)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Failed to execute the root command")
		os.Exit(1)
	}
}
