package dashboard

import (
	"context"

	"github.com/eggybyte-technology/yao-oracle/core/config"
)

// ConfigInformer defines the interface for configuration management.
//
// This interface abstracts configuration loading and watching, allowing for
// different implementations (real Kubernetes Informer or mock for testing).
//
// Implementations must be thread-safe for concurrent access.
type ConfigInformer interface {
	// GetConfig returns the current configuration.
	//
	// This method should be safe for concurrent calls and return
	// a consistent snapshot of the configuration at the time of the call.
	//
	// Returns:
	//   - config.Config: The current configuration
	GetConfig() config.Config

	// Start begins watching for configuration changes.
	//
	// For real implementations, this starts a Kubernetes Informer to watch
	// ConfigMap/Secret changes. For mock implementations, this may be a no-op.
	//
	// Parameters:
	//   - ctx: Context for cancellation
	//   - onChange: Callback function invoked when configuration changes
	//                kind: "ConfigMap" or "Secret"
	//                data: Map of configuration data (key -> value)
	//
	// Returns:
	//   - error: Error if the informer fails to start
	Start(ctx context.Context, onChange func(kind string, data map[string][]byte)) error

	// Stop gracefully shuts down the informer.
	//
	// This should clean up any resources (goroutines, connections, etc.)
	// used by the informer. After calling Stop, the informer should not
	// be reused.
	Stop()
}
