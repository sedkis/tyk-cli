package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

func TestGenerateTOMLConfig(t *testing.T) {
	config := &types.Config{
		DashURL:   "http://localhost:3000",
		AuthToken: "test-token",
		OrgID:     "test-org",
	}

	toml := generateTOMLConfig(config)
	
	assert.Contains(t, toml, `dash_url = "http://localhost:3000"`)
	assert.Contains(t, toml, `auth_token = "test-token"`)
	assert.Contains(t, toml, `org_id = "test-org"`)
	assert.Contains(t, toml, "# Tyk CLI Configuration")
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{"empty token", "", "(not set)"},
		{"short token", "abc", "***"},
		{"normal token", "abcdef123456", "abcd****3456"},
		{"long token", "ff8289874f5d45de945a2ea5c02580fe", "ff82****80fe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskToken(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetConfigDir(t *testing.T) {
	configDir, err := getConfigDir()
	require.NoError(t, err)
	
	// Should be a path under user's config directory
	assert.Contains(t, configDir, "tyk")
	
	// Should be an absolute path
	assert.True(t, filepath.IsAbs(configDir))
}

func TestEnvironmentStruct(t *testing.T) {
	env := Environment{
		Name:      "test",
		DashURL:   "http://localhost:3000",
		AuthToken: "token",
		OrgID:     "org",
		IsActive:  true,
	}
	
	assert.Equal(t, "test", env.Name)
	assert.Equal(t, "http://localhost:3000", env.DashURL)
	assert.Equal(t, "token", env.AuthToken)
	assert.Equal(t, "org", env.OrgID)
	assert.True(t, env.IsActive)
}

func TestConfigFileOperations(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "tyk-cli-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test config file creation and content
	configFile := filepath.Join(tmpDir, "cli.toml")
	config := &types.Config{
		DashURL:   "http://test:3000",
		AuthToken: "test-token-123",
		OrgID:     "test-org-456",
	}

	content := generateTOMLConfig(config)
	err = os.WriteFile(configFile, []byte(content), 0600)
	require.NoError(t, err)

	// Verify file was created with correct permissions
	info, err := os.Stat(configFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Verify content
	savedContent, err := os.ReadFile(configFile)
	require.NoError(t, err)
	
	savedStr := string(savedContent)
	assert.Contains(t, savedStr, "dash_url = \"http://test:3000\"")
	assert.Contains(t, savedStr, "auth_token = \"test-token-123\"")
	assert.Contains(t, savedStr, "org_id = \"test-org-456\"")
}

func TestNewConfigCommand(t *testing.T) {
	cmd := NewConfigCommand()
	
	assert.Equal(t, "config", cmd.Use)
	assert.Contains(t, cmd.Short, "global CLI configuration")
	
	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)
	
	var cmdNames []string
	for _, subcmd := range subcommands {
		cmdNames = append(cmdNames, subcmd.Use)
	}
	
	assert.Contains(t, cmdNames, "set")
	assert.Contains(t, cmdNames, "get") 
	assert.Contains(t, cmdNames, "unset")
}

func TestNewInitCommand(t *testing.T) {
	cmd := NewInitCommand()
	
	assert.Equal(t, "init", cmd.Use)
	assert.Contains(t, cmd.Short, "Interactive setup wizard")
	assert.Contains(t, cmd.Long, "🚀")
	
	// Check flags
	skipTestFlag := cmd.Flags().Lookup("skip-test")
	assert.NotNil(t, skipTestFlag)
	assert.Equal(t, "bool", skipTestFlag.Value.Type())
	
	quickFlag := cmd.Flags().Lookup("quick")
	assert.NotNil(t, quickFlag)
	assert.Equal(t, "bool", quickFlag.Value.Type())
}