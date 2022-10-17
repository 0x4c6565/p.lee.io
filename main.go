package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/0x4c6565/p.lee.io/pkg/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	config, err := InitConfig()
	if err != nil {
		log.Fatal().Msgf("failed to initialise config: %s", err)
	}

	if config.Debug {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		log.Logger = log.With().Caller().Stack().Logger()
	}

	storage, err := storage.NewSQLStorage(
		config.Storage.SQL.Host,
		config.Storage.SQL.Port,
		config.Storage.SQL.User,
		config.Storage.SQL.Password,
		config.Storage.SQL.DB,
	)

	// storage := storage.NewInMemoryStorage()

	if err != nil {
		log.Fatal().Msgf("failed to initialise storage: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	api := NewAPI(storage, config)
	housekeeper := NewHousekeeper(storage)

	log.Info().Msg("Starting p.lee.io")

	g.Go(func() error {
		return api.Start(ctx)
	})

	g.Go(func() error {
		return housekeeper.Start(ctx)
	})

	log.Info().Msg("p.lee.io started")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		for range c {
			log.Info().Msg("Caught signal, shutting down..")
			cancel()
		}
	}()

	err = g.Wait()
	if err != nil && err != context.Canceled {
		log.Fatal().Err(err).Msg("Failed to shut down cleanly")
	}
}
