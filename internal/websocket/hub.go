package websocket

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	fiberwebsocket "github.com/gofiber/contrib/v3/websocket"
	"github.com/redis/go-redis/v9"
)

type trackedConnection struct {
	connection *fiberwebsocket.Conn
	channel    string
	writeMutex sync.Mutex
}

type Hub struct {
	redisClient *redis.Client
	connections sync.Map
}

func NewHub(redisClient *redis.Client) *Hub {
	return &Hub{
		redisClient: redisClient,
	}
}

func (hub *Hub) Register(connection *fiberwebsocket.Conn, channel string) {
	hub.connections.Store(connection, &trackedConnection{
		connection: connection,
		channel:    channel,
	})
}

func (hub *Hub) Unregister(connection *fiberwebsocket.Conn) {
	hub.connections.Delete(connection)
	_ = connection.Close()
}

func (hub *Hub) Broadcast(ctx context.Context, channel string, message []byte) error {
	channelName := "ws:" + channel
	if publishError := hub.redisClient.Publish(ctx, channelName, message).Err(); publishError != nil {
		return fmt.Errorf("websocketHub.Broadcast: %w", publishError)
	}
	return nil
}

func (hub *Hub) Subscribe(ctx context.Context, channel string) error {
	channelName := "ws:" + channel
	subscriber := hub.redisClient.Subscribe(ctx, channelName)
	defer func() {
		if closeError := subscriber.Close(); closeError != nil {
			slog.Error("websocket subscriber close", "error", closeError)
		}
	}()

	messageChannel := subscriber.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case message, isOpen := <-messageChannel:
			if !isOpen {
				return nil
			}
			hub.forwardToLocalConnections(channel, []byte(message.Payload))
		}
	}
}

func (hub *Hub) forwardToLocalConnections(channel string, payload []byte) {
	hub.connections.Range(func(_, value any) bool {
		connection, isTrackedConnection := value.(*trackedConnection)
		if !isTrackedConnection || connection.channel != channel {
			return true
		}
		connection.writeMutex.Lock()
		writeError := connection.connection.WriteMessage(fiberwebsocket.TextMessage, payload)
		connection.writeMutex.Unlock()
		if writeError != nil {
			slog.Warn("websocket write failed", "error", writeError)
		}
		return true
	})
}
