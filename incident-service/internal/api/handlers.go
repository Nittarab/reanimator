package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/your-org/ai-sre-platform/incident-service/internal/adapters"
	"github.com/your-org/ai-sre-platform/incident-service/internal/config"
	"github.com/your-org/ai-sre-platform/incident-service/internal/database"
	"github.com/your-org/ai-sre-platform/incident-service/internal/github"
	"github.com/your-org/ai-sre-platform/incident-service/internal/models"
)

// Server represents the HTTP server
type Server struct {
	config       *config.Config
	db           *database.DB
	redis        *database.RedisClient
	repository   *database.IncidentRepository
	adapters     *adapters.Registry
	githubClient *github.Client
	logger       *Logger
	metrics      *Metrics
	router       *chi.Mux
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, db *database.DB, redis *database.RedisClient, githubClient *github.Client) *Server {
	s := &Server{
		config:       cfg,
		db:           db,
		redis:        redis,
		repository:   database.NewIncidentRepository(db),
		adapters:     adapters.NewRegistry(),
		githubClient: githubClient,
		logger:       NewLogger(),
		metrics:      NewMetrics(),
		router:       chi.NewRouter(),
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.Get("/api/v1/health", s.handleHealth)

	// Metrics endpoint
	s.router.Handle("/api/v1/metrics", promhttp.Handler())

	// Webhook endpoint
	s.router.Post("/api/v1/webhooks/incidents", s.handleWebhook)

	// Incident endpoints (to be implemented in later tasks)
	s.router.Get("/api/v1/incidents", s.handleListIncidents)
	s.router.Get("/api/v1/incidents/{id}", s.handleGetIncident)

	// Workflow status webhook endpoint
	s.router.Post("/api/v1/webhooks/workflow-status", s.handleWorkflowStatus)

	// Configuration endpoint
	s.router.Get("/api/v1/config", s.handleGetConfig)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Check database health
	if err := s.db.Health(); err != nil {
		s.logger.Error("database health check failed", map[string]interface{}{
			"error": err.Error(),
		})
		health["status"] = "unhealthy"
		health["database"] = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		health["database"] = "healthy"
	}

	// Check Redis health
	if err := s.redis.Health(ctx); err != nil {
		s.logger.Error("redis health check failed", map[string]interface{}{
			"error": err.Error(),
		})
		health["status"] = "unhealthy"
		health["redis"] = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		health["redis"] = "healthy"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleListIncidents handles listing incidents (placeholder)
func (s *Server) handleListIncidents(w http.ResponseWriter, r *http.Request) {
	incidents, err := s.repository.List()
	if err != nil {
		s.logger.Error("failed to list incidents", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Return response in the format expected by the dashboard
	response := map[string]interface{}{
		"incidents": incidents,
		"total":     len(incidents),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetIncident handles getting a single incident (placeholder)
func (s *Server) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	incident, err := s.repository.GetByID(id)
	if err != nil {
		s.logger.Error("failed to get incident", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		http.Error(w, "incident not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incident)
}

// Router returns the HTTP router
func (s *Server) Router() *chi.Mux {
	return s.router
}

// Logger returns the logger
func (s *Server) Logger() *Logger {
	return s.logger
}

// handleWebhook handles incoming webhook requests from observability platforms
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get provider from query parameter
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		s.logger.Error("missing provider parameter", nil)
		http.Error(w, "missing provider parameter", http.StatusBadRequest)
		s.metrics.IncidentReceived.WithLabelValues(provider, "error").Inc()
		return
	}

	// Get adapter for provider
	adapter, ok := s.adapters.Get(provider)
	if !ok {
		s.logger.Error("unsupported provider", map[string]interface{}{
			"provider": provider,
		})
		http.Error(w, "unsupported provider", http.StatusBadRequest)
		s.metrics.IncidentReceived.WithLabelValues(provider, "error").Inc()
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("failed to read request body", map[string]interface{}{
			"error":    err.Error(),
			"provider": provider,
		})
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		s.metrics.IncidentReceived.WithLabelValues(provider, "error").Inc()
		return
	}

	// Create a new reader for validation (since we consumed the body)
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Validate webhook
	if err := adapter.Validate(r); err != nil {
		s.logger.Error("webhook validation failed", map[string]interface{}{
			"error":    err.Error(),
			"provider": provider,
		})
		http.Error(w, "validation failed", http.StatusUnauthorized)
		s.metrics.IncidentReceived.WithLabelValues(provider, "validation_failed").Inc()
		return
	}

	// Parse incident
	incident, err := adapter.Parse(body)
	if err != nil {
		s.logger.Error("failed to parse webhook payload", map[string]interface{}{
			"error":    err.Error(),
			"provider": provider,
		})
		http.Error(w, "failed to parse payload", http.StatusBadRequest)
		s.metrics.IncidentReceived.WithLabelValues(provider, "parse_error").Inc()
		return
	}

	// Store incident
	if err := s.repository.Create(incident); err != nil {
		s.logger.Error("failed to store incident", map[string]interface{}{
			"error":       err.Error(),
			"provider":    provider,
			"incident_id": incident.ID,
		})
		http.Error(w, "internal server error", http.StatusInternalServerError)
		s.metrics.IncidentReceived.WithLabelValues(provider, "storage_error").Inc()
		return
	}

	// Log success
	s.logger.Info("incident received and stored", map[string]interface{}{
		"incident_id":  incident.ID,
		"provider":     provider,
		"service_name": incident.ServiceName,
		"severity":     incident.Severity,
		"duration_ms":  time.Since(startTime).Milliseconds(),
	})

	// Update metrics
	s.metrics.IncidentReceived.WithLabelValues(provider, "success").Inc()
	s.metrics.WebhookProcessingDuration.WithLabelValues(provider).Observe(time.Since(startTime).Seconds())

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "accepted",
		"incident_id": incident.ID,
	})
}

// WorkflowStatusPayload represents the payload from GitHub Actions workflow completion
type WorkflowStatusPayload struct {
	IncidentID     string `json:"incident_id"`
	Status         string `json:"status"` // "success", "failed", "no_fix_needed"
	PullRequestURL string `json:"pr_url,omitempty"`
	Diagnosis      string `json:"diagnosis,omitempty"`
	Repository     string `json:"repository"`
}

// handleWorkflowStatus handles workflow completion webhooks from GitHub Actions
func (s *Server) handleWorkflowStatus(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse request body
	var payload WorkflowStatusPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.logger.Error("failed to parse workflow status payload", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if payload.IncidentID == "" || payload.Status == "" || payload.Repository == "" {
		s.logger.Error("missing required fields in workflow status payload", map[string]interface{}{
			"incident_id": payload.IncidentID,
			"status":      payload.Status,
			"repository":  payload.Repository,
		})
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	// Get the incident
	incident, err := s.repository.GetByID(payload.IncidentID)
	if err != nil {
		s.logger.Error("failed to get incident for workflow status update", map[string]interface{}{
			"error":       err.Error(),
			"incident_id": payload.IncidentID,
		})
		http.Error(w, "incident not found", http.StatusNotFound)
		return
	}

	// Update incident based on workflow status
	now := time.Now()
	incident.CompletedAt = &now

	switch payload.Status {
	case "success":
		if payload.PullRequestURL != "" {
			incident.Status = models.StatusPRCreated
			incident.PullRequestURL = &payload.PullRequestURL
		} else {
			incident.Status = models.StatusNoFixNeeded
		}
	case "failed":
		incident.Status = models.StatusFailed
	case "no_fix_needed":
		incident.Status = models.StatusNoFixNeeded
	default:
		s.logger.Error("unknown workflow status", map[string]interface{}{
			"status":      payload.Status,
			"incident_id": payload.IncidentID,
		})
		http.Error(w, "unknown status", http.StatusBadRequest)
		return
	}

	// Update diagnosis if provided
	if payload.Diagnosis != "" {
		incident.Diagnosis = &payload.Diagnosis
	}

	// Update the incident in the database
	if err := s.repository.Update(incident); err != nil {
		s.logger.Error("failed to update incident after workflow completion", map[string]interface{}{
			"error":       err.Error(),
			"incident_id": payload.IncidentID,
		})
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Log the workflow completion event
	eventType := models.EventPRCreated
	if payload.Status == "failed" {
		eventType = models.EventIncidentFailed
	}

	event := &models.IncidentEvent{
		IncidentID: payload.IncidentID,
		EventType:  eventType,
		EventData: map[string]interface{}{
			"status":          payload.Status,
			"pull_request_url": payload.PullRequestURL,
			"diagnosis":       payload.Diagnosis,
		},
	}

	if err := s.repository.LogEvent(event); err != nil {
		s.logger.Error("failed to log workflow completion event", map[string]interface{}{
			"error":       err.Error(),
			"incident_id": payload.IncidentID,
		})
		// Don't fail the request if event logging fails
	}

	// Process queued incidents for this repository
	nextIncident := s.githubClient.DecrementActive(payload.Repository)
	if nextIncident != nil {
		s.logger.Info("processing queued incident", map[string]interface{}{
			"incident_id": nextIncident.ID,
			"repository":  nextIncident.Repository,
		})

		// Log dequeue event
		dequeueEvent := &models.IncidentEvent{
			IncidentID: nextIncident.ID,
			EventType:  models.EventDequeuedForRemediation,
			EventData: map[string]interface{}{
				"repository": nextIncident.Repository,
			},
		}
		if err := s.repository.LogEvent(dequeueEvent); err != nil {
			s.logger.Error("failed to log dequeue event", map[string]interface{}{
				"error":       err.Error(),
				"incident_id": nextIncident.ID,
			})
		}

		// Trigger workflow for the queued incident
		go func(inc *models.Incident) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Get the branch from config (default to "main")
			branch := "main"
			if s.config != nil && s.config.ServiceMappings != nil {
				for _, mapping := range s.config.ServiceMappings {
					if mapping.Repository == inc.Repository {
						branch = mapping.Branch
						break
					}
				}
			}

			_, err := s.githubClient.DispatchWorkflow(ctx, inc, branch)
			if err != nil {
				s.logger.Error("failed to dispatch workflow for queued incident", map[string]interface{}{
					"error":       err.Error(),
					"incident_id": inc.ID,
					"repository":  inc.Repository,
				})

				// Update incident status to failed
				if updateErr := s.repository.UpdateStatus(inc.ID, models.StatusFailed); updateErr != nil {
					s.logger.Error("failed to update queued incident status", map[string]interface{}{
						"error":       updateErr.Error(),
						"incident_id": inc.ID,
					})
				}
				return
			}

			// Update incident status to workflow_triggered
			triggerTime := time.Now()
			inc.TriggeredAt = &triggerTime
			inc.Status = models.StatusWorkflowTriggered
			if updateErr := s.repository.Update(inc); updateErr != nil {
				s.logger.Error("failed to update queued incident after dispatch", map[string]interface{}{
					"error":       updateErr.Error(),
					"incident_id": inc.ID,
				})
			}
		}(nextIncident)
	}

	// Log success
	s.logger.Info("workflow status updated", map[string]interface{}{
		"incident_id":      payload.IncidentID,
		"status":           payload.Status,
		"pull_request_url": payload.PullRequestURL,
		"repository":       payload.Repository,
		"duration_ms":      time.Since(startTime).Milliseconds(),
	})

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "updated",
		"message": "incident status updated successfully",
	})
}

// ConfigResponse represents the configuration data returned to the dashboard
type ConfigResponse struct {
	ServiceMappings []ServiceMappingResponse `json:"service_mappings"`
}

// ServiceMappingResponse represents a service-to-repository mapping
type ServiceMappingResponse struct {
	ServiceName string `json:"service_name"`
	Repository  string `json:"repository"`
	Branch      string `json:"branch"`
}

// handleGetConfig handles requests for configuration data
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	// Build response from current configuration
	response := ConfigResponse{
		ServiceMappings: make([]ServiceMappingResponse, 0, len(s.config.ServiceMappings)),
	}

	for _, mapping := range s.config.ServiceMappings {
		response.ServiceMappings = append(response.ServiceMappings, ServiceMappingResponse{
			ServiceName: mapping.ServiceName,
			Repository:  mapping.Repository,
			Branch:      mapping.Branch,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
