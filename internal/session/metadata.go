package session

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Metadata struct {
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}
