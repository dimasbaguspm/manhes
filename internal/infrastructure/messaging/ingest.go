package messaging

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/segmentio/kafka-go"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

type IngestHandler interface {
	Ingest(ctx context.Context, e domain.IngestRequested) error
}

type IngestConsumer struct {
	reader  *kafka.Reader
	handler IngestHandler
	workers int
}

func NewIngestConsumer(cfg *config.Config, h IngestHandler) *IngestConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.IngestTopic,
		GroupID: cfg.Kafka.IngestGroup,
	})
	return &IngestConsumer{reader: r, handler: h, workers: cfg.IngestWorkers}
}

func (c *IngestConsumer) Run(ctx context.Context) {
	log := slog.Default()
	defer c.reader.Close()
	sem := make(chan struct{}, c.workers)
	var wg sync.WaitGroup
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				wg.Wait()
				return
			}
			log.Error("ingest consumer: read message", "err", err)
			continue
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(msg kafka.Message) {
			defer wg.Done()
			defer func() { <-sem }()
			msgLog := log.With(
				slog.String("topic", msg.Topic),
				slog.Int("partition", msg.Partition),
				slog.Int64("offset", msg.Offset),
			)
			var e domain.IngestRequested
			if err := json.Unmarshal(msg.Value, &e); err != nil {
				msgLog.Error("ingest consumer: unmarshal", "err", err)
				return
			}
			msgLog = msgLog.With(slog.String("slug", e.Slug))
			if err := c.handler.Ingest(ctx, e); err != nil {
				msgLog.Error("ingest consumer: handle", "err", err)
			}
		}(msg)
	}
}
