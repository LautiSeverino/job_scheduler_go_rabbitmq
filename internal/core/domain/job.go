package domain

import (
	"encoding/json"
	"job_scheduler_go_rabbitmq/utils"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"   // creado, aún no encolado
	JobStatusQueued    JobStatus = "queued"    // enviado a RabbitMQ
	JobStatusRunning   JobStatus = "running"   // worker ejecutando
	JobStatusCompleted JobStatus = "completed" // ejecución exitosa
	JobStatusFailed    JobStatus = "failed"    // falló, pero puede retry
	JobStatusDead      JobStatus = "dead"      // sin retries, DLQ
	JobStatusDisabled  JobStatus = "disabled"  // cancelado manualmente
)

type AttemptStatus string

const (
	AttemptStatusSuccess AttemptStatus = "success"
	AttemptStatusFailed  AttemptStatus = "failed"
)

type EventType string

const (
	EventJobCreated   EventType = "job_created"
	EventJobQueued    EventType = "job_queued"
	EventJobRunning   EventType = "job_running"
	EventJobSucceeded EventType = "job_succeeded"
	EventJobFailed    EventType = "job_failed"
	EventJobDead      EventType = "job_dead"
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

	// Scheduler-specific
	ReadyToRun  *bool
	LockFree    *bool
	LockTimeout *time.Duration

	utils.SearchParams
}

// Attempt represents an attempt to execute a job.
type Attempt struct {
	ID            uuid.UUID     `db:"id" json:"id"`
	JobID         uuid.UUID     `db:"job_id" json:"job_id"`
	AttemptNumber int           `db:"attempt_number" json:"attempt_number"`
	StartedAt     time.Time     `db:"started_at" json:"started_at"`
	Status        AttemptStatus `db:"status" json:"status"`
	ErrorMessage  *string       `db:"error_message" json:"error_message"`
	HTTPStatus    *int          `db:"http_status" json:"http_status"`
	CreatedAt     time.Time     `db:"created_at" json:"created_at"`
}

// AttemptSearchParams defines the parameters for searching job attempts.
type AttemptSearchParams struct {
	ID    *uuid.UUID
	JobID *uuid.UUID
	utils.SearchParams
}

// Event represents an event related to a job's lifecycle.
type Event struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	JobID     uuid.UUID       `db:"job_id" json:"job_id"`
	Type      EventType       `db:"event_type" json:"event_type"`
	Message   string          `db:"message" json:"message"`
	Metadata  json.RawMessage `db:"metadata" json:"metadata"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
}

// EventSearchParams defines the parameters for searching job events.
type EventSearchParams struct {
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
	ScheduledAt *time.Time      `json:"scheduled_at"`
	MaxRetries  int             `json:"max_retries"`
	Priority    int             `json:"priority"`
}

type ExecutionResult struct {
	HTTPStatus int
	Error      error
}

func NewJob(input CreateJobInput) (*Job, error) {
	job := &Job{
		ID:          uuid.New(),
		Type:        input.Type,
		CallbackURL: input.CallbackURL,
		Payload:     input.Payload,
		MaxRetries:  input.MaxRetries,
		Priority:    input.Priority,
		Status:      JobStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if input.ScheduledAt != nil {
		job.ScheduledAt = input.ScheduledAt
	} else {
		job.ScheduledAt = nil
	}

	return job, nil
}

func NewJobSucceededEvent(jobID uuid.UUID) Event {
	return Event{
		ID:        uuid.New(),
		JobID:     jobID,
		Type:      EventJobSucceeded,
		Message:   "job completed successfully",
		CreatedAt: time.Now(),
	}
}

func NewJobFailedEvent(jobID uuid.UUID, errMsg string) Event {
	return Event{
		ID:        uuid.New(),
		JobID:     jobID,
		Type:      EventJobFailed,
		Message:   errMsg,
		CreatedAt: time.Now(),
	}
}

func NewJobDeadEvent(jobID uuid.UUID) Event {
	return Event{
		ID:        uuid.New(),
		JobID:     jobID,
		Type:      EventJobDead,
		Message:   "job moved to dead letter queue",
		CreatedAt: time.Now(),
	}
}

func NewAttempt(jobID uuid.UUID, attemptNumber int, status AttemptStatus, errMsg *string, httpStatus *int) Attempt {
	return Attempt{
		ID:            uuid.New(),
		JobID:         jobID,
		AttemptNumber: attemptNumber,
		Status:        status,
		ErrorMessage:  errMsg,
		HTTPStatus:    httpStatus,
		CreatedAt:     time.Now(),
	}
}
