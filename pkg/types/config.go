package types

import (
	"errors"
	"fmt"
	"net/url"
)

// Config holds all configuration for the Tyk CLI
type Config struct {
	DashURL   string `mapstructure:"dash_url" yaml:"dash_url" json:"dash_url"`
	AuthToken string `mapstructure:"auth_token" yaml:"auth_token" json:"auth_token"`
	OrgID     string `mapstructure:"org_id" yaml:"org_id" json:"org_id"`
	
	// Environment management
	DefaultEnvironment string                   `mapstructure:"default_environment" yaml:"default_environment" json:"default_environment"`
	Environments       map[string]*Environment  `mapstructure:"environments" yaml:"environments" json:"environments"`
}

// Environment represents a named configuration environment
type Environment struct {
	Name      string `mapstructure:"name" yaml:"name" json:"name"`
	DashURL   string `mapstructure:"dash_url" yaml:"dash_url" json:"dash_url"`
	AuthToken string `mapstructure:"auth_token" yaml:"auth_token" json:"auth_token"`
	OrgID     string `mapstructure:"org_id" yaml:"org_id" json:"org_id"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// If we have environments, validate the active one
	if c.DefaultEnvironment != "" && len(c.Environments) > 0 {
		env, exists := c.Environments[c.DefaultEnvironment]
		if !exists {
			return fmt.Errorf("default environment '%s' not found", c.DefaultEnvironment)
		}
		return env.Validate()
	}

	// Otherwise validate the direct config fields
	if c.DashURL == "" {
		return errors.New("dashboard URL is required (TYK_DASH_URL or --dash-url)")
	}

	// Validate URL format
	parsedURL, err := url.Parse(c.DashURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid dashboard URL format: %s", c.DashURL)
	}

	if c.AuthToken == "" {
		return errors.New("auth token is required (TYK_AUTH_TOKEN or --auth-token)")
	}

	if c.OrgID == "" {
		return errors.New("organization ID is required (TYK_ORG_ID or --org-id)")
	}

	return nil
}

// GetActiveEnvironment returns the active environment configuration
func (c *Config) GetActiveEnvironment() (*Environment, error) {
	if c.DefaultEnvironment != "" && len(c.Environments) > 0 {
		env, exists := c.Environments[c.DefaultEnvironment]
		if !exists {
			return nil, fmt.Errorf("default environment '%s' not found", c.DefaultEnvironment)
		}
		return env, nil
	}

	// Return config as environment for backward compatibility
	return &Environment{
		Name:      "default",
		DashURL:   c.DashURL,
		AuthToken: c.AuthToken,
		OrgID:     c.OrgID,
	}, nil
}

// GetEffectiveConfig returns the configuration values to use (from environment or direct config)
func (c *Config) GetEffectiveConfig() (string, string, string, error) {
	env, err := c.GetActiveEnvironment()
	if err != nil {
		return "", "", "", err
	}
	return env.DashURL, env.AuthToken, env.OrgID, nil
}

// Validate checks if an environment configuration is valid
func (e *Environment) Validate() error {
	if e.DashURL == "" {
		return fmt.Errorf("dashboard URL is required for environment '%s'", e.Name)
	}

	// Validate URL format
	parsedURL, err := url.Parse(e.DashURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid dashboard URL format for environment '%s': %s", e.Name, e.DashURL)
	}

	if e.AuthToken == "" {
		return fmt.Errorf("auth token is required for environment '%s'", e.Name)
	}

	if e.OrgID == "" {
		return fmt.Errorf("organization ID is required for environment '%s'", e.Name)
	}

	return nil
}

// ExitCode represents different types of CLI exit codes
type ExitCode int

const (
	ExitSuccess     ExitCode = 0 // Success
	ExitGeneral     ExitCode = 1 // Generic failure (I/O, network, unexpected)
	ExitBadArgs     ExitCode = 2 // Bad arguments (missing file, invalid flag combination)
	ExitNotFound    ExitCode = 3 // Not found (API or version)
	ExitConflict    ExitCode = 4 // Conflict (e.g. creating an API that already exists without --force)
)

// OutputFormat represents the output format for CLI commands
type OutputFormat string

const (
	OutputHuman OutputFormat = "human"
	OutputJSON  OutputFormat = "json"
)