package domain

import (
	"encoding/json"
	"job_scheduler_go_rabbitmq/utils"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusPending  JobStatus = "pending"
	JobStatusQueued   JobStatus = "queued"
	JobStatusRunning  JobStatus = "running"
	JobStatusDead     JobStatus = "dead"
	JobStatudDisabled JobStatus = "disabled"
)

type AttemptStatus string

const (
	AttemptStatusSuccess AttemptStatus = "success"
	AttemptStatusFailed  AttemptStatus = "failed"
)

// Job represents a unit of work to be processed.
type Job struct {
	ID          uuid.UUID       `db:"id" json:"id"`
	Type        string          `db:"type" json:"type"`
	CallbackURL string          `db:"callback_url" json:"callback_url"`
	Payload     json.RawMessage `db:"payload" json:"payload"`
	Status      JobStatus       `db:"status" json:"status"`
	MaxRetries  int             `db:"max_retries" json:"max_retries"`
	ScheduledAt *time.Time      `db:"scheduled_at" json:"scheduled_at"`
	LockedAt    *time.Time      `db:"locked_at" json:"locked_at"`
	LockedBy    *string         `db:"locked_by" json:"locked_by"`
	CompletedAt *time.Time      `db:"completed_at" json:"completed_at"`
	Priority    int             `db:"priority" json:"priority"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

// JobSearchParams defines the parameters for searching jobs.
type JobSearchParams struct {
	ID     *uuid.UUID
	Type   *string
	Status *JobStatus
	utils.SearchParams
}

// JobAttempt represents an attempt to execute a job.
type JobAttempt struct {
	ID            uint64        `db:"id" json:"id"`
	JobID         uuid.UUID     `db:"job_id" json:"job_id"`
	AttemptNumber int           `db:"attempt_number" json:"attempt_number"`
	StartedAt     time.Time     `db:"started_at" json:"started_at"`
	FinishedAt    *time.Time    `db:"finished_at" json:"finished_at"`
	Status        AttemptStatus `db:"status" json:"status"`
	ErrorMessage  *string       `db:"error_message" json:"error_message"`
	HTTPStatus    *int          `db:"http_status" json:"http_status"`
	CreatedAt     time.Time     `db:"created_at" json:"created_at"`
}

// JobAttemptSearchParams defines the parameters for searching job attempts.
type JobAttemptSearchParams struct {
	ID    *uuid.UUID
	JobID *uuid.UUID
	utils.SearchParams
}

// JobEvent represents an event related to a job's lifecycle.
type JobEvent struct {
	ID        uint64          `db:"id" json:"id"`
	JobID     uuid.UUID       `db:"job_id" json:"job_id"`
	EventType string          `db:"event_type" json:"event_type"`
	Message   string          `db:"message" json:"message"`
	Metadata  json.RawMessage `db:"metadata" json:"metadata"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
}

// JobEventSearchParams defines the parameters for searching job events.
type JobEventSearchParams struct {
	ID    *uuid.UUID
	JobID *uuid.UUID
	utils.SearchParams
}

// JobDeadLetter represents a job that has been moved to the dead-letter queue, dead jobs.
type JobDeadLetter struct {
	ID        uint64    `db:"id" json:"id"`
	JobID     uuid.UUID `db:"job_id" json:"job_id"`
	Reason    string    `db:"reason" json:"reason"`
	LastError *string   `db:"last_error" json:"last_error"`
	FailedAt  time.Time `db:"failed_at" json:"failed_at"`
}

type JobDeadLetterSearchParams struct {
	ID    *uuid.UUID
	JobID *uuid.UUID
	utils.SearchParams
}

// JobExecutionRequest represents the payload sent to execute a job.
type JobExecutionRequest struct {
	JobID   uuid.UUID       `json:"job_id"`
	Payload json.RawMessage `json:"payload"`
}

// JobExecutionResponse represents the response received after executing a job.
type JobExecutionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CreateJobInput represents the input required to create a new job.
type CreateJobInput struct {
	Type        string          `json:"type"`
	CallbackURL string          `json:"callback_url"`
	Payload     json.RawMessage `json:"payload"`

	ScheduledAt *time.Time `json:"scheduled_at"`
	MaxRetries  int        `json:"max_retries"`
	Priority    int        `json:"priority"`
}
