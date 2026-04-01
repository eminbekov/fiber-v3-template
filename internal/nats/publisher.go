package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

type Publisher struct {
	jetStream jetstream.JetStream
}

func NewPublisher(jetStream jetstream.JetStream) *Publisher {
	return &Publisher{
		jetStream: jetStream,
	}
}

func (publisher *Publisher) Publish(ctx context.Context, subject string, payload []byte) error {
	if _, publishError := publisher.jetStream.Publish(ctx, subject, payload); publishError != nil {
		return fmt.Errorf("publish: %w", publishError)
	}

	return nil
}

func (publisher *Publisher) PublishNotification(ctx context.Context, event NotificationEvent) error {
	payload, marshalError := json.Marshal(event)
	if marshalError != nil {
		return fmt.Errorf("marshal notification event: %w", marshalError)
	}

	if publishError := publisher.Publish(ctx, NotificationSubject, payload); publishError != nil {
		return fmt.Errorf("publish notification event: %w", publishError)
	}

	return nil
}

func (publisher *Publisher) PublishAuditLog(ctx context.Context, event AuditLogEvent) error {
	payload, marshalError := json.Marshal(event)
	if marshalError != nil {
		return fmt.Errorf("marshal audit log event: %w", marshalError)
	}

	if publishError := publisher.Publish(ctx, AuditLogSubject, payload); publishError != nil {
		return fmt.Errorf("publish audit log event: %w", publishError)
	}

	return nil
}
