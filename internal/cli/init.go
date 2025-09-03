package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tyktech/tyk-cli/internal/client"
	"github.com/tyktech/tyk-cli/internal/config"
	"github.com/tyktech/tyk-cli/pkg/types"
)


// NewInitCommand creates the 'tyk init' command for guided setup
func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Interactive setup wizard for Tyk CLI",
		Long: `🚀 Interactive setup wizard to get you started with Tyk CLI quickly!

This wizard will help you:
- Configure your Tyk Dashboard connection
- Set up multiple environments (dev, staging, prod)
- Test your configuration
- Save everything for future use

Run this command to get started in minutes!`,
		RunE: runInitWizard,
	}

	cmd.Flags().Bool("skip-test", false, "Skip connection testing")
	cmd.Flags().Bool("quick", false, "Quick setup (single environment)")

	return cmd
}

func runInitWizard(cmd *cobra.Command, args []string) error {
	skipTest, _ := cmd.Flags().GetBool("skip-test")
	quickMode, _ := cmd.Flags().GetBool("quick")

	scanner := bufio.NewScanner(os.Stdin)

	printWelcome()
	
	if quickMode {
		return runQuickSetup(scanner, skipTest)
	}

	return runFullWizard(scanner, skipTest)
}

func printWelcome() {
	fmt.Println("🚀 Welcome to Tyk CLI Setup Wizard!")
	fmt.Println("====================================")
	fmt.Println()
	fmt.Println("This wizard will help you configure the Tyk CLI for your environment(s).")
	fmt.Println("You can set up multiple environments (dev, staging, production) and")
	fmt.Println("easily switch between them.")
	fmt.Println()
}

func runQuickSetup(scanner *bufio.Scanner, skipTest bool) error {
	fmt.Println("⚡ Quick Setup Mode")
	fmt.Println("------------------")
	fmt.Println()

	env, err := gatherEnvironmentInfo(scanner, "default", true)
	if err != nil {
		return err
	}

	if !skipTest {
		if err := testConnection(env); err != nil {
			fmt.Printf("⚠️  Connection test failed: %v\n", err)
			if !askYesNo(scanner, "Continue anyway?") {
				return fmt.Errorf("setup cancelled")
			}
		} else {
			fmt.Println("✅ Connection test successful!")
		}
	}

	if err := saveEnvironment(env, true); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	printSuccess("default")
	return nil
}

func runFullWizard(scanner *bufio.Scanner, skipTest bool) error {
	fmt.Println("🎯 Full Setup Wizard")
	fmt.Println("-------------------")
	fmt.Println()

	var environments []*types.Environment

	// Ask how many environments to set up
	fmt.Println("How many environments do you want to configure?")
	fmt.Println("1. Just one (development)")  
	fmt.Println("2. Two (development + production)")
	fmt.Println("3. Three (development + staging + production)")
	fmt.Println("4. Custom")
	fmt.Println()

	choice := askChoice(scanner, "Enter your choice (1-4)", []string{"1", "2", "3", "4"})
	
	var envNames []string
	switch choice {
	case "1":
		envNames = []string{"development"}
	case "2":
		envNames = []string{"development", "production"}
	case "3":
		envNames = []string{"development", "staging", "production"}
	case "4":
		envNames = askCustomEnvironments(scanner)
	}

	fmt.Printf("\n🔧 Setting up %d environment(s)...\n\n", len(envNames))

	for i, envName := range envNames {
		fmt.Printf("--- Environment %d/%d: %s ---\n", i+1, len(envNames), envName)
		
		env, err := gatherEnvironmentInfo(scanner, envName, i == 0)
		if err != nil {
			return err
		}
		
		if !skipTest {
			fmt.Printf("\n🔍 Testing connection to %s...\n", envName)
			if err := testConnection(env); err != nil {
				fmt.Printf("⚠️  Connection test failed: %v\n", err)
				if !askYesNo(scanner, "Continue with this environment anyway?") {
					continue
				}
			} else {
				fmt.Println("✅ Connection successful!")
			}
		}

		environments = append(environments, env)
		fmt.Println()
	}

	if len(environments) == 0 {
		return fmt.Errorf("no environments configured")
	}

	// Ask which environment should be active
	activeEnv := selectActiveEnvironment(scanner, environments)
	
	// Save all environments
	for _, env := range environments {
		isDefault := (env.Name == activeEnv)
		if err := saveEnvironment(env, isDefault); err != nil {
			return fmt.Errorf("failed to save %s environment: %w", env.Name, err)
		}
	}

	printSuccess(activeEnv)
	return nil
}

func gatherEnvironmentInfo(scanner *bufio.Scanner, envName string, isFirst bool) (*types.Environment, error) {
	env := &types.Environment{Name: envName}

	fmt.Printf("📝 Configuring '%s' environment:\n", envName)
	fmt.Println()

	// Gather Dashboard URL
	if isFirst {
		fmt.Println("Enter your Tyk Dashboard URL:")
		fmt.Println("Examples:")
		fmt.Println("  • http://localhost:3000 (local development)")
		fmt.Println("  • https://admin.cloud.tyk.io (Tyk Cloud)")
		fmt.Println("  • https://dashboard.yourcompany.com (self-hosted)")
		fmt.Println()
	}
	
	env.DashboardURL = askString(scanner, "Dashboard URL", "")
	if env.DashboardURL == "" {
		return nil, fmt.Errorf("dashboard URL is required")
	}

	// Gather Auth Token
	if isFirst {
		fmt.Println("\nEnter your Dashboard API Auth Token:")
		fmt.Println("💡 You can find this in your Tyk Dashboard under 'Users' → your user → 'API Access Credentials'")
		fmt.Println()
	}
	
	env.AuthToken = askString(scanner, "Auth Token", "")
	if env.AuthToken == "" {
		return nil, fmt.Errorf("auth token is required")
	}

	// Gather Org ID
	if isFirst {
		fmt.Println("\nEnter your Organization ID:")
		fmt.Println("💡 You can find this in your Dashboard URL or in the Dashboard under 'System Management'")
		fmt.Println()
	}
	
	env.OrgID = askString(scanner, "Organization ID", "")
	if env.OrgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	return env, nil
}

func askCustomEnvironments(scanner *bufio.Scanner) []string {
	var envNames []string
	
	fmt.Println("\nEnter environment names (one per line, empty line to finish):")
	
	for {
		name := askString(scanner, "Environment name", "")
		if name == "" {
			break
		}
		envNames = append(envNames, name)
	}
	
	if len(envNames) == 0 {
		envNames = []string{"development"} // Default fallback
	}
	
	return envNames
}

func selectActiveEnvironment(scanner *bufio.Scanner, environments []*types.Environment) string {
	if len(environments) == 1 {
		return environments[0].Name
	}

	fmt.Println("🎯 Which environment should be active by default?")
	for i, env := range environments {
		fmt.Printf("%d. %s\n", i+1, env.Name)
	}
	fmt.Println()

	choices := make([]string, len(environments))
	for i := range environments {
		choices[i] = fmt.Sprintf("%d", i+1)
	}

	choice := askChoice(scanner, "Select active environment", choices)
	idx := 0
	fmt.Sscanf(choice, "%d", &idx)
	
	return environments[idx-1].Name
}

func testConnection(env *types.Environment) error {
	config := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": env,
		},
	}

	client, err := client.NewClient(config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.Health(ctx)
}

func saveEnvironment(env *types.Environment, setAsGlobal bool) error {
	// Get config directory
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "cli.toml")
	
	// Create config manager and load existing config if it exists
	manager := config.NewManager()
	if _, err := os.Stat(configFile); err == nil {
		if err := manager.LoadConfig(); err != nil {
			return fmt.Errorf("failed to load existing config: %w", err)
		}
	}

	// Add the environment
	if err := manager.SaveEnvironment(env, setAsGlobal); err != nil {
		return err
	}

	// Generate and save the updated TOML config
	cfg := manager.GetConfig()
	content := generateTOMLConfigUnified(cfg)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
		return err
	}

	return nil
}

func printSuccess(activeEnv string) {
	fmt.Println("🎉 Setup Complete!")
	fmt.Println("==================")
	fmt.Println()
	fmt.Printf("✅ Active environment: %s\n", activeEnv)
	fmt.Println("✅ Configuration saved")
	fmt.Println()
	fmt.Println("🚀 You're ready to go! Try these commands:")
	fmt.Println("   tyk config get                    # View your configuration")
	fmt.Println("   tyk api --help                    # Explore API commands")
	fmt.Println("   tyk --version                     # Check CLI version")
	fmt.Println()
	fmt.Println("📚 For more help: tyk --help")
	fmt.Println()
}

func askString(scanner *bufio.Scanner, prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	
	if input == "" && defaultValue != "" {
		return defaultValue
	}
	
	return input
}

func askYesNo(scanner *bufio.Scanner, prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	scanner.Scan()
	input := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return input == "y" || input == "yes"
}

func askChoice(scanner *bufio.Scanner, prompt string, choices []string) string {
	for {
		fmt.Printf("%s: ", prompt)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
		
		for _, choice := range choices {
			if input == choice {
				return input
			}
		}
		
		fmt.Printf("Invalid choice. Please select from: %s\n", strings.Join(choices, ", "))
	}
}