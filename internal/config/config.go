package config

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog/log"
	tclient "go.temporal.io/sdk/client"
	tworker "go.temporal.io/sdk/worker"
)

type GitConfig struct {
	Username string `json:"Username"` // Username
	Password string `json:"Password"` // Password or token
	Insecure bool   `json:"Insecure"` // Allow insecure server connections when using SSL
}
type Module struct {
	GitUrl         string    `json:"GitUrl"`         // Git repository path
	SubStorage     string    `json:"SubStorage"`     // Sub storage to store the module
	ModuleLocation string    `json:"ModuleLocation"` // Path to the module in the git repository
	GitConfig      GitConfig `json:"GitConfig"`
}

// TDWS configuration
type Config struct {
	AlwaysDownloadModules bool            `json:"AlwaysDownloadModules"` // Always download the modules even if they are already downloaded (if not it will only download if the module doesnt exist)
	TemporalWorkerName    string          `json:"TemporalWorkerName"`    // Name of the temporal worker, this will replace tworker.Options.Identity (TemporalWorkerOptions.Identity)
	TemporalTaskQueue     string          `json:"TemporalTaskQueue"`     // Name of the temporal task queue
	TemporalClientOptions tclient.Options `json:"TemporalClientOptions"` // Temporal client options (see https://pkg.go.dev/go.temporal.io/sdk@v1.25.1/internal#ClientOptions)
	TemporalWorkerOptions tworker.Options `json:"TemporalWorkerOptions"` // Temporal worker options (see https://pkg.go.dev/go.temporal.io/sdk@v1.25.1/internal#WorkerOptions)
	Storage               string          `json:"Storage"`               // Where to git clone the repositories (modules)
	GitConfig             GitConfig       `json:"GitConfig"`             // Git configuration, will be used to clone the repositories if not provided in the module
	Modules               []Module        `json:"Modules"`               // Modules to be cloned
}

// LoadConfig loads the configuration from the configuration file
func LoadConfig() Config {
	// Load the configuration file tdws.json or TDWS_CONFIG_FILE environment variable
	tdwsFileName := os.Getenv("TDWS_CONFIG_FILE")
	if tdwsFileName == "" {
		tdwsFileName = "tdws.json"
	}

	// Open the configuration file
	jsonFile, err := os.Open(tdwsFileName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read the configuration file")
	}

	// Base configuration
	config := Config{
		TemporalWorkerName: "tdws-worker",
		TemporalTaskQueue:  "tdws-task-queue",
		TemporalClientOptions: tclient.Options{
			HostPort: "temporal:7233",
		},
		Storage: "tdws-storage",
		GitConfig: GitConfig{
			Insecure: false,
		},
	}

	// Load the configuration from the json and update the base configuration
	err = json.NewDecoder(jsonFile).Decode(&config)
	if err != nil {
		// log config struct
		log.Fatal().Err(err).Msg("Failed to decode the configuration file")
	}

	// Create storage if it doesn't exist
	if _, err := os.Stat(config.Storage); os.IsNotExist(err) {
		os.Mkdir(config.Storage, 0755)
	}

	// Log the configuration loaded
	log.Info().Interface("config", json.NewDecoder(jsonFile)).Msg("Configuration loaded")

	return config
}
