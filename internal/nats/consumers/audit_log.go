package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	appnats "github.com/eminbekov/fiber-v3-template/internal/nats"
	"github.com/nats-io/nats.go/jetstream"
)

type AuditLogConsumer struct {
	jetStream jetstream.JetStream
}

func NewAuditLogConsumer(jetStream jetstream.JetStream) *AuditLogConsumer {
	return &AuditLogConsumer{
		jetStream: jetStream,
	}
}

func (consumer *AuditLogConsumer) Run(ctx context.Context) error {
	stream, streamError := consumer.jetStream.Stream(ctx, "AUDIT")
	if streamError != nil {
		return fmt.Errorf("audit stream: %w", streamError)
	}

	jetStreamConsumer, consumerError := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:          "audit-log-consumer",
		FilterSubject: appnats.AuditLogSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		MaxDeliver:    10,
	})
	if consumerError != nil {
		return fmt.Errorf("audit consumer: %w", consumerError)
	}

	consumeContext, consumeError := jetStreamConsumer.Consume(consumer.handle)
	if consumeError != nil {
		return fmt.Errorf("audit consume: %w", consumeError)
	}
	defer consumeContext.Stop()

	<-ctx.Done()
	return nil
}

func (consumer *AuditLogConsumer) handle(message jetstream.Msg) {
	var event appnats.AuditLogEvent
	if unmarshalError := json.Unmarshal(message.Data(), &event); unmarshalError != nil {
		slog.Error("invalid audit event payload", "error", unmarshalError)
		if nakError := message.Nak(); nakError != nil {
			slog.Error("nats audit message nak failed", "error", nakError)
		}
		return
	}

	slog.Info(
		"audit event consumed",
		"actor_id",
		event.ActorID,
		"action",
		event.Action,
		"resource",
		event.Resource,
		"resource_id",
		event.ResourceID,
	)
	if ackError := message.Ack(); ackError != nil {
		slog.Error("nats audit message ack failed", "error", ackError)
	}
}
