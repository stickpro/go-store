// Package mail accepts transactional email requests, queues them, and (in the
// mail worker) renders and delivers them over SMTP.
//
// Callers use the typed Message structs from message.go plus Enqueue; they never
// wait on rendering or SMTP. The worker calls Deliver, and HandleFailure on
// error for retry / dead-lettering.
package mail

import (
	"context"
	"fmt"
	"time"

	"github.com/goccy/go-json"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/mailer"
	"github.com/stickpro/go-store/pkg/queue"
)

// IMailService is the enqueue-side API plus the two hooks the worker needs.
type IMailService interface {
	// Enqueue schedules a message for delivery and returns as soon as it is on
	// the queue.
	Enqueue(ctx context.Context, to string, msg Message) error
	// Deliver renders and sends a task pulled from the queue. Worker only.
	Deliver(ctx context.Context, task Task) error
	// HandleFailure re-queues a failed task with backoff or dead-letters it.
	// Worker only.
	HandleFailure(task Task, cause error)
}

type Service struct {
	logger   logger.Logger
	mailer   mailer.Mailer
	queue    queue.IQueue
	renderer *renderer

	// afterFunc schedules delayed retries; overridable in tests. Defaults to
	// time.AfterFunc.
	afterFunc func(time.Duration, func()) *time.Timer
}

func New(conf *config.Config, log logger.Logger, q queue.IQueue) (*Service, error) {
	r, err := newRenderer(conf.Email.FromName, conf.Email.BaseURL)
	if err != nil {
		return nil, err
	}

	var m mailer.Mailer = mailer.Noop{}
	if conf.Email.Enabled {
		m = mailer.NewSMTP(mailer.Config{
			Host:     conf.Email.Host,
			Port:     conf.Email.Port,
			Username: conf.Email.Username,
			Password: conf.Email.Password,
			From:     conf.Email.From,
			FromName: conf.Email.FromName,
		})
	} else {
		log.Info("mail: delivery disabled (email.enabled=false), messages will be dropped")
	}

	return &Service{
		logger:    log,
		mailer:    m,
		queue:     q,
		renderer:  r,
		afterFunc: time.AfterFunc,
	}, nil
}

func (s *Service) Enqueue(ctx context.Context, to string, msg Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("mail: marshal %s payload: %w", msg.Kind(), err)
	}
	return s.push(ctx, Task{Kind: msg.Kind(), To: to, Payload: payload})
}

func (s *Service) Deliver(ctx context.Context, task Task) error {
	newMsg, ok := registry[task.Kind]
	if !ok {
		return fmt.Errorf("mail: unknown task kind %q", task.Kind)
	}

	msg := newMsg()
	if err := json.Unmarshal(task.Payload, msg); err != nil {
		return fmt.Errorf("mail: unmarshal %s payload: %w", task.Kind, err)
	}

	subject, body, err := s.renderer.render(msg)
	if err != nil {
		return err
	}

	return s.mailer.Send(ctx, mailer.Message{
		To:      []string{task.To},
		Subject: subject,
		HTML:    body,
	})
}

func (s *Service) push(ctx context.Context, task Task) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("mail: marshal %s task: %w", task.Kind, err)
	}
	if err := s.queue.Push(ctx, QueueName, payload); err != nil {
		return fmt.Errorf("mail: enqueue %s task: %w", task.Kind, err)
	}
	return nil
}
