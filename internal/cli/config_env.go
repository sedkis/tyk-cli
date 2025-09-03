package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tyktech/tyk-cli/internal/config"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// NewConfigEnvCommand creates the 'tyk config env' command for environment management
func NewConfigEnvCommand() *cobra.Command {
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage configuration environments",
		Long:  "Commands for managing multiple configuration environments",
	}

	envCmd.AddCommand(NewConfigEnvListCommand())
	envCmd.AddCommand(NewConfigEnvAddCommand())
	envCmd.AddCommand(NewConfigEnvSwitchCommand())
	envCmd.AddCommand(NewConfigEnvRemoveCommand())

	return envCmd
}

// NewConfigEnvListCommand creates the 'tyk config env list' command
func NewConfigEnvListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured environments",
		Long:  "Display all configured environments and show which one is active",
		RunE:  runConfigEnvList,
	}

	return cmd
}

// NewConfigEnvAddCommand creates the 'tyk config env add' command
func NewConfigEnvAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [environment-name]",
		Short: "Add a new environment",
		Long: `Add a new environment configuration.

Examples:
  tyk config env add development --dash-url http://localhost:3000
  tyk config env add production --dash-url https://prod-dashboard.com --set-default`,
		Args: cobra.ExactArgs(1),
		RunE: runConfigEnvAdd,
	}

	cmd.Flags().String("dash-url", "", "Tyk Dashboard URL")
	cmd.Flags().String("auth-token", "", "Dashboard API auth token")
	cmd.Flags().String("org-id", "", "Organization ID")
	cmd.Flags().Bool("set-default", false, "Set this environment as the default")

	cmd.MarkFlagRequired("dash-url")
	cmd.MarkFlagRequired("auth-token")
	cmd.MarkFlagRequired("org-id")

	return cmd
}

// NewConfigEnvSwitchCommand creates the 'tyk config env switch' command
func NewConfigEnvSwitchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch [environment-name]",
		Short: "Switch to a different environment",
		Long:  "Change the active/default environment. If no environment name is provided, you'll be prompted to select from available environments.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runConfigEnvSwitch,
	}

	return cmd
}

// NewConfigEnvRemoveCommand creates the 'tyk config env remove' command
func NewConfigEnvRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [environment-name]",
		Short: "Remove an environment",
		Long:  "Remove an environment from the configuration",
		Args:  cobra.ExactArgs(1),
		RunE:  runConfigEnvRemove,
	}

	return cmd
}

func runConfigEnvList(cmd *cobra.Command, args []string) error {
	manager := config.NewManager()
	if err := manager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()
	environments := manager.ListEnvironments()

	if len(environments) == 0 {
		yellow := color.New(color.FgYellow)
		yellow.Println("No environments configured.")
		fmt.Println("Use 'tyk config env add' to add an environment.")
		return nil
	}

	blue := color.New(color.FgBlue, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)
	
	blue.Printf("Default environment: ")
	green.Printf("%s\n\n", cfg.DefaultEnvironment)
	
	blue.Println("Environments:")

	// Sort environment names for consistent display
	var envNames []string
	for name := range environments {
		envNames = append(envNames, name)
	}
	sort.Strings(envNames)

	for _, name := range envNames {
		env := environments[name]
		if name == cfg.DefaultEnvironment {
			green.Printf("● %s (active):\n", name)
		} else {
			fmt.Printf("  %s:\n", name)
		}
		cyan.Printf("    dash_url   = %s\n", env.DashURL)
		cyan.Printf("    auth_token = %s\n", maskToken(env.AuthToken))
		cyan.Printf("    org_id     = %s\n", env.OrgID)
		fmt.Println()
	}

	return nil
}

func runConfigEnvAdd(cmd *cobra.Command, args []string) error {
	envName := args[0]
	dashURL, _ := cmd.Flags().GetString("dash-url")
	authToken, _ := cmd.Flags().GetString("auth-token")
	orgID, _ := cmd.Flags().GetString("org-id")
	setDefault, _ := cmd.Flags().GetBool("set-default")

	// Create the environment
	env := &types.Environment{
		Name:      envName,
		DashURL:   dashURL,
		AuthToken: authToken,
		OrgID:     orgID,
	}

	// Validate the environment
	if err := env.Validate(); err != nil {
		return err
	}

	// Load existing configuration
	manager := config.NewManager()
	if err := manager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Save the environment
	if err := manager.SaveEnvironment(env, setDefault); err != nil {
		return fmt.Errorf("failed to save environment: %w", err)
	}

	// Save to file
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	cfg := manager.GetConfig()
	content := generateTOMLConfigWithEnvironments(cfg)

	configFile := fmt.Sprintf("%s/cli.toml", configDir)
	if err := saveConfigToFile(configFile, content); err != nil {
		return err
	}

	green := color.New(color.FgGreen, color.Bold)
	green.Printf("✓ Environment '%s' added successfully.\n", envName)
	if setDefault {
		green.Printf("✓ Environment '%s' set as default.\n", envName)
	}

	return nil
}

func runConfigEnvSwitch(cmd *cobra.Command, args []string) error {
	manager := config.NewManager()
	if err := manager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()
	environments := manager.ListEnvironments()

	if len(environments) == 0 {
		return fmt.Errorf("no environments configured. Use 'tyk config env add' to add an environment")
	}

	var envName string
	
	// If environment name was provided as argument, use it
	if len(args) > 0 {
		envName = args[0]
	} else {
		// Interactive selection
		var err error
		envName, err = selectEnvironmentInteractively(environments, cfg.DefaultEnvironment)
		if err != nil {
			return err
		}
		if envName == "" {
			return fmt.Errorf("no environment selected")
		}
	}

	// Check if environment exists
	if _, err := manager.GetEnvironment(envName); err != nil {
		return err
	}

	// Set as default
	if err := manager.SetDefaultEnvironment(envName); err != nil {
		return err
	}

	// Save to file
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	cfg = manager.GetConfig()
	content := generateTOMLConfigWithEnvironments(cfg)

	configFile := fmt.Sprintf("%s/cli.toml", configDir)
	if err := saveConfigToFile(configFile, content); err != nil {
		return err
	}

	green := color.New(color.FgGreen, color.Bold)
	green.Printf("✓ Switched to environment '%s'.\n", envName)
	return nil
}

func runConfigEnvRemove(cmd *cobra.Command, args []string) error {
	envName := args[0]

	manager := config.NewManager()
	if err := manager.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()

	// Check if environment exists
	if cfg.Environments[envName] == nil {
		return fmt.Errorf("environment '%s' not found", envName)
	}

	// Don't allow removing the last environment
	if len(cfg.Environments) == 1 {
		return fmt.Errorf("cannot remove the last environment")
	}

	// Remove environment
	delete(cfg.Environments, envName)

	// If this was the default, pick another one
	if cfg.DefaultEnvironment == envName {
		for name := range cfg.Environments {
			cfg.DefaultEnvironment = name
			break
		}
		yellow := color.New(color.FgYellow)
		yellow.Printf("⚠ Default environment changed to '%s'.\n", cfg.DefaultEnvironment)
	}

	// Save to file
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	content := generateTOMLConfigWithEnvironments(cfg)

	configFile := fmt.Sprintf("%s/cli.toml", configDir)
	if err := saveConfigToFile(configFile, content); err != nil {
		return err
	}

	green := color.New(color.FgGreen, color.Bold)
	green.Printf("✓ Environment '%s' removed successfully.\n", envName)
	return nil
}

func saveConfigToFile(configFile, content string) error {
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func selectEnvironmentInteractively(environments map[string]*types.Environment, currentDefault string) (string, error) {
	// Create sorted list of environment names for consistent display
	var envNames []string
	for name := range environments {
		envNames = append(envNames, name)
	}
	sort.Strings(envNames)

	// Create display options with colors and current default indicator
	blue := color.New(color.FgBlue, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	
	var options []string
	for _, name := range envNames {
		env := environments[name]
		var displayName string
		if name == currentDefault {
			displayName = green.Sprintf("● %s (current)", name)
		} else {
			displayName = fmt.Sprintf("  %s", name)
		}
		
		// Add environment details
		displayName += yellow.Sprintf(" - %s", env.DashURL)
		options = append(options, displayName)
	}

	// Create the interactive prompt
	prompt := &survey.Select{
		Message: blue.Sprint("Select environment to switch to:"),
		Options: options,
		Help:    "Use arrow keys to navigate, Enter to select, Ctrl+C to cancel",
	}

	var selectedIndex int
	if err := survey.AskOne(prompt, &selectedIndex); err != nil {
		return "", fmt.Errorf("environment selection cancelled: %w", err)
	}

	return envNames[selectedIndex], nil
}