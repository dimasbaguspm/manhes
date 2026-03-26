package messaging

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

type SyncHandler interface {
	HandleChapterDownloaded(ctx context.Context, e domain.ChapterDownloaded) error
}

type SyncConsumer struct {
	reader  *kafka.Reader
	handler SyncHandler
}

func NewSyncConsumer(cfg *config.Config, h SyncHandler) *SyncConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.SyncTopic,
		GroupID: cfg.Kafka.SyncGroup,
	})
	return &SyncConsumer{reader: r, handler: h}
}

func (c *SyncConsumer) Run(ctx context.Context) {
	log := slog.Default()
	defer c.reader.Close()
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Error("sync consumer: read message", "err", err)
			continue
		}
		msgLog := log.With(
			slog.String("topic", msg.Topic),
			slog.Int("partition", msg.Partition),
			slog.Int64("offset", msg.Offset),
		)
		var e domain.ChapterDownloaded
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			msgLog.Error("sync consumer: unmarshal", "err", err)
			continue
		}
		msgLog = msgLog.With(slog.String("slug", e.Slug))
		if err := c.handler.HandleChapterDownloaded(ctx, e); err != nil {
			msgLog.Error("sync consumer: handle", "err", err)
		}
	}
}
