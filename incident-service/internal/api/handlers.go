package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/your-org/ai-sre-platform/incident-service/internal/config"
	"github.com/your-org/ai-sre-platform/incident-service/internal/database"
)

// Server represents the HTTP server
type Server struct {
	config     *config.Config
	db         *database.DB
	redis      *database.RedisClient
	repository *database.IncidentRepository
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
