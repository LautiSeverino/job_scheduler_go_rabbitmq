package handler

import (
	"job_scheduler_go_rabbitmq/internal/core/ports"
	"net/http"

	"github.com/gorilla/mux"
)

type JobHandler struct {
	service ports.IJobService
}

func NewJobHandler(service ports.IJobService) ports.IJobHandler {
	return &JobHandler{service: service}
}

// Create implements ports.IJobHandler.
func (j *JobHandler) Create() http.HandlerFunc {
	panic("unimplemented")
}

// EnqueueDueJobs implements ports.IJobHandler.
func (j *JobHandler) EnqueueDueJobs() http.HandlerFunc {
	panic("unimplemented")
}

// GetJobTimeline implements ports.IJobHandler.
func (j *JobHandler) GetJobTimeline() http.HandlerFunc {
	panic("unimplemented")
}

// GetOne implements ports.IJobHandler.
func (j *JobHandler) GetOne() http.HandlerFunc {
	panic("unimplemented")
}

// RegisterPrivateRoutes implements ports.IJobHandler.
func (j *JobHandler) RegisterPrivateRoutes(router *mux.Router) {
	panic("unimplemented")
}

// RegisterPublicRoutes implements ports.IJobHandler.
func (j *JobHandler) RegisterPublicRoutes(router *mux.Router) {
	panic("unimplemented")
}

// RetryJob implements ports.IJobHandler.
func (j *JobHandler) RetryJob() http.HandlerFunc {
	panic("unimplemented")
}
