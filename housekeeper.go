package main

import (
	"context"
	"time"

	"github.com/0x4c6565/p.lee.io/pkg/storage"
	"github.com/rs/zerolog/log"
)

type Housekeeper struct {
	storage storage.Storage
}

func NewHousekeeper(storage storage.Storage) *Housekeeper {
	return &Housekeeper{
		storage: storage,
	}
}

func (h *Housekeeper) Start(ctx context.Context) error {
	log.Info().Msg("Starting housekeeper")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Housekeeper stopped")
			return ctx.Err()
		case <-time.After(60 * time.Second):
			h.work(ctx)
		}
	}
}

func (h Housekeeper) work(ctx context.Context) {
	expiredPastes, err := h.storage.GetExpired(ctx)
	if err != nil {
		log.Err(err).Msg("Failed to retrieve expired pastes")
	}

	log.Trace().Msg("Retrieving expired pastes")
	for _, paste := range expiredPastes {
		log.Debug().Msgf("Removing paste %s", paste.ID)
		err = h.storage.Delete(ctx, paste.ID)
		if err != nil {
			log.Err(err).Msg("Failed to burn expired paste")
		}
	}
}
