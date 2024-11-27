package workers

import (
	"plugin"
	"sync"

	"github.com/rs/zerolog/log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/codekuu/tdws/internal/config"
	"github.com/codekuu/tdws/internal/module"
)

func dynamicRegister(w worker.Worker, pluginPath string) {
	// Open the plugin
	p, err := plugin.Open(pluginPath + "/main.so")
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to open the plugin %s", pluginPath)
	}

	// Look for exported symbol "TdwsRegister"
	tdwsRegisterSymbol, err := p.Lookup("TdwsRegister")
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to find the TdwsRegister symbol in the %s", pluginPath)
	} else {
		// Check if the symbol is a function
		tdwsRegisterFunc, ok := tdwsRegisterSymbol.(func(client worker.Worker))
		if !ok {
			log.Fatal().Msgf("Unable to load the workflows & activities from the %s", pluginPath)
		}

		// Call tdwsRegisterSymbol function to register workflow methods
		tdwsRegisterFunc(w)
	}
}

// Load and register the modules into the worker in parallel go routines
func loadGoRegisterModules(w worker.Worker) {
	log.Info().Msgf("Found %d modules in the storage, loading them into the worker...", len(config.ModulesInStorage))
	var wg sync.WaitGroup
	for _, modulePath := range config.ModulesInStorage {
		wg.Add(1)
		go func(modulePath string) {
			defer wg.Done()
			// Delete the module if it's built (main.so file exists)
			if module.IsBuilt(modulePath) {
				log.Info().Msgf("Module %s is already built, deleting (this is to ensure that plugin is built with same go version)", modulePath)
				module.Delete(modulePath)

			}
			log.Info().Msgf("Building the module %s", modulePath)
			module.Build(modulePath)

			// Register the workflows and activities
			dynamicRegister(w, modulePath)
		}(modulePath)
	}
	// Wait for the go routines to finish
	wg.Wait()
	log.Info().Msg("Modules loaded into the worker, starting the worker...")
}

// StartWorkerGo starts the Go worker
func StartWorkerGo() {
	log.Info().Msg("Starting the Go worker")

	// Create the client object just once per process
	config.TdwsConfig.Temporal.ClientOptions.Identity = config.TdwsConfig.Temporal.ClientOptions.Identity + "-go"
	c, err := client.Dial(config.TdwsConfig.Temporal.ClientOptions)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create Temporal client")
	}
	defer c.Close()

	// This worker hosts both Workflow and Activity functions
	w := worker.New(c, config.TdwsConfig.Temporal.TaskQueue, config.TdwsConfig.Temporal.WorkerOptions)

	loadGoRegisterModules(w)

	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to start worker")
	}
}
