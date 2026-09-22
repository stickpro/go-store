package kafka

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/messaging/topics"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/twmb/franz-go/pkg/kgo"
)

type HandlerFunc[T any] func(ctx context.Context, payload T) error

// kgoLogger bridges franz-go's internal client logger (join/sync/heartbeat,
// rebalances, coordinator lookups - none of which surface through
// fetches.Errors()) into the app logger, so consumer-group problems are
// visible instead of silently swallowed.
type kgoLogger struct {
	log logger.Logger
}

func (l kgoLogger) Level() kgo.LogLevel { return kgo.LogLevelInfo }

func (l kgoLogger) Log(level kgo.LogLevel, msg string, keyvals ...any) {
	args := append([]any{"kgo_level", level.String()}, keyvals...)
	switch level {
	case kgo.LogLevelError:
		l.log.Errorw(msg, args...)
	case kgo.LogLevelWarn:
		l.log.Warnw(msg, args...)
	default:
		l.log.Infow(msg, args...)
	}
}

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
		kgo.WithLogger(kgoLogger{log: log}),
		kgo.OnPartitionsAssigned(func(_ context.Context, _ *kgo.Client, assigned map[string][]int32) {
			log.Infow("kafka consumer: partitions assigned", "assigned", assigned)
		}),
		kgo.OnPartitionsRevoked(func(_ context.Context, _ *kgo.Client, revoked map[string][]int32) {
			log.Infow("kafka consumer: partitions revoked", "revoked", revoked)
		}),
		kgo.OnPartitionsLost(func(_ context.Context, _ *kgo.Client, lost map[string][]int32) {
			log.Infow("kafka consumer: partitions lost", "lost", lost)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer: new client: %w", err)
	}

	log.Infow("kafka consumer: configured",
		"brokers", cfg.Brokers,
		"group_id", cfg.Consumer.GroupID,
		"topic", topics.Products,
	)

	return &Consumer{client: client, logger: log}, nil
}

func (c *Consumer) Run(ctx context.Context, onProduct HandlerFunc[contracts.ProductPayload]) error {
	for {
		fetches := c.client.PollFetches(ctx)

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, e := range errs {
				c.logger.Error("kafka consumer: fetch error", "topic", e.Topic, "error", e.Err)
			}
		}

		if n := fetches.NumRecords(); n > 0 {
			c.logger.Infow("kafka consumer: polled records", "count", n)
		}

		fetches.EachRecord(func(r *kgo.Record) {
			c.logger.Infow("kafka consumer: record received",
				"topic", r.Topic,
				"partition", r.Partition,
				"offset", r.Offset,
				"key", string(r.Key),
			)

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
				return
			}

			c.logger.Infow("kafka consumer: record committed",
				"topic", r.Topic,
				"partition", r.Partition,
				"offset", r.Offset,
			)
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
