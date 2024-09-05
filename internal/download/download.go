package download

import (
	"os"
	"sync"

	git "github.com/go-git/go-git/v5"
	http "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rs/zerolog/log"

	"github.com/codekuu/tdws/internal/config"
	"github.com/codekuu/tdws/internal/module"
)

func cloneRepository(gitURL string, gitConfig config.GitConfig, storage string) {
	// Make sure gitURL ends with .git
	if gitURL[len(gitURL)-4:] != ".git" {
		gitURL += ".git"
	}
	// Make sure gitURL starts with https://
	if gitURL[:8] != "https://" {
		gitURL = "https://" + gitURL
	}

	// Check if the storage directory exists
	if _, err := os.Stat(storage); os.IsNotExist(err) {
		os.Mkdir(storage, 0755)
	}
	// Set the clone options
	cloneOptions := &git.CloneOptions{
		URL:             gitURL,
		InsecureSkipTLS: gitConfig.Insecure,
		Progress:        os.Stdout,
	}
	if gitConfig.Username != "" && gitConfig.Password != "" {
		cloneOptions.Auth = &http.BasicAuth{
			Username: gitConfig.Username,
			Password: gitConfig.Password,
		}
	}
	// Clone the repository
	_, err := git.PlainClone(storage, false, cloneOptions)
	if err != nil {
		// Delete the directory if the clone failed
		os.RemoveAll(storage)
		log.Fatal().Err(err).Msgf("Failed to clone the repository %s", gitURL)
	}
}

func WorkflowsActivities(cfg config.Config) {
	// Check if there is any modules to download
	if len(cfg.Modules) == 0 {
		log.Info().Msg("No modules to download")
		return
	}

	// Download the modules in parallel
	var wg sync.WaitGroup
	for _, mod := range cfg.Modules {
		// Check if the path is set
		if mod.GitUrl == "" {
			log.Fatal().Msgf("Path is not set for one of the modules")
		}

		// If the item.GitConfig is not set, use the global git config
		if mod.GitConfig == (config.GitConfig{}) {
			mod.GitConfig = cfg.GitConfig
		}

		// Skip the module if it is already downloaded
		storagePath := module.GetPathFromModule(cfg, mod)
		storagePathMainGo := storagePath + "/main.go"
		if _, err := os.Stat(storagePathMainGo); err == nil {
			if cfg.AlwaysDownloadModules {
				log.Info().Msgf("Module %s is already downloaded, deleting the old one", mod.GitUrl)
				os.RemoveAll(storagePath)
			} else {
				log.Info().Msgf("Module %s is already downloaded, wont download", mod.GitUrl)
				continue
			}
		}

		wg.Add(1)
		go func(mod config.Module) {
			defer wg.Done()
			storagePath := module.GetPathGitPathFromModule(cfg, mod)
			cloneRepository(mod.GitUrl, mod.GitConfig, storagePath)
		}(mod)
		// Wait for the go routines to finish
		wg.Wait()
	}
}
