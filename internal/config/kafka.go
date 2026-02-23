package config

type KafkaConfig struct {
	Brokers  []string            `yaml:"brokers"`
	Topics   KafkaTopicsConfig   `yaml:"topics"`
	Producer KafkaProducerConfig `yaml:"producer"`
	Consumer KafkaConsumerConfig `yaml:"consumer"`
}

type KafkaTopicsConfig struct {
	Products string `yaml:"products" default:"store.products"`
	Variants string `yaml:"variants" default:"store.product-variants"`
}

type KafkaProducerConfig struct {
	RequiredAcks string `yaml:"required_acks" default:"all"`
}

type KafkaConsumerConfig struct {
	GroupID string `yaml:"group_id" default:"go-store"`
}
