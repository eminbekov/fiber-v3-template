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

type NotificationConsumer struct {
	jetStream jetstream.JetStream
}

func NewNotificationConsumer(jetStream jetstream.JetStream) *NotificationConsumer {
	return &NotificationConsumer{
		jetStream: jetStream,
	}
}

func (consumer *NotificationConsumer) Run(ctx context.Context) error {
	stream, streamError := consumer.jetStream.Stream(ctx, "NOTIFICATIONS")
	if streamError != nil {
		return fmt.Errorf("notification stream: %w", streamError)
	}

	jetStreamConsumer, consumerError := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:          "notification-consumer",
		FilterSubject: appnats.NotificationSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		MaxDeliver:    10,
	})
	if consumerError != nil {
		return fmt.Errorf("notification consumer: %w", consumerError)
	}

	consumeContext, consumeError := jetStreamConsumer.Consume(consumer.handle)
	if consumeError != nil {
		return fmt.Errorf("notification consume: %w", consumeError)
	}
	defer consumeContext.Stop()

	<-ctx.Done()
	return nil
}

func (consumer *NotificationConsumer) handle(message jetstream.Msg) {
	var event appnats.NotificationEvent
	if unmarshalError := json.Unmarshal(message.Data(), &event); unmarshalError != nil {
		slog.Error("invalid notification event payload", "error", unmarshalError)
		if nakError := message.Nak(); nakError != nil {
			slog.Error("nats notification message nak failed", "error", nakError)
		}
		return
	}

	// Placeholder for real side effects (email, telegram, push notifications).
	slog.Info("notification event consumed", "chat_id", event.ChatID)
	if ackError := message.Ack(); ackError != nil {
		slog.Error("nats notification message ack failed", "error", ackError)
	}
}
