package adapters

import (
	"net/http"

	"github.com/your-org/ai-sre-platform/incident-service/internal/models"
)

// WebhookAdapter defines the interface for webhook adapters
type WebhookAdapter interface {
	// Validate checks if the webhook payload is valid and authentic
	Validate(r *http.Request) error

	// Parse transforms the provider-specific payload into our Incident struct
	Parse(body []byte) (*models.Incident, error)

	// ProviderName returns the name of the observability provider
	ProviderName() string
}

// Registry manages webhook adapters
type Registry struct {
	adapters map[string]WebhookAdapter
}

// NewRegistry creates a new adapter registry
func NewRegistry() *Registry {
	r := &Registry{
		adapters: make(map[string]WebhookAdapter),
	}

	// Register all adapters
	r.Register(NewDatadogAdapter())
	r.Register(NewPagerDutyAdapter())
	r.Register(NewGrafanaAdapter())
	r.Register(NewSentryAdapter())

	return r
}

// Register adds an adapter to the registry
func (r *Registry) Register(adapter WebhookAdapter) {
	r.adapters[adapter.ProviderName()] = adapter
}

// Get retrieves an adapter by provider name
func (r *Registry) Get(provider string) (WebhookAdapter, bool) {
	adapter, ok := r.adapters[provider]
	return adapter, ok
}

// List returns all registered provider names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}
