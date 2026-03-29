package subscriber

import (
	"context"
	"log/slog"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

// DictionarySubscriberConfig holds dependencies for DictionarySubscriber.
type DictionarySubscriberConfig struct {
	DictionaryManager domain.DictionaryManager
	Cfg              *config.Config
}

// DictionarySubscriber handles dictionary-related events.
type DictionarySubscriber struct {
	dict domain.DictionaryManager
	cfg  *config.Config
	log  *slog.Logger
}

func NewDictionarySubscriber(cfg DictionarySubscriberConfig) *DictionarySubscriber {
	return &DictionarySubscriber{
		dict: cfg.DictionaryManager,
		cfg:  cfg.Cfg,
		log:  slog.With("component", "subscriber"),
	}
}

// HandleDictionaryRefreshed processes a DictionaryRefreshed event by calling Refresh.
func (s *DictionarySubscriber) HandleDictionaryRefreshed(ctx context.Context, e domain.DictionaryRefreshed) error {
	s.log.Info("[Dictionary Subscriber]: HandleDictionaryRefreshed: received event",
		"dictionaryID", e.DictionaryID,
	)
	_, err := s.dict.Refresh(ctx, e.DictionaryID)
	if err != nil {
		s.log.Error("[Dictionary Subscriber]: HandleDictionaryRefreshed: Refresh failed",
			"dictionaryID", e.DictionaryID,
			"err", err,
		)
		return err
	}
	s.log.Info("[Dictionary Subscriber]: HandleDictionaryRefreshed: completed",
		"dictionaryID", e.DictionaryID,
	)
	return nil
}
