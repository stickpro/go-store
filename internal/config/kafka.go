package config

import "time"

type KafkaConfig struct {
	Brokers  []string            `yaml:"brokers"`
	Topics   KafkaTopicsConfig   `yaml:"topics"`
	Producer KafkaProducerConfig `yaml:"producer"`
	Consumer KafkaConsumerConfig `yaml:"consumer"`
	// OrderOutbox controls shipping order events to the ERP (topic store.orders).
	OrderOutbox KafkaOrderOutboxConfig `yaml:"order_outbox"`
}

type KafkaTopicsConfig struct {
	Products string `yaml:"products" default:"store.products"`
	Variants string `yaml:"variants" default:"store.product-variants"`
}

type KafkaProducerConfig struct {
	RequiredAcks string `yaml:"required_acks" default:"all"`
}

type KafkaConsumerConfig struct {
	GroupID string `yaml:"group_id" default:"go-store" usage:"consumer group id; must be unique per independent deployment sharing a broker"`
}

// KafkaOrderOutboxConfig tunes the transactional outbox behind the order
// exchange with the ERP. While disabled, no order events are recorded at all.
type KafkaOrderOutboxConfig struct {
	Enabled   bool          `yaml:"enabled" default:"true" usage:"record order events (created / updated / paid / cancelled) and publish them to the ERP topic"`
	Interval  time.Duration `yaml:"interval" default:"2s" usage:"how often the relay polls for pending order events"`
	BatchSize int           `yaml:"batch_size" default:"100" usage:"max order events published per relay transaction"`
	Retention time.Duration `yaml:"retention" default:"168h" usage:"how long published order events are kept before being deleted; 0 keeps them forever"`
}
