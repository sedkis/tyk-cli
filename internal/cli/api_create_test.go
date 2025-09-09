package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOASForCreate(t *testing.T) {
	tests := []struct {
		name         string
		apiName      string
		description  string
		version      string
		upstreamURL  string
		listenPath   string
		customDomain string
		checkResult  func(t *testing.T, result map[string]interface{})
	}{
		{
			name:        "basic API creation",
			apiName:     "User Service",
			description: "User management API",
			version:     "v1",
			upstreamURL: "https://users.api.com",
			listenPath:  "/user-service/",
			checkResult: func(t *testing.T, result map[string]interface{}) {
				// Check basic OAS structure
				assert.Equal(t, "3.0.0", result["openapi"])
				
				info, ok := result["info"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "User Service", info["title"])
				assert.Equal(t, "User management API", info["description"])
				assert.Equal(t, "v1", info["version"])
				
				servers, ok := result["servers"].([]interface{})
				require.True(t, ok)
				require.Len(t, servers, 1)
				
				server, ok := servers[0].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "https://users.api.com", server["url"])
				
				// Check Tyk extensions
				tykExt, ok := result["x-tyk-api-gateway"].(map[string]interface{})
				require.True(t, ok)
				
				extInfo, ok := tykExt["info"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "User Service", extInfo["name"])
				
				state, ok := extInfo["state"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, state["active"])
				
				upstream, ok := tykExt["upstream"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "https://users.api.com", upstream["url"])
				
				server, ok = tykExt["server"].(map[string]interface{})
				require.True(t, ok)
				
				listenPath, ok := server["listenPath"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "/user-service/", listenPath["value"])
				assert.Equal(t, true, listenPath["strip"])
			},
		},
		{
			name:         "API with custom domain",
			apiName:      "Payment API",
			description:  "Payment processing API",
			version:      "v2",
			upstreamURL:  "https://payments.internal",
			listenPath:   "/payments/v2/",
			customDomain: "api.company.com",
			checkResult: func(t *testing.T, result map[string]interface{}) {
				// Check custom domain in Tyk extensions
				tykExt, ok := result["x-tyk-api-gateway"].(map[string]interface{})
				require.True(t, ok)
				
				server, ok := tykExt["server"].(map[string]interface{})
				require.True(t, ok)
				
				customDomain, ok := server["customDomain"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, customDomain["enabled"])
				assert.Equal(t, "api.company.com", customDomain["name"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generateOASForCreate(
				tt.apiName,
				tt.description,
				tt.version,
				tt.upstreamURL,
				tt.listenPath,
				tt.customDomain,
			)
			
			require.NoError(t, err)
			require.NotNil(t, result)
			
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestNewAPICreateCommand(t *testing.T) {
	cmd := NewAPICreateCommand()
	
	assert.Equal(t, "create", cmd.Use)
	assert.Equal(t, "Create a new API from scratch", cmd.Short)
	assert.Contains(t, cmd.Long, "Create a new OAS API from scratch")
	
	// Check required flags
	nameFlag := cmd.Flags().Lookup("name")
	require.NotNil(t, nameFlag)
	
	upstreamFlag := cmd.Flags().Lookup("upstream-url")
	require.NotNil(t, upstreamFlag)
	
	// Check optional flags
	listenPathFlag := cmd.Flags().Lookup("listen-path")
	require.NotNil(t, listenPathFlag)
	
	versionFlag := cmd.Flags().Lookup("version-name")
	require.NotNil(t, versionFlag)
	assert.Equal(t, "v1", versionFlag.DefValue)
	
	customDomainFlag := cmd.Flags().Lookup("custom-domain")
	require.NotNil(t, customDomainFlag)
	
	descriptionFlag := cmd.Flags().Lookup("description")
	require.NotNil(t, descriptionFlag)
}