package api

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	IncidentIngestionTotal   *prometheus.CounterVec
	IncidentIngestionLatency *prometheus.HistogramVec
	WorkflowDispatchTotal    *prometheus.CounterVec
	WorkflowDispatchLatency  *prometheus.HistogramVec
	IncidentQueueDepth       prometheus.Gauge
	ActiveWorkflows          *prometheus.GaugeVec
}

// NewMetrics creates and registers Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		IncidentIngestionTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "incident_ingestion_total",
				Help: "Total number of incidents received",
			},
			[]string{"provider", "status"},
		),
		IncidentIngestionLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "incident_ingestion_latency_seconds",
				Help:    "Latency of incident ingestion",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"provider"},
		),
		WorkflowDispatchTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "workflow_dispatch_total",
				Help: "Total number of workflow dispatches",
			},
			[]string{"repository", "status"},
		),
		WorkflowDispatchLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "workflow_dispatch_latency_seconds",
				Help:    "Latency of workflow dispatch",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"repository"},
		),
		IncidentQueueDepth: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "incident_queue_depth",
				Help: "Number of incidents waiting in queue",
			},
		),
		ActiveWorkflows: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "active_workflows",
				Help: "Number of active workflows per repository",
			},
			[]string{"repository"},
		),
	}
}
