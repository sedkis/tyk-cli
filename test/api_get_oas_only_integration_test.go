// +build integration

package test

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/internal/client"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// TestAPIGetOASOnly_IntegrationTests validates the --oas-only flag functionality
// against a live Tyk Dashboard environment
func TestAPIGetOASOnly_IntegrationTests(t *testing.T) {
	// Use the live environment configuration
	liveConfig := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {
				Name:         "test",
				DashboardURL: "http://tyk-dashboard.localhost:3000",
				AuthToken:    "ff8289874f5d45de945a2ea5c02580fe",
				OrgID:        "5e9d9544a1dcd60001d0ed20",
			},
		},
	}

	t.Logf("Testing --oas-only flag with live environment: %s", liveConfig.Environments["test"].DashboardURL)

	// Test API ID that we know exists in the test environment
	testAPIID := "b84fe1a04e5648927971c0557971565c"

	t.Run("Prerequisites - API exists", func(t *testing.T) {
		// Verify the API exists before running our tests
		c, err := client.NewClient(liveConfig)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		api, err := c.GetOASAPI(ctx, testAPIID, "")
		if err != nil {
			t.Skipf("Test API %s not available, skipping integration tests: %v", testAPIID, err)
		}

		require.NotNil(t, api)
		t.Logf("✓ Test API '%s' (%s) is available", testAPIID, api.Name)
	})

	t.Run("OAS-only JSON output removes Tyk extensions", func(t *testing.T) {
		cmd := exec.Command("./build/tyk", "api", "get", testAPIID, "--oas-only", "--json")
		cmd.Env = []string{
			"TYK_DASH_URL=http://tyk-dashboard.localhost:3000",
			"TYK_AUTH_TOKEN=ff8289874f5d45de945a2ea5c02580fe",
			"TYK_ORG_ID=5e9d9544a1dcd60001d0ed20",
		}

		output, err := cmd.Output()
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(output, &result)
		require.NoError(t, err)

		// Verify x-tyk-api-gateway is NOT present
		_, hasTykExt := result["x-tyk-api-gateway"]
		assert.False(t, hasTykExt, "x-tyk-api-gateway should not be present in --oas-only output")

		// Verify standard OAS fields are present
		assert.Equal(t, "3.0.3", result["openapi"])
		assert.NotNil(t, result["info"])
		assert.NotNil(t, result["paths"])

		t.Log("✓ --oas-only JSON output correctly removes Tyk extensions")
	})

	t.Run("OAS-only YAML output removes Tyk extensions", func(t *testing.T) {
		cmd := exec.Command("./build/tyk", "api", "get", testAPIID, "--oas-only")
		cmd.Env = []string{
			"TYK_DASH_URL=http://tyk-dashboard.localhost:3000",
			"TYK_AUTH_TOKEN=ff8289874f5d45de945a2ea5c02580fe",
			"TYK_ORG_ID=5e9d9544a1dcd60001d0ed20",
		}

		output, err := cmd.Output()
		require.NoError(t, err)

		yamlOutput := string(output)

		// Verify x-tyk-api-gateway is NOT present
		assert.NotContains(t, yamlOutput, "x-tyk-api-gateway", "x-tyk-api-gateway should not be present in --oas-only YAML output")

		// Verify standard OAS fields are present
		assert.Contains(t, yamlOutput, "openapi: 3.0.3")
		assert.Contains(t, yamlOutput, "info:")
		assert.Contains(t, yamlOutput, "paths:")

		t.Log("✓ --oas-only YAML output correctly removes Tyk extensions")
	})

	t.Run("OAS-only mode produces no stderr output", func(t *testing.T) {
		cmd := exec.Command("./build/tyk", "api", "get", testAPIID, "--oas-only")
		cmd.Env = []string{
			"TYK_DASH_URL=http://tyk-dashboard.localhost:3000",
			"TYK_AUTH_TOKEN=ff8289874f5d45de945a2ea5c02580fe",
			"TYK_ORG_ID=5e9d9544a1dcd60001d0ed20",
		}

		_, stderr, err := runCmdWithOutput(cmd)
		require.NoError(t, err)

		// Stderr should be empty in OAS-only mode
		assert.Empty(t, strings.TrimSpace(stderr), "No API summary should be shown in --oas-only mode")

		t.Log("✓ --oas-only mode produces clean output with no API summary")
	})

	t.Run("Regular mode includes Tyk extensions and API summary", func(t *testing.T) {
		cmd := exec.Command("./build/tyk", "api", "get", testAPIID, "--json")
		cmd.Env = []string{
			"TYK_DASH_URL=http://tyk-dashboard.localhost:3000",
			"TYK_AUTH_TOKEN=ff8289874f5d45de945a2ea5c02580fe",
			"TYK_ORG_ID=5e9d9544a1dcd60001d0ed20",
		}

		output, err := cmd.Output()
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(output, &result)
		require.NoError(t, err)

		// Verify this is the full API response structure
		assert.NotNil(t, result["id"])
		assert.NotNil(t, result["name"])
		assert.NotNil(t, result["oas"])

		// Verify that OAS contains x-tyk-api-gateway
		oasData, ok := result["oas"].(map[string]interface{})
		require.True(t, ok, "OAS field should be a map")
		_, hasTykExt := oasData["x-tyk-api-gateway"]
		assert.True(t, hasTykExt, "x-tyk-api-gateway should be present in normal mode")

		t.Log("✓ Regular mode correctly includes Tyk extensions")
	})

	t.Run("Regular mode YAML shows API summary", func(t *testing.T) {
		cmd := exec.Command("./build/tyk", "api", "get", testAPIID)
		cmd.Env = []string{
			"TYK_DASH_URL=http://tyk-dashboard.localhost:3000",
			"TYK_AUTH_TOKEN=ff8289874f5d45de945a2ea5c02580fe",
			"TYK_ORG_ID=5e9d9544a1dcd60001d0ed20",
		}

		stdout, stderr, err := runCmdWithOutput(cmd)
		require.NoError(t, err)

		// Verify API summary is shown in stderr
		assert.Contains(t, stderr, "API Summary:")
		assert.Contains(t, stderr, "ID:")
		assert.Contains(t, stderr, "Name:")
		assert.Contains(t, stderr, "OpenAPI Specification:")

		// Verify YAML contains x-tyk-api-gateway
		assert.Contains(t, stdout, "x-tyk-api-gateway:")

		t.Log("✓ Regular mode shows API summary and includes Tyk extensions")
	})

	t.Run("Error handling with --oas-only flag", func(t *testing.T) {
		cmd := exec.Command("./build/tyk", "api", "get", "non-existent-api-12345", "--oas-only", "--json")
		cmd.Env = []string{
			"TYK_DASH_URL=http://tyk-dashboard.localhost:3000",
			"TYK_AUTH_TOKEN=ff8289874f5d45de945a2ea5c02580fe",
			"TYK_ORG_ID=5e9d9544a1dcd60001d0ed20",
		}

		err := cmd.Run()
		assert.Error(t, err, "Command should fail for non-existent API")

		t.Log("✓ --oas-only flag correctly handles API not found errors")
	})
}

// Helper function to run command and capture both stdout and stderr
func runCmdWithOutput(cmd *exec.Cmd) (stdout, stderr string, err error) {
	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}