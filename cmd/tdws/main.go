package main

import (
	"github.com/codekuu/tdws/internal/config"
	"github.com/codekuu/tdws/internal/download"
	"github.com/codekuu/tdws/internal/workers"
)

func main() {
	// Load the configuration
	cfg := config.LoadConfig()

	// Download the Workflows and Activities
	download.WorkflowsActivities(cfg)

	// Start the workers
	workers.Start(cfg)
}
