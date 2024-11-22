package workers

import (
	"github.com/rs/zerolog/log"

	"github.com/codekuu/tdws/internal/config"
)

func Start(cfg config.Config) {
	log.Info().Msg("Starting the worker")
	StartWorkerGo(cfg)
}
