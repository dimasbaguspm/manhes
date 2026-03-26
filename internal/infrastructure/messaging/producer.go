package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

var _ domain.EventPublisher = (*Producer)(nil)

type Producer struct {
	ingest *kafka.Writer
	sync   *kafka.Writer
}

func NewProducer(cfg *config.Config) *Producer {
	return &Producer{
		ingest: newWriter(cfg.Kafka.Brokers, cfg.Kafka.IngestTopic),
		sync:   newWriter(cfg.Kafka.Brokers, cfg.Kafka.SyncTopic),
	}
}

func (p *Producer) PublishIngestRequested(ctx context.Context, e domain.IngestRequested) error {
	return publish(ctx, p.ingest, e)
}

func (p *Producer) PublishChapterDownloaded(ctx context.Context, e domain.ChapterDownloaded) error {
	return publish(ctx, p.sync, e)
}

func (p *Producer) Close() error {
	var last error
	if err := p.ingest.Close(); err != nil {
		last = err
	}
	if err := p.sync.Close(); err != nil {
		last = err
	}
	return last
}

func newWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func publish(ctx context.Context, w *kafka.Writer, event any) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return w.WriteMessages(ctx, kafka.Message{Value: data})
}
