package download

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	httpproto "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rs/zerolog/log"

	"github.com/codekuu/tdws/internal/config"
	"github.com/codekuu/tdws/internal/module"
)

// configFIle returns the filepath of the downloaded configuration file
func ConfigFile(configFileUrl string) string {
	// Parse the URL
	u, err := url.Parse(configFileUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse the URL")
	}
	// Get the filename
	filename := u.Path
	if filename == "" {
		log.Fatal().Msg("No filename in the URL")
	}
	// Open the file
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create the file")
	}
	// Close the file
	defer file.Close()
	// Download the file
	resp, err := http.Get(configFileUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to download the file")
	}
	// Close the response
	defer resp.Body.Close()
	// Copy the response to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to copy the response to the file")
	}

	return filename
}

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
		Auth:            nil,
		Progress:        os.Stdout,
	}
	if gitConfig.Username != "" && gitConfig.Password != "" {
		cloneOptions.Auth = &httpproto.BasicAuth{
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

func getBranches(repo *git.Repository) []string {
	// Get the branches
	refs, err := repo.Branches()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get the branches")
	}
	// Get the branch names
	var branches []string
	refs.ForEach(func(ref *plumbing.Reference) error {
		branches = append(branches, ref.Name().Short())
		return nil
	})
	return branches
}

func checkOutBranch(storage string, branch string) {
	// Open the repository
	repo, err := git.PlainOpen(storage)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to open the repository %s", storage)
	}
	// Get the worktree
	wt, err := repo.Worktree()
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to get the worktree for the repository %s", storage)
	}
	// Check out the branch
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(branch),
	})
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to checkout the branch %s, branches that exists: %v", branch, getBranches(repo))
	}
}

func WorkflowsActivities() {
	// Check if there is any modules to download
	if len(config.TdwsConfig.Modules) == 0 {
		log.Info().Msg("No modules to download")
		return
	}

	// Download the modules in parallel
	var wg sync.WaitGroup
	for _, mod := range config.TdwsConfig.Modules {
		// Check if the path is set
		if mod.GitUrl == "" {
			log.Fatal().Msgf("Path is not set for one of the modules")
		}

		// If the item.GitConfig is not set, use the global git config
		if mod.GitConfig == (config.GitConfig{}) {
			mod.GitConfig = config.TdwsConfig.GitConfig
		}

		// Skip the module if it is already downloaded
		modulePath := config.GetPathFromModule(mod)
		modulePathMainGo := modulePath + "/main.go"
		if _, err := os.Stat(modulePathMainGo); err == nil {
			if config.TdwsConfig.AlwaysDownloadModules {
				log.Info().Msgf("Module %s is already downloaded, deleting the old one", mod.GitUrl)
				os.RemoveAll(modulePath)
			} else {
				log.Info().Msgf("Module %s is already downloaded, wont download", mod.GitUrl)
				continue
			}
		}
		module.GetMetadata(modulePath)

		wg.Add(1)
		go func(mod config.Module) {
			defer wg.Done()
			modulePath := config.GetGitPathFromModule(mod)
			cloneRepository(mod.GitUrl, mod.GitConfig, modulePath)
			if mod.Branch != "" {
				checkOutBranch(modulePath, mod.Branch)
			}
		}(mod)
		// Wait for the go routines to finish
		wg.Wait()
	}
}
