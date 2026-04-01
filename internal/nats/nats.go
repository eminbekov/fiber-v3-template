package nats

import (
	"context"
	"fmt"
	"log/slog"

	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func Connect(ctx context.Context, natsURL string) (*natsgo.Conn, jetstream.JetStream, error) {
	connection, connectError := natsgo.Connect(
		natsURL,
		natsgo.Name("fiber-v3-template"),
		natsgo.RetryOnFailedConnect(true),
		natsgo.MaxReconnects(-1),
		natsgo.DisconnectErrHandler(func(connection *natsgo.Conn, err error) {
			slog.Warn("nats disconnected", "url", connection.ConnectedUrl(), "error", err)
		}),
		natsgo.ReconnectHandler(func(connection *natsgo.Conn) {
			slog.Info("nats reconnected", "url", connection.ConnectedUrl())
		}),
		natsgo.ClosedHandler(func(connection *natsgo.Conn) {
			slog.Warn("nats connection closed", "last_error", connection.LastError())
		}),
	)
	if connectError != nil {
		return nil, nil, fmt.Errorf("nats connect: %w", connectError)
	}

	jetStream, jetStreamError := jetstream.New(connection)
	if jetStreamError != nil {
		connection.Close()
		return nil, nil, fmt.Errorf("nats jetstream: %w", jetStreamError)
	}

	if ensureError := ensureStreams(ctx, jetStream); ensureError != nil {
		connection.Close()
		return nil, nil, ensureError
	}

	return connection, jetStream, nil
}

func ensureStreams(ctx context.Context, jetStream jetstream.JetStream) error {
	streamConfigurations := []jetstream.StreamConfig{
		{
			Name:     "NOTIFICATIONS",
			Subjects: []string{NotificationSubject},
		},
		{
			Name:     "AUDIT",
			Subjects: []string{AuditLogSubject},
		},
	}

	for _, streamConfiguration := range streamConfigurations {
		_, streamError := jetStream.CreateOrUpdateStream(ctx, streamConfiguration)
		if streamError != nil {
			return fmt.Errorf("nats ensure stream %s: %w", streamConfiguration.Name, streamError)
		}
	}

	return nil
}
