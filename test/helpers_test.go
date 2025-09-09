package test

import (
    "bytes"
    "os"
    "testing"

    "github.com/tyktech/tyk-cli/internal/cli"
    "github.com/tyktech/tyk-cli/pkg/types"
)

// IsLiveTestEnabled indicates whether to run live integration tests.
// Controlled via TYK_TEST_LIVE=true
func IsLiveTestEnabled() bool {
    return os.Getenv("TYK_TEST_LIVE") == "true"
}

// GetTestConfig builds a Config from environment variables for live tests.
// TYK_DASH_URL, TYK_AUTH_TOKEN, TYK_ORG_ID are used if present.
func GetTestConfig(t *testing.T) *types.Config {
    t.Helper()
    dash := os.Getenv("TYK_DASH_URL")
    token := os.Getenv("TYK_AUTH_TOKEN")
    org := os.Getenv("TYK_ORG_ID")
    if dash == "" {
        dash = "http://tyk-dashboard.localhost:3000"
    }
    return &types.Config{
        DefaultEnvironment: "test",
        Environments: map[string]*types.Environment{
            "test": {
                Name:         "test",
                DashboardURL: dash,
                AuthToken:    token,
                OrgID:        org,
            },
        },
    }
}

// ExecuteCLICommand runs the CLI with the provided config and arguments,
// returning captured stdout as string and any execution error.
func ExecuteCLICommand(config *types.Config, args ...string) (string, error) {
    // Ensure env matches config so PersistentPreRunE validation passes
    if env, ok := config.Environments[config.DefaultEnvironment]; ok && env != nil {
        if env.DashboardURL != "" {
            os.Setenv("TYK_DASH_URL", env.DashboardURL)
        }
        if env.AuthToken != "" {
            os.Setenv("TYK_AUTH_TOKEN", env.AuthToken)
        }
        if env.OrgID != "" {
            os.Setenv("TYK_ORG_ID", env.OrgID)
        }
    }

    root := cli.NewRootCommand("test", "commit", "time")

    var stdout bytes.Buffer
    root.SetOut(&stdout)
    root.SetErr(&stdout)
    root.SetArgs(args)
    err := root.Execute()
    return stdout.String(), err
}
