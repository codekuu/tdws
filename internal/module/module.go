package module

import (
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/codekuu/tdws/internal/config"
)

func validateMetadata(metadata config.ModuleMetadata) {
	if metadata.Name == "" {
		log.Fatal().Msg("Module name is required")
	}
	if metadata.Description == "" {
		log.Fatal().Msg("Module description is required")
	}
	if metadata.Version == "" {
		log.Fatal().Msg("Module version is required")
	}
	if metadata.Creator == "" {
		log.Fatal().Msg("Module creator is required")
	}
	if metadata.Contact == "" {
		log.Fatal().Msg("Module contact is required")
	}
	if !slices.Contains(config.SupportedLangs, metadata.Lang) {
		log.Fatal().Msgf("Module language is not an allowed language. Allowed languages: %v", config.SupportedLangs)
	}
}

func GetMetadata(modulePath string) config.ModuleMetadata {
	// Look in the module metadata.json for information about the workers
	metadataFile, err := os.Open(modulePath + "/metadata.json")
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to open the metadata file for %s", modulePath)
	}

	// Read the metadata file
	metadata := config.ModuleMetadata{}
	err = json.NewDecoder(metadataFile).Decode(&metadata)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to read the metadata file for %s", modulePath)
	}

	return metadata
}

// Builds the module to a plugin (main.so file)
func Build(modulePath string) {
	mainGo := path.Join(modulePath, "main.go")
	mainSo := path.Join(modulePath, "main.so")
	// if the module is already built, delete it before building
	if _, err := os.Stat(mainSo); err == nil {
		log.Info().Msgf("Deleting the existing main.so file %s", mainSo)
		err := os.Remove(mainSo)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to delete the existing main.so file %s", mainSo)
		}
	}
	log.Info().Msgf("Running go build -buildmode=plugin -o %s %s", mainSo, mainGo)
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", mainSo, mainGo)
	err := cmd.Run()
	// change the current working directory back
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to build the module %s", modulePath)
	}
}

func Delete(modulePath string) {
	mainSo := path.Join(modulePath, "main.so")
	err := os.Remove(mainSo)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to delete the main.so file %s", mainSo)
	}
}

// Returns true if the module is built (main.so file exists)
func IsBuilt(modulePath string) bool {
	_, err := os.Stat(path.Join(modulePath, "main.so"))
	return err == nil
}

// Returns the path of the git repository
func GetPathGitPathFromModule(cfg config.Config, module config.Module) string {
	parts := strings.Split(module.GitUrl, "/")
	repoNameWithExt := parts[len(parts)-1]
	repoName := strings.Split(repoNameWithExt, ".")[0]

	return path.Join(cfg.Storage, module.SubStorage, repoName)
}
