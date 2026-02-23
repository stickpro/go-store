package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/messaging/topics"
	"github.com/twmb/franz-go/pkg/kgo"
)

type IProducer interface {
	PublishProductEvent(ctx context.Context, payload contracts.ProductPayload) error
	Close()
}

type Producer struct {
	client *kgo.Client
	cfg    config.KafkaTopicsConfig
}

func NewProducer(cfg config.KafkaConfig) (*Producer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ProducerBatchMaxBytes(1_000_000),
	}

	if cfg.Producer.RequiredAcks == "all" {
		opts = append(opts, kgo.RequiredAcks(kgo.AllISRAcks()))
	} else {
		opts = append(opts, kgo.RequiredAcks(kgo.LeaderAck()))
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("kafka producer: new client: %w", err)
	}

	return &Producer{client: client, cfg: cfg.Topics}, nil
}

func (p *Producer) PublishProductEvent(ctx context.Context, payload contracts.ProductPayload) error {
	return p.publish(ctx, topics.Products, payload.ExternalID, payload)
}

func (p *Producer) publish(ctx context.Context, topic, key string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("kafka producer: marshal: %w", err)
	}

	record := &kgo.Record{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	}

	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("kafka producer: produce to %s: %w", topic, err)
	}

	return nil
}

func (p *Producer) Close() {
	p.client.Close()
}
