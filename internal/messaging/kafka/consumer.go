package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/messaging/topics"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/twmb/franz-go/pkg/kgo"
)

type HandlerFunc[T any] func(ctx context.Context, payload T) error

type Consumer struct {
	client *kgo.Client
	logger logger.Logger
}

func NewConsumer(cfg config.KafkaConfig, log logger.Logger) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.Consumer.GroupID),
		kgo.ConsumeTopics(topics.Products),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer: new client: %w", err)
	}

	return &Consumer{client: client, logger: log}, nil
}

func (c *Consumer) Run(ctx context.Context, onProduct HandlerFunc[contracts.ProductPayload]) error {
	for {
		fetches := c.client.PollFetches(ctx)

		if ctx.Err() != nil {
			return nil
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, e := range errs {
				c.logger.Error("kafka consumer: fetch error", "topic", e.Topic, "error", e.Err)
			}
		}

		fetches.EachRecord(func(r *kgo.Record) {
			var err error
			switch r.Topic {
			case topics.Products:
				err = dispatch(ctx, r.Value, onProduct)
			default:
				c.logger.Warn("kafka consumer: unknown topic", "topic", r.Topic)
			}

			if err != nil {
				c.logger.Errorw("kafka consumer: handler error",
					"topic", r.Topic,
					"error", err,
				)
				return
			}

			if err := c.client.CommitRecords(ctx, r); err != nil {
				c.logger.Error("kafka consumer: commit error", "topic", r.Topic, "error", err)
			}
		})
	}
}

func (c *Consumer) Close() {
	c.client.Close()
}

func dispatch[T any](ctx context.Context, data []byte, handler HandlerFunc[T]) error {
	var payload T
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	return handler(ctx, payload)
}
