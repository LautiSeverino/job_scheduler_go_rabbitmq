package domain

import (
	"encoding/json"

	"github.com/google/uuid"
)

type RabbitJobMessage struct {
	JobID       uuid.UUID       `json:"job_id"`
	Type        string          `json:"type"`
	CallbackURL string          `json:"callback_url"`
	Payload     json.RawMessage `json:"payload"`
	Attempt     int             `json:"attempt"`
}
