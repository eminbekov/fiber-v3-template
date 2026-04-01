package websocket

import (
	"context"
	"fmt"
	"strings"

	fiberwebsocket "github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

const DefaultChannel = "global"

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

func RequireUpgrade(ctx fiber.Ctx) error {
	if !fiberwebsocket.IsWebSocketUpgrade(ctx) {
		return fiber.ErrUpgradeRequired
	}
	if nextError := ctx.Next(); nextError != nil {
		return fmt.Errorf("RequireUpgrade: %w", nextError)
	}
	return nil
}

func (handler *Handler) HandleConnection(connection *fiberwebsocket.Conn) {
	channel := strings.TrimSpace(connection.Query("channel", DefaultChannel))
	if channel == "" {
		channel = DefaultChannel
	}

	handler.hub.Register(connection, channel)
	defer handler.hub.Unregister(connection)

	for {
		messageType, messagePayload, readError := connection.ReadMessage()
		if readError != nil {
			return
		}
		if messageType != fiberwebsocket.TextMessage && messageType != fiberwebsocket.BinaryMessage {
			continue
		}
		_ = handler.hub.Broadcast(context.Background(), channel, messagePayload)
	}
}
