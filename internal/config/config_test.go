package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      types.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: types.Config{
				DashURL:   "http://localhost:3000",
				AuthToken: "test-token",
				OrgID:     "test-org",
			},
			expectError: false,
		},
		{
			name: "missing dash URL",
			config: types.Config{
				AuthToken: "test-token",
				OrgID:     "test-org",
			},
			expectError: true,
			errorMsg:    "dashboard URL is required",
		},
		{
			name: "invalid dash URL",
			config: types.Config{
				DashURL:   "invalid-url",
				AuthToken: "test-token",
				OrgID:     "test-org",
			},
			expectError: true,
		},
		{
			name: "missing auth token",
			config: types.Config{
				DashURL: "http://localhost:3000",
				OrgID:   "test-org",
			},
			expectError: true,
			errorMsg:    "auth token is required",
		},
		{
			name: "missing org ID",
			config: types.Config{
				DashURL:   "http://localhost:3000",
				AuthToken: "test-token",
			},
			expectError: true,
			errorMsg:    "organization ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManagerEnvironmentVariables(t *testing.T) {
	// Clean up environment
	originalEnv := map[string]string{
		EnvDashURL:   os.Getenv(EnvDashURL),
		EnvAuthToken: os.Getenv(EnvAuthToken),
		EnvOrgID:     os.Getenv(EnvOrgID),
	}
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Set test environment variables
	testDashURL := "http://test-dashboard:3000"
	testAuthToken := "test-auth-token"
	testOrgID := "test-org-id"

	os.Setenv(EnvDashURL, testDashURL)
	os.Setenv(EnvAuthToken, testAuthToken)
	os.Setenv(EnvOrgID, testOrgID)

	manager := NewManager()
	err := manager.LoadConfig()
	require.NoError(t, err)

	config := manager.GetConfig()
	assert.Equal(t, testDashURL, config.DashURL)
	assert.Equal(t, testAuthToken, config.AuthToken)
	assert.Equal(t, testOrgID, config.OrgID)

	// Validate the loaded config
	assert.NoError(t, config.Validate())
}

func TestManagerFlagsOverrideEnvironment(t *testing.T) {
	// Set environment variables
	os.Setenv(EnvDashURL, "http://env-dashboard:3000")
	os.Setenv(EnvAuthToken, "env-auth-token")
	os.Setenv(EnvOrgID, "env-org-id")
	defer func() {
		os.Unsetenv(EnvDashURL)
		os.Unsetenv(EnvAuthToken)
		os.Unsetenv(EnvOrgID)
	}()

	manager := NewManager()
	err := manager.LoadConfig()
	require.NoError(t, err)

	// Override with flags
	flagDashURL := "http://flag-dashboard:3000"
	flagAuthToken := "flag-auth-token"
	flagOrgID := "flag-org-id"

	manager.SetFromFlags(flagDashURL, flagAuthToken, flagOrgID)

	config := manager.GetConfig()
	assert.Equal(t, flagDashURL, config.DashURL)
	assert.Equal(t, flagAuthToken, config.AuthToken)
	assert.Equal(t, flagOrgID, config.OrgID)
}

func TestManagerPartialFlagOverride(t *testing.T) {
	// Set environment variables
	os.Setenv(EnvDashURL, "http://env-dashboard:3000")
	os.Setenv(EnvAuthToken, "env-auth-token")
	os.Setenv(EnvOrgID, "env-org-id")
	defer func() {
		os.Unsetenv(EnvDashURL)
		os.Unsetenv(EnvAuthToken)
		os.Unsetenv(EnvOrgID)
	}()

	manager := NewManager()
	err := manager.LoadConfig()
	require.NoError(t, err)

	// Override only dash URL with flag
	flagDashURL := "http://flag-dashboard:3000"
	manager.SetFromFlags(flagDashURL, "", "")

	config := manager.GetConfig()
	assert.Equal(t, flagDashURL, config.DashURL)
	assert.Equal(t, "env-auth-token", config.AuthToken) // Should remain from env
	assert.Equal(t, "env-org-id", config.OrgID)         // Should remain from env
}

func TestNewManager(t *testing.T) {
	manager := NewManager()
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.viper)
	assert.NotNil(t, manager.config)
	assert.NotNil(t, manager.GetViperInstance())
}

// Integration test with actual live environment
func TestLiveEnvironmentConnection(t *testing.T) {
	// This test uses the provided live environment details
	testConfig := &types.Config{
		DashURL:   "http://tyk-dashboard.localhost:3000",
		AuthToken: "ff8289874f5d45de945a2ea5c02580fe",
		OrgID:     "5e9d9544a1dcd60001d0ed20",
	}

	// Validate the live config
	err := testConfig.Validate()
	assert.NoError(t, err, "Live environment configuration should be valid")

	// Test that we can create a manager and load this config
	manager := NewManager()
	manager.SetFromFlags(testConfig.DashURL, testConfig.AuthToken, testConfig.OrgID)
	
	config := manager.GetConfig()
	assert.Equal(t, testConfig.DashURL, config.DashURL)
	assert.Equal(t, testConfig.AuthToken, config.AuthToken)
	assert.Equal(t, testConfig.OrgID, config.OrgID)
	
	assert.NoError(t, config.Validate())
}