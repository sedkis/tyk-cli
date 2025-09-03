package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tyktech/tyk-cli/internal/config"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// NewConfigCommand creates the 'tyk config' command for managing global configuration
func NewConfigCommand() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage global CLI configuration",
		Long:  "Commands for setting up and managing global CLI configuration",
	}

	configCmd.AddCommand(NewConfigSetCommand())
	configCmd.AddCommand(NewConfigGetCommand())
	configCmd.AddCommand(NewConfigUnsetCommand())
	configCmd.AddCommand(NewConfigEnvCommand())

	return configCmd
}

// NewConfigSetCommand creates the 'tyk config set' command
func NewConfigSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set global configuration values",
		Long: `Set global configuration values that will be used by all CLI commands.
Configuration is stored in ~/.config/tyk/cli.toml

Examples:
  tyk config set --dash-url http://localhost:3000
  tyk config set --auth-token your-api-token
  tyk config set --org-id your-org-id
  
  # Set all at once
  tyk config set --dash-url http://localhost:3000 --auth-token token --org-id org`,
		RunE: runConfigSet,
	}

	cmd.Flags().String("dash-url", "", "Tyk Dashboard URL")
	cmd.Flags().String("auth-token", "", "Dashboard API auth token")  
	cmd.Flags().String("org-id", "", "Organization ID")

	return cmd
}

// NewConfigGetCommand creates the 'tyk config get' command
func NewConfigGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Display current configuration",
		Long:  "Display the current global configuration values",
		RunE:  runConfigGet,
	}

	return cmd
}

// NewConfigUnsetCommand creates the 'tyk config unset' command
func NewConfigUnsetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset",
		Short: "Remove configuration values",
		Long: `Remove global configuration values.

Examples:
  tyk config unset --auth-token
  tyk config unset --all`,
		RunE: runConfigUnset,
	}

	cmd.Flags().Bool("dash-url", false, "Remove dashboard URL")
	cmd.Flags().Bool("auth-token", false, "Remove auth token")
	cmd.Flags().Bool("org-id", false, "Remove organization ID")
	cmd.Flags().Bool("all", false, "Remove all configuration")

	return cmd
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	dashURL, _ := cmd.Flags().GetString("dash-url")
	authToken, _ := cmd.Flags().GetString("auth-token")
	orgID, _ := cmd.Flags().GetString("org-id")

	if dashURL == "" && authToken == "" && orgID == "" {
		return fmt.Errorf("at least one configuration value must be provided")
	}

	// Get config directory
	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "cli.toml")
	
	// Load existing config
	manager := config.NewManager()
	
	// Try to load existing config file (ignore errors if file doesn't exist)
	if _, err := os.Stat(configFile); err == nil {
		if err := manager.LoadConfig(); err != nil {
			return fmt.Errorf("failed to load existing config: %w", err)
		}
	}

	// Update with new values
	if dashURL != "" || authToken != "" || orgID != "" {
		manager.SetFromFlags(dashURL, authToken, orgID)
	}

	// Get the updated config
	cfg := manager.GetConfig()

	// Create config file content
	content := generateTOMLConfig(cfg)

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Configuration saved to %s\n", configFile)
	
	// Show what was set
	if dashURL != "" {
		fmt.Printf("  dash_url = %s\n", dashURL)
	}
	if authToken != "" {
		fmt.Printf("  auth_token = %s\n", maskToken(authToken))
	}
	if orgID != "" {
		fmt.Printf("  org_id = %s\n", orgID)
	}

	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	// Load configuration
	manager := config.NewManager()
	if err := manager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()
	
	fmt.Println("Current configuration:")
	
	// Show environment information if available
	if len(cfg.Environments) > 0 {
		fmt.Printf("  default_environment = %s\n\n", cfg.DefaultEnvironment)
		
		fmt.Println("Environments:")
		for name, env := range cfg.Environments {
			marker := ""
			if name == cfg.DefaultEnvironment {
				marker = " (active)"
			}
			fmt.Printf("  %s%s:\n", name, marker)
			fmt.Printf("    dash_url   = %s\n", env.DashURL)
			fmt.Printf("    auth_token = %s\n", maskToken(env.AuthToken))
			fmt.Printf("    org_id     = %s\n", env.OrgID)
			fmt.Println()
		}
	} else {
		// Show direct config for backward compatibility
		fmt.Printf("  dash_url   = %s\n", cfg.DashURL)
		fmt.Printf("  auth_token = %s\n", maskToken(cfg.AuthToken))
		fmt.Printf("  org_id     = %s\n", cfg.OrgID)
		fmt.Println()
	}
	
	// Show source information
	fmt.Println("Configuration sources (in order of precedence):")
	fmt.Println("  1. Command line flags")
	fmt.Println("  2. Environment variables (TYK_*)")
	
	configDir, err := getConfigDir()
	if err == nil {
		configFile := filepath.Join(configDir, "cli.toml")
		if _, err := os.Stat(configFile); err == nil {
			fmt.Printf("  3. Config file: %s\n", configFile)
		} else {
			fmt.Printf("  3. Config file: %s (not found)\n", configFile)
		}
	}

	return nil
}

func runConfigUnset(cmd *cobra.Command, args []string) error {
	removeDashURL, _ := cmd.Flags().GetBool("dash-url")
	removeAuthToken, _ := cmd.Flags().GetBool("auth-token")
	removeOrgID, _ := cmd.Flags().GetBool("org-id")
	removeAll, _ := cmd.Flags().GetBool("all")

	if !removeDashURL && !removeAuthToken && !removeOrgID && !removeAll {
		return fmt.Errorf("specify what to unset: --dash-url, --auth-token, --org-id, or --all")
	}

	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "cli.toml")

	if removeAll {
		if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove config file: %w", err)
		}
		fmt.Printf("All configuration removed from %s\n", configFile)
		return nil
	}

	// Load existing config
	manager := config.NewManager()
	if err := manager.LoadConfig(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load existing config: %w", err)
	}

	cfg := manager.GetConfig()

	// Remove specified values
	if removeDashURL {
		cfg.DashURL = ""
		fmt.Println("Removed dash_url")
	}
	if removeAuthToken {
		cfg.AuthToken = ""
		fmt.Println("Removed auth_token")
	}
	if removeOrgID {
		cfg.OrgID = ""
		fmt.Println("Removed org_id")
	}

	// Write updated config
	content := generateTOMLConfig(cfg)
	
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Configuration updated in %s\n", configFile)
	return nil
}

func getConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userConfigDir, "tyk"), nil
}

func generateTOMLConfig(cfg *types.Config) string {
	content := "# Tyk CLI Configuration\n"
	content += "# This file stores global configuration for the Tyk CLI\n\n"
	
	if cfg.DashURL != "" {
		content += fmt.Sprintf("dash_url = \"%s\"\n", cfg.DashURL)
	}
	if cfg.AuthToken != "" {
		content += fmt.Sprintf("auth_token = \"%s\"\n", cfg.AuthToken)
	}
	if cfg.OrgID != "" {
		content += fmt.Sprintf("org_id = \"%s\"\n", cfg.OrgID)
	}
	
	return content
}

func generateTOMLConfigWithEnvironments(cfg *types.Config) string {
	content := "# Tyk CLI Configuration\n"
	content += "# This file stores global configuration for the Tyk CLI\n\n"
	
	// If we have environments, use the environment-based structure
	if len(cfg.Environments) > 0 {
		// Set default environment
		if cfg.DefaultEnvironment != "" {
			content += fmt.Sprintf("default_environment = \"%s\"\n\n", cfg.DefaultEnvironment)
		}
		
		// Add environments section
		content += "[environments]\n"
		for name, env := range cfg.Environments {
			content += fmt.Sprintf("\n[environments.%s]\n", name)
			content += fmt.Sprintf("name = \"%s\"\n", env.Name)
			content += fmt.Sprintf("dash_url = \"%s\"\n", env.DashURL)
			content += fmt.Sprintf("auth_token = \"%s\"\n", env.AuthToken)
			content += fmt.Sprintf("org_id = \"%s\"\n", env.OrgID)
		}
	} else {
		// Backward compatibility: use direct config
		if cfg.DashURL != "" {
			content += fmt.Sprintf("dash_url = \"%s\"\n", cfg.DashURL)
		}
		if cfg.AuthToken != "" {
			content += fmt.Sprintf("auth_token = \"%s\"\n", cfg.AuthToken)
		}
		if cfg.OrgID != "" {
			content += fmt.Sprintf("org_id = \"%s\"\n", cfg.OrgID)
		}
	}
	
	return content
}

func maskToken(token string) string {
	if token == "" {
		return "(not set)"
	}
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "****" + token[len(token)-4:]
}