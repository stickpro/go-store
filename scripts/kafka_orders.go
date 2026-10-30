//go:build ignore

// Reads order events from Kafka (store.orders) and prints them.
// No consumer group is used, so offsets are not committed and nobody else's
// consumption is affected.
//
//	go run scripts/kafka_orders.go                      # print everything and exit
//	go run scripts/kafka_orders.go -f                   # keep waiting for new events
//	go run scripts/kafka_orders.go -raw                 # print raw JSON
//	go run scripts/kafka_orders.go -brokers localhost:19092
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/messaging/topics"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	brokers := flag.String("brokers", "46.148.236.163:9092", "comma-separated broker list")
	topic := flag.String("topic", topics.Orders, "topic to read")
	follow := flag.Bool("f", false, "keep waiting for new events instead of exiting")
	raw := flag.Bool("raw", false, "print raw message JSON")
	idle := flag.Duration("idle", 5*time.Second, "exit after this long without new messages (ignored with -f)")
	flag.Parse()

	client, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.ConsumeTopics(*topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		log.Fatalf("kafka client: %v", err)
	}
	defer client.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	count := 0
	for {
		pollCtx := ctx
		var cancel context.CancelFunc = func() {}
		if !*follow {
			pollCtx, cancel = context.WithTimeout(ctx, *idle)
		}
		fetches := client.PollFetches(pollCtx)
		cancel()

		if ctx.Err() != nil || (!*follow && errors.Is(pollCtx.Err(), context.DeadlineExceeded) && fetches.NumRecords() == 0) {
			break
		}
		fetches.EachError(func(t string, p int32, err error) {
			if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				log.Printf("fetch error %s[%d]: %v", t, p, err)
			}
		})
		fetches.EachRecord(func(r *kgo.Record) {
			count++
			if *raw {
				fmt.Printf("p%d@%d key=%s %s\n", r.Partition, r.Offset, r.Key, r.Value)
				return
			}
			printRecord(r)
		})
	}

	fmt.Printf("\ntotal events: %d\n", count)
}

func printRecord(r *kgo.Record) {
	var env contracts.OrderEnvelope
	if err := json.Unmarshal(r.Value, &env); err != nil {
		fmt.Printf("p%d@%d: bad envelope: %v\n%s\n\n", r.Partition, r.Offset, err, r.Value)
		return
	}

	fmt.Printf("=== %s  %s  (p%d@%d, schema v%d, event %s)\n",
		env.OccurredAt.Local().Format("2006-01-02 15:04:05"), env.EventType,
		r.Partition, r.Offset, env.SchemaVersion, env.EventID)

	switch env.EventType {
	case contracts.OrderEventCreated, contracts.OrderEventUpdated:
		var o contracts.OrderSnapshotPayload
		if err := json.Unmarshal(env.Payload, &o); err != nil {
			fmt.Printf("  bad payload: %v\n\n", err)
			return
		}
		fmt.Printf("  order #%d  id=%s  version=%d  source=%s\n", o.OrderNumber, o.OrderID, o.Version, o.Source)
		fmt.Printf("  status=%s  payment=%s/%s\n", o.Status, o.PaymentStatus, deref(o.PaymentMethod))
		fmt.Printf("  customer: %s  %s\n", o.Email, deref(o.Phone))
		fmt.Printf("  shipping: %s, %s, %s (%s %s)\n",
			o.Shipping.Recipient, o.Shipping.CityName, o.Shipping.Address,
			deref(o.Shipping.Provider), deref(o.Shipping.Method))
		for _, it := range o.Items {
			fmt.Printf("    - %s [%s] %s x %d = %s\n", it.Name, deref(it.Sku), it.UnitPrice, it.Quantity, it.LineTotal)
		}
		fmt.Printf("  subtotal=%s discount=%s shipping=%s tax=%s  TOTAL=%s %s\n",
			o.Subtotal, o.DiscountTotal, o.ShippingTotal, o.TaxTotal, o.GrandTotal, o.Currency)
		if o.Comment != nil {
			fmt.Printf("  comment: %s\n", *o.Comment)
		}
	case contracts.OrderEventPaid:
		var p contracts.OrderPaidPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			fmt.Printf("  bad payload: %v\n\n", err)
			return
		}
		fmt.Printf("  order id=%s paid at %s via %s (provider_payment_id=%s)\n",
			p.OrderID, p.PaidAt.Local().Format(time.DateTime), deref(p.PaymentMethod), deref(p.ProviderPaymentID))
	case contracts.OrderEventCancelled:
		var p contracts.OrderCancelledPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			fmt.Printf("  bad payload: %v\n\n", err)
			return
		}
		fmt.Printf("  order id=%s cancelled at %s: %s\n", p.OrderID, p.CancelledAt.Local().Format(time.DateTime), p.Reason)
	default:
		fmt.Printf("  payload: %s\n", env.Payload)
	}
	fmt.Println()
}

func deref(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}
