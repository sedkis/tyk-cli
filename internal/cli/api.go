package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tyktech/tyk-cli/internal/client"
	"github.com/tyktech/tyk-cli/internal/filehandler"
	"github.com/tyktech/tyk-cli/pkg/types"
	"gopkg.in/yaml.v3"
)

// NewAPICommand creates the 'tyk api' command and its subcommands
func NewAPICommand() *cobra.Command {
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Manage OAS APIs",
		Long:  "Commands for managing OAS-native APIs in Tyk Dashboard",
	}

	// Add API subcommands
	apiCmd.AddCommand(NewAPIGetCommand())
	apiCmd.AddCommand(NewAPICreateCommand())
	apiCmd.AddCommand(NewAPIImportCommand())
	apiCmd.AddCommand(NewAPIUpdateCommand())
	apiCmd.AddCommand(NewAPIDeleteCommand())
	apiCmd.AddCommand(NewAPIConvertCommand())
	apiCmd.AddCommand(NewAPIVersionsCommand())

	return apiCmd
}

// NewAPIVersionsCommand creates the 'tyk api versions' command and its subcommands
func NewAPIVersionsCommand() *cobra.Command {
	versionsCmd := &cobra.Command{
		Use:   "versions",
		Short: "Manage API versions",
		Long:  "Commands for managing versions of OAS APIs",
	}

	// Add version subcommands
	versionsCmd.AddCommand(NewAPIVersionsListCommand())
	versionsCmd.AddCommand(NewAPIVersionsCreateCommand())
	versionsCmd.AddCommand(NewAPIVersionsSwitchDefaultCommand())

	return versionsCmd
}

// Placeholder functions for version commands - these will be implemented in phase 3

func NewAPIVersionsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List versions for an API",
		Long:  "List all versions for a given API ID",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("API versions list command will be implemented in phase 3")
		},
	}
}

func NewAPIVersionsCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new API version",
		Long:  "Create a new version for an existing API",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("API versions create command will be implemented in phase 3")
		},
	}
}

func NewAPIVersionsSwitchDefaultCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "switch-default",
		Short: "Switch default version for an API",
		Long:  "Switch the default version for a given API",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("API versions switch-default command will be implemented in phase 3")
		},
	}
}

// Placeholder functions for API commands - these will be implemented in the next phases

func NewAPIGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <api-id>",
		Short: "Get an API by ID",
		Long:  "Retrieve an OAS API by its ID, optionally specifying a version",
		Args:  cobra.ExactArgs(1),
		RunE:  runAPIGet,
	}
	
	cmd.Flags().String("version-name", "", "Specific version name to retrieve")
	
	return cmd
}

func NewAPICreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API from OAS file",
		Long:  "Create a new OAS API from a local OpenAPI specification file",
		RunE:  runAPICreate,
	}
	
	cmd.Flags().StringP("file", "f", "", "Path to OpenAPI specification file (required)")
	cmd.Flags().String("version-name", "", "Version name for the API (defaults to info.version or v1)")
	cmd.Flags().String("upstream-url", "", "Upstream URL for the API")
	cmd.Flags().String("listen-path", "", "Listen path for the API")
	cmd.Flags().String("custom-domain", "", "Custom domain for the API")
	cmd.Flags().Bool("set-default", true, "Set this version as the default")
	
	cmd.MarkFlagRequired("file")
	
	return cmd
}

func NewAPIImportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Import an API (create or update)",
		Long:  "Create or update an API from a local OAS file with explicit mode selection",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("API import command will be implemented in phase 2")
		},
	}
}

func NewAPIUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update an existing API",
		Long:  "Replace the OAS document of an existing API/version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("API update command will be implemented in phase 2")
		},
	}
}

func NewAPIDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete an API by ID",
		Long:  "Delete an OAS API by its ID",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("API delete command will be implemented in phase 2")
		},
	}
}

func NewAPIConvertCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "convert",
		Short: "Convert OAS to Tyk API definition",
		Long:  "Convert a local OAS file to Tyk API definition format (local operation)",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("API convert command will be implemented in phase 2")
		},
	}
}

// runAPIGet implements the 'tyk api get' command
func runAPIGet(cmd *cobra.Command, args []string) error {
	apiID := args[0]
	versionName, _ := cmd.Flags().GetString("version-name")
	
	// Get configuration from context
	config := GetConfigFromContext(cmd.Context())
	if config == nil {
		return fmt.Errorf("configuration not found")
	}
	
	// Create client
	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Get the API
	api, err := c.GetOASAPI(ctx, apiID, versionName)
	if err != nil {
		// Check if it's a not found error
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return &ExitError{Code: 3, Message: fmt.Sprintf("API '%s' not found", apiID)}
		}
		return fmt.Errorf("failed to get API: %w", err)
	}
	
	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())
	
	if outputFormat == types.OutputJSON {
		return outputAPIAsJSON(api)
	}
	
	return outputAPIAsHuman(api, versionName)
}

// outputAPIAsJSON outputs the API in JSON format
func outputAPIAsJSON(api *types.OASAPI) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(api)
}

// outputAPIAsHuman outputs the API in human-readable format
func outputAPIAsHuman(api *types.OASAPI, requestedVersion string) error {
	if api == nil {
		return fmt.Errorf("API data is nil")
	}
	
	blue := color.New(color.FgBlue, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	
	// API Summary
	blue.Println("API Summary:")
	fmt.Printf("  ID:             %s\n", api.ID)
	fmt.Printf("  Name:           %s\n", api.Name)
	fmt.Printf("  Listen Path:    %s\n", api.ListenPath)
	fmt.Printf("  Default Version: ")
	green.Printf("%s\n", api.DefaultVersion)
	
	if api.CustomDomain != "" {
		fmt.Printf("  Custom Domain:  %s\n", api.CustomDomain)
	}
	if api.UpstreamURL != "" {
		fmt.Printf("  Upstream URL:   %s\n", api.UpstreamURL)
	}
	
	fmt.Printf("  Created:        %s\n", api.CreatedAt)
	fmt.Printf("  Updated:        %s\n", api.UpdatedAt)
	
	// Versions summary
	if len(api.VersionData) > 0 {
		fmt.Println()
		blue.Println("Available Versions:")
		for versionName := range api.VersionData {
			marker := ""
			if versionName == api.DefaultVersion {
				marker = green.Sprint(" (default)")
			}
			fmt.Printf("  - %s%s\n", versionName, marker)
		}
	}
	
	fmt.Println()
	
	// Determine which OAS to show
	var oasData map[string]interface{}
	var versionToShow string
	
	if requestedVersion != "" {
		// Show specific version if requested and exists
		if versionData, exists := api.VersionData[requestedVersion]; exists && versionData.OAS != nil {
			oasData = versionData.OAS
			versionToShow = requestedVersion
		} else if api.OAS != nil {
			// Fallback to main OAS if version not found
			oasData = api.OAS
			versionToShow = "main"
			yellow.Printf("Warning: Version '%s' not found, showing main OAS document\n\n", requestedVersion)
		}
	} else {
		// No specific version requested, show main OAS
		oasData = api.OAS
		versionToShow = "main"
	}
	
	if oasData != nil {
		blue.Printf("OpenAPI Specification")
		if versionToShow != "main" {
			blue.Printf(" (version: %s)", versionToShow)
		}
		blue.Println(":")
		
		// Convert to YAML for better readability
		yamlData, err := yaml.Marshal(oasData)
		if err != nil {
			return fmt.Errorf("failed to convert OAS to YAML: %w", err)
		}
		
		cyan.Print(string(yamlData))
	} else {
		yellow.Println("No OAS document available")
	}
	
	return nil
}

// runAPICreate implements the 'tyk api create' command
func runAPICreate(cmd *cobra.Command, args []string) error {
	// Get flags
	filePath, _ := cmd.Flags().GetString("file")
	versionName, _ := cmd.Flags().GetString("version-name")
	upstreamURL, _ := cmd.Flags().GetString("upstream-url")
	listenPath, _ := cmd.Flags().GetString("listen-path")
	customDomain, _ := cmd.Flags().GetString("custom-domain")
	setDefault, _ := cmd.Flags().GetBool("set-default")
	
	// Get configuration from context
	config := GetConfigFromContext(cmd.Context())
	if config == nil {
		return fmt.Errorf("configuration not found")
	}
	
	// Validate and read the OAS file
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return &ExitError{Code: 2, Message: fmt.Sprintf("failed to resolve file path: %v", err)}
		}
		filePath = absPath
	}
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &ExitError{Code: 2, Message: fmt.Sprintf("file not found: %s", filePath)}
	}
	
	// Load and parse the OAS file
	fileInfo, err := filehandler.LoadFile(filePath)
	if err != nil {
		return &ExitError{Code: 2, Message: fmt.Sprintf("failed to load OAS file: %v", err)}
	}
	oasData := fileInfo.Content
	
	// Extract version name from OAS if not provided
	if versionName == "" {
		versionName = extractVersionFromOAS(oasData)
		if versionName == "" {
			versionName = "v1" // fallback
		}
	}
	
	// Convert OAS data to JSON for API request
	oasJSON, err := json.Marshal(oasData)
	if err != nil {
		return fmt.Errorf("failed to marshal OAS data: %w", err)
	}
	
	// Create the API request
	createReq := &types.CreateOASAPIRequest{
		OAS:            json.RawMessage(oasJSON),
		NewVersionName: versionName,
		SetDefault:     setDefault,
		UpstreamURL:    upstreamURL,
		ListenPath:     listenPath,
		CustomDomain:   customDomain,
	}
	
	// Create client
	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Create the API
	api, err := c.CreateOASAPI(ctx, createReq)
	if err != nil {
		// Check for conflict errors
		if strings.Contains(err.Error(), "409") || strings.Contains(err.Error(), "conflict") {
			return &ExitError{Code: 4, Message: fmt.Sprintf("API creation failed due to conflict: %v", err)}
		}
		return fmt.Errorf("failed to create API: %w", err)
	}
	
	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())
	
	if outputFormat == types.OutputJSON {
		return outputCreatedAPIAsJSON(api, versionName)
	}
	
	return outputCreatedAPIAsHuman(api, versionName)
}

// extractVersionFromOAS extracts version from OAS info.version field
func extractVersionFromOAS(oasData map[string]interface{}) string {
	if info, ok := oasData["info"].(map[string]interface{}); ok {
		if version, ok := info["version"].(string); ok && version != "" {
			return version
		}
	}
	return ""
}

// outputCreatedAPIAsJSON outputs the created API result in JSON format
func outputCreatedAPIAsJSON(api *types.OASAPI, versionName string) error {
	result := map[string]interface{}{
		"api_id":       api.ID,
		"version_name": versionName,
		"name":         api.Name,
		"listen_path":  api.ListenPath,
		"default_version": api.DefaultVersion,
	}
	
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// outputCreatedAPIAsHuman outputs the created API result in human-readable format
func outputCreatedAPIAsHuman(api *types.OASAPI, versionName string) error {
	green := color.New(color.FgGreen, color.Bold)
	blue := color.New(color.FgBlue, color.Bold)
	
	green.Println("✓ API created successfully!")
	fmt.Printf("  API ID:         %s\n", api.ID)
	fmt.Printf("  Name:           %s\n", api.Name)
	fmt.Printf("  Version:        %s\n", versionName)
	fmt.Printf("  Listen Path:    %s\n", api.ListenPath)
	
	if api.CustomDomain != "" {
		fmt.Printf("  Custom Domain:  %s\n", api.CustomDomain)
	}
	if api.UpstreamURL != "" {
		fmt.Printf("  Upstream URL:   %s\n", api.UpstreamURL)
	}
	
	blue.Printf("  Default Version: %s\n", api.DefaultVersion)
	
	return nil
}

