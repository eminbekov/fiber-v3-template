package v1

import "time"

type LoginResponse struct {
	Token     string    `json:"token" example:"session_01HQW1K4N2A9D7YZ8R3B5M6C7E"`
	ExpiresAt time.Time `json:"expires_at" example:"2026-04-02T10:00:00Z"`
}
