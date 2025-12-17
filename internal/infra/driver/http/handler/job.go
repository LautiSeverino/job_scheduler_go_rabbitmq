package handler

import (
	"encoding/json"
	"job_scheduler_go_rabbitmq/internal/core/domain"
	"job_scheduler_go_rabbitmq/internal/core/ports"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type JobHandler struct {
	service ports.IJobService
}

func NewJobHandler(service ports.IJobService) *JobHandler {
	return &JobHandler{service: service}
}

func RegisterJobRoutes(r *mux.Router, handler *JobHandler) {
	r.HandleFunc("/jobs", handler.Create()).Methods(http.MethodPost)                   // POST para crear un job
	r.HandleFunc("/jobs/{id}", handler.GetOne()).Methods(http.MethodGet)               // GET para obtener un job por ID
	r.HandleFunc("/jobs/{id}/timeline", handler.GetTimeline()).Methods(http.MethodGet) // GET para el timeline de un job
}

// Create implements ports.IJobHandler.
func (j *JobHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse body para obtener input de creación
		var input domain.CreateJobInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Crear job usando el servicio
		job, err := j.service.Create(r.Context(), input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Responder con el job creado
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(job)
	}
}

// GetOne implements ports.IJobHandler.
func (j *JobHandler) GetOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtener el ID del job de los parámetros de la URL
		vars := mux.Vars(r)
		jobID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "Invalid job ID", http.StatusBadRequest)
			return
		}

		// Obtener el job desde el servicio
		job, err := j.service.GetOne(r.Context(), domain.JobSearchParams{ID: &jobID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Responder con el job encontrado
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(job)
	}
}

// GetTimeline implements ports.IJobHandler.
func (j *JobHandler) GetTimeline() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtener el ID del job de los parámetros de la URL
		vars := mux.Vars(r)
		jobID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "Invalid job ID", http.StatusBadRequest)
			return
		}

		// Obtener el timeline del job
		timeline, err := j.service.GetTimeline(r.Context(), jobID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Responder con el timeline
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(timeline)
	}
}
