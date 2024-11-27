package config

import (
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
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
	Branch         string    `json:"Branch"` // Branch to clone
}

type TemporalConfiguration struct {
	TaskQueue     string          `json:"TaskQueue"`     // Name of the temporal task queue
	ClientOptions tclient.Options `json:"ClientOptions"` // Temporal client options (see https://pkg.go.dev/go.temporal.io/sdk@v1.25.1/internal#ClientOptions)
	WorkerOptions tworker.Options `json:"WorkerOptions"` // Temporal worker options (see https://pkg.go.dev/go.temporal.io/sdk@v1.25.1/internal#WorkerOptions)
}

// TDWS configuration
type Config struct {
	AlwaysDownloadModules bool                  `json:"AlwaysDownloadModules"` // Always download the modules even if they are already downloaded (if not it will only download if the module doesnt exist)
	Storage               string                `json:"Storage"`               // Where to git clone the repositories (Modules)
	Temporal              TemporalConfiguration `json:"Temporal"`              // Temporal configuration
	GitConfig             GitConfig             `json:"GitConfig"`             // Git configuration, will be used to clone the repositories if not provided in the module
	Modules               []Module              `json:"Modules"`               // Modules to be cloned
}

var (
	CfgFile          string
	TdwsConfig       Config
	ModulesInStorage []string
)

func Init() {
	viper.SetConfigFile(CfgFile)

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("Failed to read the configuration file")
		os.Exit(1)
	}

	if err := viper.Unmarshal(&TdwsConfig); err != nil {
		log.Fatal().Err(err).Msg("Failed to unmarshal the configuration")
		os.Exit(1)
	}

	ModulesInStorage = getModulePaths()
}

// Returns a list of all modules in the storage with full path
func getModulePaths() []string {
	modules := []string{}
	for _, mod := range TdwsConfig.Modules {
		modulePath := GetPathFromModule(mod)
		modules = append(modules, modulePath)
	}
	return modules
}

// Returns the path of the module
func GetPathFromModule(module Module) string {
	parts := strings.Split(module.GitUrl, "/")
	repoNameWithExt := parts[len(parts)-1]
	repoName := strings.Split(repoNameWithExt, ".")[0]

	return path.Join(TdwsConfig.Storage, module.SubStorage, repoName, module.ModuleLocation)
}
