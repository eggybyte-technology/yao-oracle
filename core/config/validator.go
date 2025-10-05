package config

import (
	"fmt"
)

// ValidateConfig validates the complete configuration structure and business rules.
//
// This function performs comprehensive validation including:
//   - Proxy configuration validation (if present)
//   - Dashboard configuration validation (if present)
//   - Cross-service configuration consistency checks
//
// Parameters:
//   - cfg: The configuration to validate
//
// Returns:
//   - error: nil if valid, error describing the validation failure otherwise
//
// Example:
//
//	cfg, err := loader.LoadFullConfig(ctx, namespace, secretName)
//	if err != nil {
//	    return err
//	}
//
//	if err := ValidateConfig(cfg); err != nil {
//	    log.Fatal("Invalid configuration:", err)
//	}
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Validate proxy configuration if present
	if cfg.Proxy != nil {
		if err := ValidateProxyConfig(cfg.Proxy); err != nil {
			return fmt.Errorf("proxy config validation failed: %w", err)
		}
	}

	// Validate dashboard configuration if present
	if cfg.Dashboard != nil {
		if err := ValidateDashboardConfig(cfg.Dashboard); err != nil {
			return fmt.Errorf("dashboard config validation failed: %w", err)
		}
	}

	// At least one service configuration must be present
	if cfg.Proxy == nil && cfg.Dashboard == nil {
		return fmt.Errorf("configuration must contain at least proxy or dashboard config")
	}

	return nil
}

// ValidateProxyConfig validates proxy-specific configuration.
//
// Validation rules:
//   - At least one namespace must be defined
//   - Namespace names must be unique and non-empty
//   - API keys must be non-empty for each namespace
//   - Resource limits must be non-negative if specified
//
// Parameters:
//   - cfg: The proxy configuration to validate
//
// Returns:
//   - error: nil if valid, error describing the validation failure otherwise
func ValidateProxyConfig(cfg *ProxyConfig) error {
	if cfg == nil {
		return fmt.Errorf("proxy configuration cannot be nil")
	}

	// At least one namespace is required
	if len(cfg.Namespaces) == 0 {
		return fmt.Errorf("at least one namespace must be defined")
	}

	// Check each namespace
	namespaceNames := make(map[string]bool)
	apiKeys := make(map[string]bool)

	for i, ns := range cfg.Namespaces {
		// Validate namespace name
		if ns.Name == "" {
			return fmt.Errorf("namespace[%d]: name cannot be empty", i)
		}

		// Check for duplicate namespace names
		if namespaceNames[ns.Name] {
			return fmt.Errorf("namespace[%d]: duplicate namespace name '%s'", i, ns.Name)
		}
		namespaceNames[ns.Name] = true

		// Validate API key
		if ns.APIKey == "" {
			return fmt.Errorf("namespace[%d] (%s): API key cannot be empty", i, ns.Name)
		}

		// Warn about duplicate API keys (not an error, but not recommended)
		if apiKeys[ns.APIKey] {
			// Note: This is a warning, not an error
			// Multiple namespaces could theoretically share the same key
			// but it's not recommended for security reasons
		}
		apiKeys[ns.APIKey] = true

		// Validate resource limits (if specified)
		if ns.MaxMemoryMB < 0 {
			return fmt.Errorf("namespace[%d] (%s): maxMemoryMB cannot be negative, got %d", i, ns.Name, ns.MaxMemoryMB)
		}

		if ns.DefaultTTL < 0 {
			return fmt.Errorf("namespace[%d] (%s): defaultTTL cannot be negative, got %d", i, ns.Name, ns.DefaultTTL)
		}

		if ns.RateLimitQPS < 0 {
			return fmt.Errorf("namespace[%d] (%s): rateLimitQPS cannot be negative, got %d", i, ns.Name, ns.RateLimitQPS)
		}
	}

	return nil
}

// ValidateDashboardConfig validates dashboard-specific configuration.
//
// Validation rules:
//   - Password must not be empty
//   - Password must be at least 8 characters for security
//   - Refresh interval must be non-negative if specified
//
// Parameters:
//   - cfg: The dashboard configuration to validate
//
// Returns:
//   - error: nil if valid, error describing the validation failure otherwise
func ValidateDashboardConfig(cfg *DashboardConfig) error {
	if cfg == nil {
		return fmt.Errorf("dashboard configuration cannot be nil")
	}

	// Validate password
	if cfg.Password == "" {
		return fmt.Errorf("dashboard password cannot be empty")
	}

	if len(cfg.Password) < 8 {
		return fmt.Errorf("dashboard password must be at least 8 characters, got %d", len(cfg.Password))
	}

	// Validate refresh interval
	if cfg.RefreshInterval < 0 {
		return fmt.Errorf("refresh interval must be non-negative, got %d", cfg.RefreshInterval)
	}

	// Validate theme (if specified)
	if cfg.Theme != "" && cfg.Theme != "light" && cfg.Theme != "dark" {
		return fmt.Errorf("theme must be 'light' or 'dark', got '%s'", cfg.Theme)
	}

	return nil
}

// ValidateNamespace validates a single namespace configuration.
//
// This is a helper function for validating namespaces independently.
//
// Parameters:
//   - ns: The namespace to validate
//
// Returns:
//   - error: nil if valid, error describing the validation failure otherwise
func ValidateNamespace(ns *Namespace) error {
	if ns == nil {
		return fmt.Errorf("namespace cannot be nil")
	}

	if ns.Name == "" {
		return fmt.Errorf("namespace name cannot be empty")
	}

	if ns.APIKey == "" {
		return fmt.Errorf("namespace '%s': API key cannot be empty", ns.Name)
	}

	if ns.MaxMemoryMB < 0 {
		return fmt.Errorf("namespace '%s': maxMemoryMB cannot be negative, got %d", ns.Name, ns.MaxMemoryMB)
	}

	if ns.DefaultTTL < 0 {
		return fmt.Errorf("namespace '%s': defaultTTL cannot be negative, got %d", ns.Name, ns.DefaultTTL)
	}

	if ns.RateLimitQPS < 0 {
		return fmt.Errorf("namespace '%s': rateLimitQPS cannot be negative, got %d", ns.Name, ns.RateLimitQPS)
	}

	return nil
}
