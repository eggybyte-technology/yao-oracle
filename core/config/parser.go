package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// ParseConfig parses configuration from JSON bytes.
//
// This is a low-level utility function for parsing configuration JSON.
// It validates the configuration after parsing.
//
// Parameters:
//   - data: JSON bytes to parse
//
// Returns:
//   - *Config: Parsed and validated configuration
//   - error: Error if JSON is invalid or configuration fails validation
//
// Example:
//
//	jsonData := []byte(`{"proxy": {"namespaces": [...]}}`)
//	cfg, err := config.ParseConfig(jsonData)
//	if err != nil {
//	    log.Fatal("Failed to parse config:", err)
//	}
func ParseConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// ParseProxyConfig parses proxy configuration from JSON bytes.
//
// Parameters:
//   - data: JSON bytes containing proxy configuration
//
// Returns:
//   - *ProxyConfig: Parsed and validated proxy configuration
//   - error: Error if JSON is invalid or configuration fails validation
func ParseProxyConfig(data []byte) (*ProxyConfig, error) {
	var cfg ProxyConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proxy config: %w", err)
	}

	if err := ValidateProxyConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid proxy configuration: %w", err)
	}

	return &cfg, nil
}

// ParseDashboardConfig parses dashboard configuration from JSON bytes.
//
// Parameters:
//   - data: JSON bytes containing dashboard configuration
//
// Returns:
//   - *DashboardConfig: Parsed and validated dashboard configuration
//   - error: Error if JSON is invalid or configuration fails validation
func ParseDashboardConfig(data []byte) (*DashboardConfig, error) {
	var cfg DashboardConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dashboard config: %w", err)
	}

	if err := ValidateDashboardConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid dashboard configuration: %w", err)
	}

	return &cfg, nil
}

// MarshalConfig converts configuration to JSON bytes.
//
// This function validates the configuration before marshaling.
// It produces formatted JSON with indentation for readability.
//
// Parameters:
//   - cfg: Configuration to marshal
//
// Returns:
//   - []byte: JSON bytes
//   - error: Error if configuration is invalid or marshaling fails
func MarshalConfig(cfg *Config) ([]byte, error) {
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal configuration: %w", err)
	}

	return data, nil
}

// LoadConfigFromFile loads and parses configuration from a JSON file.
//
// This is a convenience function for development and testing. Production
// deployments should use K8sConfigLoader to read directly from Kubernetes API.
//
// Parameters:
//   - filePath: Path to JSON configuration file
//
// Returns:
//   - *Config: Parsed and validated configuration
//   - error: Error if file cannot be read or parsed
//
// Example:
//
//	cfg, err := config.LoadConfigFromFile("/tmp/test-config.json")
//	if err != nil {
//	    log.Fatal("Failed to load config:", err)
//	}
func LoadConfigFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return ParseConfig(data)
}

// SaveConfigToFile saves configuration to a JSON file.
//
// This is a convenience function for development and testing.
// The configuration is validated before saving.
//
// Parameters:
//   - cfg: Configuration to save
//   - filePath: Path to save the JSON file
//
// Returns:
//   - error: Error if validation fails or file cannot be written
//
// Example:
//
//	err := config.SaveConfigToFile(cfg, "/tmp/test-config.json")
//	if err != nil {
//	    log.Fatal("Failed to save config:", err)
//	}
func SaveConfigToFile(cfg *Config, filePath string) error {
	data, err := MarshalConfig(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}
