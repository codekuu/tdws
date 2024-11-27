package workers

import (
	"github.com/rs/zerolog/log"
)

func Start() {
	log.Info().Msg("Starting the worker")
	StartWorkerGo()
}
