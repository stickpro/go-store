package mail

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/mailer"
	"github.com/stickpro/go-store/pkg/queue"
)

type captureMailer struct {
	last mailer.Message
	err  error
}

func (c *captureMailer) Send(_ context.Context, msg mailer.Message) error {
	if c.err != nil {
		return c.err
	}
	c.last = msg
	return nil
}

func newTestService(t *testing.T) (*Service, *captureMailer, queue.IQueue) {
	t.Helper()

	conf := &config.Config{Email: config.EmailConfig{
		FromName: "go-store",
		BaseURL:  "https://shop.example.com",
	}}

	q := queue.NewInMemoryQueue()
	svc, err := New(conf, logger.New(), q)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	mc := &captureMailer{}
	svc.mailer = mc
	// run scheduled retries immediately
	svc.afterFunc = func(_ time.Duration, f func()) *time.Timer {
		f()
		return nil
	}
	return svc, mc, q
}

func TestEnqueueThenDeliver(t *testing.T) {
	svc, mc, q := newTestService(t)
	ctx := context.Background()

	if err := svc.Enqueue(ctx, "user@example.com", Welcome{Email: "user@example.com"}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	payload, err := q.Pop(ctx, QueueName)
	if err != nil || payload == nil {
		t.Fatalf("queue Pop: payload=%v err=%v", payload, err)
	}

	var task Task
	if err := json.Unmarshal(payload, &task); err != nil {
		t.Fatalf("unmarshal task: %v", err)
	}
	if task.Kind != KindWelcome || task.To != "user@example.com" {
		t.Fatalf("unexpected task: %+v", task)
	}

	if err := svc.Deliver(ctx, task); err != nil {
		t.Fatalf("Deliver: %v", err)
	}

	if got := mc.last.To; len(got) != 1 || got[0] != "user@example.com" {
		t.Fatalf("recipient = %v", got)
	}
	if mc.last.Subject != "Добро пожаловать в go-store" {
		t.Errorf("subject = %q", mc.last.Subject)
	}
	for _, want := range []string{"user@example.com", "https://shop.example.com", "<!doctype html>"} {
		if !strings.Contains(mc.last.HTML, want) {
			t.Errorf("body missing %q:\n%s", want, mc.last.HTML)
		}
	}
}

func TestDeliverUnknownKind(t *testing.T) {
	svc, _, _ := newTestService(t)

	err := svc.Deliver(context.Background(), Task{Kind: "bogus", To: "x@example.com", Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("expected error for unknown task kind")
	}
}

func TestHandleFailureRetryThenDeadLetter(t *testing.T) {
	svc, _, q := newTestService(t)

	// Attempt just below the limit -> retry is scheduled on QueueName.
	svc.HandleFailure(Task{Kind: KindWelcome, To: "a@b.c", Attempt: maxAttempts - 2}, errors.New("smtp down"))
	waitForQueue(t, q, QueueName)

	// Last attempt -> dead-letter.
	svc.HandleFailure(Task{Kind: KindWelcome, To: "a@b.c", Attempt: maxAttempts - 1}, errors.New("smtp down"))
	dead := waitForQueue(t, q, DeadQueueName)

	var task Task
	if err := json.Unmarshal(dead, &task); err != nil {
		t.Fatalf("unmarshal dead task: %v", err)
	}
	if task.Attempt != maxAttempts {
		t.Errorf("dead task attempt = %d, want %d", task.Attempt, maxAttempts)
	}
}

// waitForQueue pops with a short deadline; HandleFailure schedules the retry
// push via time.AfterFunc so it is not instant.
func waitForQueue(t *testing.T, q queue.IQueue, name string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload, err := q.Pop(ctx, name)
	if err != nil || payload == nil {
		t.Fatalf("queue %s: no message (err=%v)", name, err)
	}
	return payload
}
