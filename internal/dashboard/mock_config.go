package dashboard

import (
	"context"

	"github.com/eggybyte-technology/yao-oracle/core/config"
)

// MockConfigInformer implements a mock Kubernetes Informer for testing.
//
// This provides static configuration without requiring a Kubernetes cluster.
// It simulates the config.K8sInformer interface for testing purposes.
type MockConfigInformer struct {
	cfg      config.Config
	password string
}

// NewMockConfigInformer creates a new mock config informer with test data.
//
// Parameters:
//   - password: Dashboard password for authentication
//
// Returns:
//   - *MockConfigInformer: A mock informer ready to use
func NewMockConfigInformer(password string) *MockConfigInformer {
	return &MockConfigInformer{
		cfg: config.Config{
			Proxy: &config.ProxyConfig{
				Namespaces: []config.Namespace{
					{
						Name:         "game-app",
						Description:  "Gaming application cache",
						APIKey:       "test-game-key",
						MaxMemoryMB:  512,
						DefaultTTL:   60,
						RateLimitQPS: 100,
					},
					{
						Name:         "ads-service",
						Description:  "Advertisement service cache",
						APIKey:       "test-ads-key",
						MaxMemoryMB:  256,
						DefaultTTL:   120,
						RateLimitQPS: 50,
					},
					{
						Name:         "analytics",
						Description:  "Analytics data cache",
						APIKey:       "test-analytics-key",
						MaxMemoryMB:  1024,
						DefaultTTL:   300,
						RateLimitQPS: 200,
					},
				},
			},
			Dashboard: &config.DashboardConfig{
				Password:        password,
				JWTSecret:       "test-jwt-secret",
				RefreshInterval: 5,
				Theme:           "dark",
			},
		},
		password: password,
	}
}

// GetConfig returns the mock configuration.
func (m *MockConfigInformer) GetConfig() config.Config {
	return m.cfg
}

// Start implements the informer interface (no-op for mock).
func (m *MockConfigInformer) Start(ctx context.Context, onChange func(kind string, data map[string][]byte)) error {
	// No-op for mock - configuration is static
	return nil
}

// Stop implements the informer interface (no-op for mock).
func (m *MockConfigInformer) Stop() {
	// No-op for mock
}
