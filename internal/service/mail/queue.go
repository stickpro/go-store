package mail

import (
	"context"
	"time"

	"github.com/goccy/go-json"
)

const (
	// QueueName holds tasks waiting for their first (or a retried) delivery.
	QueueName = "mail_send"
	// DeadQueueName parks tasks that exhausted their retries, for manual
	// inspection / replay.
	DeadQueueName = "mail_dead"

	maxAttempts = 5
	// enqueueTimeout bounds the background re-queue / dead-letter push.
	enqueueTimeout = 10 * time.Second
)

// Task is the queue payload. Payload is the JSON of the concrete Message.
type Task struct {
	Kind    string          `json:"kind"`
	To      string          `json:"to"`
	Payload json.RawMessage `json:"payload"`
	Attempt int             `json:"attempt,omitempty"`
}

// HandleFailure is called by the worker after Deliver fails. It bumps the
// attempt counter and either schedules a retry with exponential backoff or moves
// the task to the dead-letter queue.
//
// The retry is scheduled with time.AfterFunc, so a process crash during the
// backoff window loses that pending retry — acceptable for the current tier;
// swap in a delayed queue if that stops being true.
func (s *Service) HandleFailure(task Task, cause error) {
	task.Attempt++

	if task.Attempt >= maxAttempts {
		s.logger.Errorw("mail: delivery failed permanently, moving to dead-letter",
			"kind", task.Kind, "to", task.To, "attempts", task.Attempt, "error", cause)
		s.repush(DeadQueueName, task)
		return
	}

	delay := backoff(task.Attempt)
	s.logger.Warnw("mail: delivery failed, scheduling retry",
		"kind", task.Kind, "to", task.To, "attempt", task.Attempt, "retry_in", delay.String(), "error", cause)

	s.afterFunc(delay, func() { s.repush(QueueName, task) })
}

func (s *Service) repush(queueName string, task Task) {
	ctx, cancel := context.WithTimeout(context.Background(), enqueueTimeout)
	defer cancel()

	payload, err := json.Marshal(task)
	if err != nil {
		s.logger.Errorw("mail: marshal task for re-enqueue", "kind", task.Kind, "to", task.To, "error", err)
		return
	}
	if err := s.queue.Push(ctx, queueName, payload); err != nil {
		s.logger.Errorw("mail: re-enqueue failed", "queue", queueName, "kind", task.Kind, "to", task.To, "error", err)
	}
}

// backoff returns the delay before retry N (1-indexed): 1m, 2m, 4m, 8m, capped.
func backoff(attempt int) time.Duration {
	const base = time.Minute
	const maxDelay = 30 * time.Minute

	d := base << (attempt - 1)
	if d <= 0 || d > maxDelay {
		return maxDelay
	}
	return d
}
