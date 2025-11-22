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
)

// Server represents the HTTP server
type Server struct {
	config     *config.Config
	db         *database.DB
	redis      *database.RedisClient
	repository *database.IncidentRepository
	adapters   *adapters.Registry
	logger     *Logger
	metrics    *Metrics
	router     *chi.Mux
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, db *database.DB, redis *database.RedisClient) *Server {
	s := &Server{
		config:     cfg,
		db:         db,
		redis:      redis,
		repository: database.NewIncidentRepository(db),
		adapters:   adapters.NewRegistry(),
		logger:     NewLogger(),
		metrics:    NewMetrics(),
		router:     chi.NewRouter(),
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incidents)
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
