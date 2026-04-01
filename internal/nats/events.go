package nats

import (
	"time"
)

const (
	NotificationSubject = "notifications.send"
	AuditLogSubject     = "audit.log"
)

type NotificationEvent struct {
	ChatID    int64     `json:"chat_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type AuditLogEvent struct {
	ActorID    string    `json:"actor_id"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id"`
	CreatedAt  time.Time `json:"created_at"`
}
