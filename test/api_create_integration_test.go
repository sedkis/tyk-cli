package test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPICreateIntegration(t *testing.T) {
	if !IsLiveTestEnabled() {
		t.Skip("Live tests are not enabled. Set TYK_TEST_LIVE=true to run.")
	}

	// Setup test configuration
	config := GetTestConfig(t)
	
	tests := []struct {
		name            string
		apiName         string
		upstreamURL     string
		listenPath      string
		customDomain    string
		description     string
		versionName     string
		expectedError   bool
		cleanup         bool
	}{
		{
			name:        "create basic API",
			apiName:     fmt.Sprintf("Test User Service %d", time.Now().Unix()),
			upstreamURL: "https://httpbin.org/anything",
			cleanup:     true,
		},
		{
			name:         "create API with custom path",
			apiName:      fmt.Sprintf("Test Payment API %d", time.Now().Unix()),
			upstreamURL:  "https://httpbin.org/anything",
			listenPath:   "/payments/test/",
			description:  "Test payment processing API",
			versionName:  "v2",
			cleanup:      true,
		},
		{
			name:         "create API with custom domain",
			apiName:      fmt.Sprintf("Test Analytics API %d", time.Now().Unix()),
			upstreamURL:  "https://httpbin.org/anything",
			customDomain: "test-api.example.com",
			description:  "Test analytics and reporting API",
			cleanup:      true,
		},
		{
			name:          "create API with invalid upstream URL",
			apiName:       "Test Invalid API",
			upstreamURL:   "invalid-url",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build command arguments
			args := []string{
				"api", "create",
				"--name", tt.apiName,
				"--upstream-url", tt.upstreamURL,
				"--json", // Use JSON output for easier parsing
			}
			
			if tt.listenPath != "" {
				args = append(args, "--listen-path", tt.listenPath)
			}
			if tt.customDomain != "" {
				args = append(args, "--custom-domain", tt.customDomain)
			}
			if tt.description != "" {
				args = append(args, "--description", tt.description)
			}
			if tt.versionName != "" {
				args = append(args, "--version-name", tt.versionName)
			}

			// Execute command
			output, err := ExecuteCLICommand(config, args...)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err, "Command should succeed")
			require.NotEmpty(t, output, "Output should not be empty")

			// Parse JSON output
			var result map[string]interface{}
			err = json.Unmarshal([]byte(output), &result)
			require.NoError(t, err, "Should be valid JSON")

			// Verify response structure
			assert.Equal(t, "created", result["operation"])
			
			apiID, ok := result["api_id"].(string)
			require.True(t, ok, "api_id should be a string")
			require.NotEmpty(t, apiID, "api_id should not be empty")
			
			assert.Equal(t, tt.apiName, result["name"])
			assert.NotEmpty(t, result["listen_path"])
			assert.NotEmpty(t, result["default_version"])

			// Verify optional fields
			if tt.versionName != "" {
				assert.Equal(t, tt.versionName, result["version_name"])
			} else {
				assert.Equal(t, "v1", result["version_name"])
			}

			if tt.customDomain != "" {
				assert.Equal(t, tt.customDomain, result["custom_domain"])
			}

			// Verify API was actually created by fetching it
			getOutput, err := ExecuteCLICommand(config, "api", "get", apiID, "--json")
			require.NoError(t, err, "Should be able to fetch created API")
			
			var apiData map[string]interface{}
			err = json.Unmarshal([]byte(getOutput), &apiData)
			require.NoError(t, err, "Get response should be valid JSON")
			
			assert.Equal(t, apiID, apiData["id"])
			assert.Equal(t, tt.apiName, apiData["name"])

			// Verify OAS structure by fetching OAS-only
			oasOutput, err := ExecuteCLICommand(config, "api", "get", apiID, "--oas-only")
			require.NoError(t, err, "Should be able to fetch OAS document")
			require.NotEmpty(t, oasOutput, "OAS output should not be empty")
			
			// Basic OAS validation
			assert.Contains(t, oasOutput, "openapi: 3.0.0")
			assert.Contains(t, oasOutput, fmt.Sprintf("title: %s", tt.apiName))
			assert.Contains(t, oasOutput, tt.upstreamURL)

			// Cleanup if requested
			if tt.cleanup {
				t.Cleanup(func() {
					_, err := ExecuteCLICommand(config, "api", "delete", apiID, "--yes", "--json")
					if err != nil {
						t.Logf("Failed to cleanup API %s: %v", apiID, err)
					}
				})
			}
		})
	}
}

func TestAPICreateHumanOutput(t *testing.T) {
	if !IsLiveTestEnabled() {
		t.Skip("Live tests are not enabled. Set TYK_TEST_LIVE=true to run.")
	}

	config := GetTestConfig(t)
	apiName := fmt.Sprintf("Test Human Output API %d", time.Now().Unix())
	
	// Execute command without --json flag for human-readable output
	output, err := ExecuteCLICommand(config, 
		"api", "create",
		"--name", apiName,
		"--upstream-url", "https://httpbin.org/anything",
		"--description", "Test API for human output validation",
	)
	
	require.NoError(t, err)
	require.NotEmpty(t, output)

	// Verify human-readable output format
	assert.Contains(t, output, "✓ API created successfully!")
	assert.Contains(t, output, "API ID:")
	assert.Contains(t, output, "Name:")
	assert.Contains(t, output, "Version:")
	assert.Contains(t, output, "Listen Path:")
	assert.Contains(t, output, "Upstream URL:")
	assert.Contains(t, output, "Default Version:")
	assert.Contains(t, output, "Next steps:")
	assert.Contains(t, output, "tyk api get")
	assert.Contains(t, output, "tyk api get")
	assert.Contains(t, output, "--oas-only")

	// Extract API ID for cleanup
	lines := strings.Split(output, "\n")
	var apiID string
	for _, line := range lines {
		if strings.Contains(line, "API ID:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				apiID = strings.TrimSpace(parts[1])
				break
			}
		}
	}
	
	require.NotEmpty(t, apiID, "Should be able to extract API ID from output")
	
	// Cleanup
	t.Cleanup(func() {
		_, err := ExecuteCLICommand(config, "api", "delete", apiID, "--yes", "--json")
		if err != nil {
			t.Logf("Failed to cleanup API %s: %v", apiID, err)
		}
	})
}

func TestAPICreateWorkflowIntegration(t *testing.T) {
	if !IsLiveTestEnabled() {
		t.Skip("Live tests are not enabled. Set TYK_TEST_LIVE=true to run.")
	}

	config := GetTestConfig(t)
	apiName := fmt.Sprintf("Test Workflow API %d", time.Now().Unix())
	
	// Step 1: Create API
	createOutput, err := ExecuteCLICommand(config, 
		"api", "create",
		"--name", apiName,
		"--upstream-url", "https://httpbin.org/anything",
		"--json",
	)
	
	require.NoError(t, err)
	
	var createResult map[string]interface{}
	err = json.Unmarshal([]byte(createOutput), &createResult)
	require.NoError(t, err)
	
	apiID := createResult["api_id"].(string)
	require.NotEmpty(t, apiID)
	
	// Step 2: Fetch full API configuration
	getOutput, err := ExecuteCLICommand(config, "api", "get", apiID, "--json")
	require.NoError(t, err)
	require.NotEmpty(t, getOutput)
	
	// Step 3: Export OAS-only document
	oasOutput, err := ExecuteCLICommand(config, "api", "get", apiID, "--oas-only")
	require.NoError(t, err)
	require.NotEmpty(t, oasOutput)
	
	// Save to temporary file
	tmpFile, err := os.CreateTemp("", "test-api-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	
	_, err = tmpFile.WriteString(oasOutput)
	require.NoError(t, err)
	tmpFile.Close()
	
	// Step 4: Update API with the same OAS (should work)
	updateOutput, err := ExecuteCLICommand(config, 
		"api", "update-oas", apiID,
		"--file", tmpFile.Name(),
		"--json",
	)
	require.NoError(t, err)
	require.NotEmpty(t, updateOutput)
	
	var updateResult map[string]interface{}
	err = json.Unmarshal([]byte(updateOutput), &updateResult)
	require.NoError(t, err)
	
	assert.Equal(t, "updated", updateResult["operation"])
	assert.Equal(t, apiID, updateResult["api_id"])
	
	// Cleanup
	t.Cleanup(func() {
		_, err := ExecuteCLICommand(config, "api", "delete", apiID, "--yes", "--json")
		if err != nil {
			t.Logf("Failed to cleanup API %s: %v", apiID, err)
		}
	})
}
