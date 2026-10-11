package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_order_outbox"
	"github.com/stickpro/go-store/pkg/logger"
)

const (
	// orderOutboxPublishTimeout bounds one Kafka publish, so a broker outage
	// can't hold the relay transaction (and its row locks) open indefinitely.
	orderOutboxPublishTimeout = 10 * time.Second
	// orderOutboxPurgeEvery is how often published events past retention are deleted.
	orderOutboxPurgeEvery = time.Hour
)

// OrderEventPublisher is the Kafka side of the relay.
type OrderEventPublisher interface {
	PublishOrderEvent(ctx context.Context, orderID uuid.UUID, env contracts.OrderEnvelope) error
}

// OrderOutboxRelay moves order events from the order_outbox table to Kafka.
// Delivery is at-least-once: an event is marked published only after the
// broker acknowledged it, so a crash in between re-sends it (the receiver
// deduplicates on event_id). Events are published strictly in outbox order —
// a failed publish stops the batch and is retried on the next tick rather
// than skipped.
type OrderOutboxRelay struct {
	storage   storage.IStorage
	publisher OrderEventPublisher
	cfg       config.KafkaOrderOutboxConfig
	logger    logger.Logger
}

func NewOrderOutboxRelay(
	st storage.IStorage,
	publisher OrderEventPublisher,
	cfg config.KafkaOrderOutboxConfig,
	log logger.Logger,
) *OrderOutboxRelay {
	if cfg.Interval <= 0 {
		cfg.Interval = 2 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	return &OrderOutboxRelay{storage: st, publisher: publisher, cfg: cfg, logger: log}
}

// Run polls the outbox until ctx is cancelled.
func (r *OrderOutboxRelay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.Interval)
	defer ticker.Stop()

	var lastPurge time.Time
	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Order outbox relay stopped")
			return
		case <-ticker.C:
		}

		r.drain(ctx)

		if r.cfg.Retention > 0 && time.Since(lastPurge) >= orderOutboxPurgeEvery {
			lastPurge = time.Now()
			r.purge(ctx)
		}
	}
}

// drain publishes batch after batch until the outbox is empty or a publish fails.
func (r *OrderOutboxRelay) drain(ctx context.Context) {
	for ctx.Err() == nil {
		sent, err := r.relayBatch(ctx)
		if err != nil {
			if ctx.Err() == nil {
				r.logger.Errorw("order outbox relay: publish batch", "sent", sent, "error", err)
			}
			return
		}
		if sent < r.cfg.BatchSize {
			return
		}
	}
}

// relayBatch publishes up to BatchSize pending events in one transaction and
// returns how many were sent. On a publish failure the attempt is recorded on
// the failed row, the rows already sent stay marked published, and the error
// is returned.
func (r *OrderOutboxRelay) relayBatch(ctx context.Context) (int, error) {
	var (
		sent    int
		sendErr error
	)

	txErr := repository.BeginTxFunc(ctx, r.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		q := r.storage.OrderOutbox(repository.WithTx(tx))

		rows, err := q.LockPending(ctx, int32(r.cfg.BatchSize)) //nolint:gosec // batch size is a small config value
		if err != nil {
			return fmt.Errorf("lock pending: %w", err)
		}

		for _, row := range rows {
			if err := r.publish(ctx, row); err != nil {
				sendErr = fmt.Errorf("event %s (%s): %w", row.EventID, row.EventType, err)
				return q.MarkFailed(ctx, repository_order_outbox.MarkFailedParams{
					ID:        row.ID,
					LastError: pgtype.Text{String: err.Error(), Valid: true},
				})
			}
			if err := q.MarkPublished(ctx, row.ID); err != nil {
				return fmt.Errorf("mark published: %w", err)
			}
			sent++
		}
		return nil
	})
	if txErr != nil {
		// Rolled back: nothing is marked, every event of the batch is retried.
		return 0, txErr
	}
	return sent, sendErr
}

func (r *OrderOutboxRelay) publish(ctx context.Context, row *models.OrderOutbox) error {
	ctx, cancel := context.WithTimeout(ctx, orderOutboxPublishTimeout)
	defer cancel()

	return r.publisher.PublishOrderEvent(ctx, row.OrderID, orderEnvelope(row))
}

func (r *OrderOutboxRelay) purge(ctx context.Context) {
	cutoff := pgtype.Timestamptz{Time: time.Now().Add(-r.cfg.Retention), Valid: true}
	n, err := r.storage.OrderOutbox().DeletePublishedBefore(ctx, cutoff)
	if err != nil {
		if ctx.Err() == nil {
			r.logger.Errorw("order outbox relay: purge published", "error", err)
		}
		return
	}
	if n > 0 {
		r.logger.Infow("order outbox relay: purged published events", "count", n)
	}
}

// orderEnvelope rebuilds the wire message from an outbox row.
func orderEnvelope(row *models.OrderOutbox) contracts.OrderEnvelope {
	return contracts.OrderEnvelope{
		EventID:       row.EventID,
		EventType:     row.EventType,
		SchemaVersion: int(row.SchemaVersion),
		OccurredAt:    row.OccurredAt.Time,
		Payload:       row.Payload,
	}
}
