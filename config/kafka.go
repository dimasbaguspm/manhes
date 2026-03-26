package config

type KafkaConfig struct {
	Brokers     []string
	IngestTopic string
	SyncTopic   string
	IngestGroup string
	SyncGroup   string
}

func loadKafkaConfig() KafkaConfig {
	return KafkaConfig{
		Brokers:     envList("KAFKA_BROKERS", "localhost:9092"),
		IngestTopic: "manga.ingest",
		SyncTopic:   "manga.sync",
		IngestGroup: "manhes-ingest",
		SyncGroup:   "manhes-sync",
	}
}
