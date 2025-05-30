package httpapi

import "time"

type APIKeyRequest struct {
	Key           string `json:"key"`
	Comment       string `json:"comment"`
	HighlyTrusted bool   `json:"highly_trusted"`
}

type APIKeyResponse struct {
	UUID          string    `json:"uuid"`
	Comment       string    `json:"comment"`
	HighlyTrusted bool      `json:"highly_trusted"`
	CreatedAt     time.Time `json:"created_at"`
}
