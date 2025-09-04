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
	"github.com/tyktech/tyk-cli/internal/oas"
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
	apiCmd.AddCommand(NewAPIApplyCommand())  // New declarative upsert command
	apiCmd.AddCommand(NewAPIUpdateCommand())
	apiCmd.AddCommand(NewAPIDeleteCommand())
	apiCmd.AddCommand(NewAPIConvertCommand())
	// Note: Versioning commands moved to post-v0

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
		Short: "Create a new API from OAS file (explicit creation only)",
		Long: `Create a new OAS API from a local OpenAPI specification file.

This command ALWAYS creates a new API and generates a new API ID, 
ignoring any existing x-tyk-api-gateway.info.id in the file.

For declarative create/update based on ID presence, use 'tyk api apply'.`,
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

func NewAPIApplyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply API configuration (declarative upsert)",
		Long: `Apply API configuration from an OAS file with declarative upsert logic.

Behavior:
- If x-tyk-api-gateway.info.id is present in the file: UPDATE existing API
- If x-tyk-api-gateway.info.id is missing: ERROR (use --create or 'tyk api create')
- With --create flag and no ID: CREATE new API

This is the GitOps-friendly command for infrastructure-as-code workflows.`,
		RunE:  runAPIApply,
	}
	
	cmd.Flags().StringP("file", "f", "", "Path to OpenAPI specification file (required)")
	cmd.Flags().Bool("create", false, "Allow creation of new APIs when ID is missing")
	cmd.Flags().String("version-name", "", "Version name (defaults to info.version or v1)")
	cmd.Flags().Bool("set-default", true, "Set this version as the default")
	
	cmd.MarkFlagRequired("file")
	
	return cmd
}

func NewAPIUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing API (explicit update only)",
		Long: `Update an existing OAS API by replacing its OAS document.

Requires either --api-id flag or x-tyk-api-gateway.info.id in the file.
Always updates an existing API - never creates new ones.`,
		RunE:  runAPIUpdate,
	}
	
	cmd.Flags().String("api-id", "", "API ID to update (alternative: ID in file)")
	cmd.Flags().StringP("file", "f", "", "Path to OpenAPI specification file (required)")
	cmd.Flags().String("version-name", "", "Target version name")
	cmd.Flags().Bool("set-default", false, "Set this version as the default")
	
	cmd.MarkFlagRequired("file")
	
	return cmd
}

func NewAPIDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <api-id>",
		Short: "Delete an API by ID",
		Long:  "Delete an OAS API by its ID with confirmation prompt",
		Args:  cobra.ExactArgs(1),
		RunE:  runAPIDelete,
	}
	
	cmd.Flags().Bool("yes", false, "Skip confirmation prompt")
	
	return cmd
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
	yellow := color.New(color.FgYellow)
	
	// API Summary - output to stderr so stdout can be cleanly redirected
	blue.Fprintln(os.Stderr, "API Summary:")
	fmt.Fprintf(os.Stderr, "  ID:             %s\n", api.ID)
	fmt.Fprintf(os.Stderr, "  Name:           %s\n", api.Name)
	fmt.Fprintf(os.Stderr, "  Listen Path:    %s\n", api.ListenPath)
	fmt.Fprintf(os.Stderr, "  Default Version: ")
	green.Fprintf(os.Stderr, "%s\n", api.DefaultVersion)
	
	if api.CustomDomain != "" {
		fmt.Fprintf(os.Stderr, "  Custom Domain:  %s\n", api.CustomDomain)
	}
	if api.UpstreamURL != "" {
		fmt.Fprintf(os.Stderr, "  Upstream URL:   %s\n", api.UpstreamURL)
	}
	
	fmt.Fprintf(os.Stderr, "  Created:        %s\n", api.CreatedAt)
	fmt.Fprintf(os.Stderr, "  Updated:        %s\n", api.UpdatedAt)
	
	// Versions summary
	if len(api.VersionData) > 0 {
		fmt.Fprintln(os.Stderr)
		blue.Fprintln(os.Stderr, "Available Versions:")
		for versionName := range api.VersionData {
			marker := ""
			if versionName == api.DefaultVersion {
				marker = green.Sprint(" (default)")
			}
			fmt.Fprintf(os.Stderr, "  - %s%s\n", versionName, marker)
		}
	}
	
	fmt.Fprintln(os.Stderr)
	
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
			yellow.Fprintf(os.Stderr, "Warning: Version '%s' not found, showing main OAS document\n\n", requestedVersion)
		}
	} else {
		// No specific version requested, show main OAS
		oasData = api.OAS
		versionToShow = "main"
	}
	
	if oasData != nil {
		// Header to stderr
		blue.Fprintf(os.Stderr, "OpenAPI Specification")
		if versionToShow != "main" {
			blue.Fprintf(os.Stderr, " (version: %s)", versionToShow)
		}
		blue.Fprintln(os.Stderr, ":")
		
		// Convert to YAML for better readability and output to stdout
		yamlData, err := yaml.Marshal(oasData)
		if err != nil {
			return fmt.Errorf("failed to convert OAS to YAML: %w", err)
		}
		
		// Output YAML to stdout (no color for clean piping)
		fmt.Print(string(yamlData))
	} else {
		yellow.Fprintln(os.Stderr, "No OAS document available")
	}
	
	return nil
}

// runAPICreate implements the 'tyk api create' command
func runAPICreate(cmd *cobra.Command, args []string) error {
	// Get flags
	filePath, _ := cmd.Flags().GetString("file")
	versionName, _ := cmd.Flags().GetString("version-name")
	// TODO: Handle set-default flag when sending raw OAS document
	
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
	
	// Auto-generate x-tyk-api-gateway extensions for plain OAS documents
	if !oas.HasTykExtensions(oasData) {
		oasData, err = oas.AddTykExtensions(oasData)
		if err != nil {
			return &ExitError{Code: 2, Message: fmt.Sprintf("failed to generate Tyk extensions: %v", err)}
		}
	}
	
	// Strip any existing API ID from OAS file (create always generates new ID)
	oasData = stripExistingAPIID(oasData)
	
	// Extract version name from OAS if not provided
	if versionName == "" {
		versionName = extractVersionFromOAS(oasData)
		if versionName == "" {
			versionName = "v1" // fallback
		}
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
	api, err := c.CreateOASAPI(ctx, oasData)
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

// stripExistingAPIID removes x-tyk-api-gateway.info.id from OAS document
// This ensures create command always generates new ID
func stripExistingAPIID(oasData map[string]interface{}) map[string]interface{} {
	if xTyk, exists := oasData["x-tyk-api-gateway"]; exists {
		if xTykMap, ok := xTyk.(map[string]interface{}); ok {
			if info, exists := xTykMap["info"]; exists {
				if infoMap, ok := info.(map[string]interface{}); ok {
					delete(infoMap, "id") // Remove existing API ID
				}
			}
		}
	}
	return oasData
}

// extractAPIIDFromOAS extracts API ID from x-tyk-api-gateway.info.id
func extractAPIIDFromOAS(oasData map[string]interface{}) (string, bool) {
	if xTyk, exists := oasData["x-tyk-api-gateway"]; exists {
		if xTykMap, ok := xTyk.(map[string]interface{}); ok {
			if info, exists := xTykMap["info"]; exists {
				if infoMap, ok := info.(map[string]interface{}); ok {
					if id, exists := infoMap["id"]; exists {
						if idStr, ok := id.(string); ok && idStr != "" {
							return idStr, true
						}
					}
				}
			}
		}
	}
	return "", false
}

// runAPIApply implements the 'tyk api apply' command (declarative upsert)
func runAPIApply(cmd *cobra.Command, args []string) error {
	// Get flags
	filePath, _ := cmd.Flags().GetString("file")
	allowCreate, _ := cmd.Flags().GetBool("create")
	versionName, _ := cmd.Flags().GetString("version-name")
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
	
	// Check for existing API ID in the file
	apiID, hasID := oas.ExtractAPIIDFromTykExtensions(oasData)
	
	if hasID {
		// API ID present - update existing API
		return updateExistingAPI(cmd, config, apiID, oasData, versionName, setDefault)
	} else {
		// No API ID present
		if !allowCreate {
			// Check if it's a plain OAS document
			if !oas.HasTykExtensions(oasData) {
				return &ExitError{
					Code: 2, 
					Message: "Plain OAS document detected (missing x-tyk-api-gateway extensions). Use 'tyk api create' for plain OAS files, or add --create flag to apply",
				}
			} else {
				return &ExitError{
					Code: 2, 
					Message: "API ID not found in x-tyk-api-gateway.info.id. Use 'tyk api create' or add --create flag to apply",
				}
			}
		}
		
		// Create new API via apply
		return createNewAPIViaApply(cmd, config, oasData, versionName, setDefault)
	}
}

// updateExistingAPI handles updating an existing API via apply
func updateExistingAPI(cmd *cobra.Command, config *types.Config, apiID string, oasData map[string]interface{}, versionName string, setDefault bool) error {
	// Create client
	c, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Check if API exists first
	_, err = c.GetOASAPI(ctx, apiID, "")
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return &ExitError{Code: 3, Message: fmt.Sprintf("API with ID '%s' not found. Verify the API exists or use 'tyk api create'", apiID)}
		}
		return fmt.Errorf("failed to verify API exists: %w", err)
	}
	
	// Extract version name from OAS if not provided
	if versionName == "" {
		versionName = extractVersionFromOAS(oasData)
		if versionName == "" {
			versionName = "v1" // fallback
		}
	}
	
	// Update the API
	api, err := c.UpdateOASAPI(ctx, apiID, oasData)
	if err != nil {
		return fmt.Errorf("failed to update API: %w", err)
	}
	
	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())
	
	if outputFormat == types.OutputJSON {
		return outputUpdatedAPIAsJSON(api, versionName)
	}
	
	return outputUpdatedAPIAsHuman(api, versionName)
}

// createNewAPIViaApply handles creating a new API via apply --create
func createNewAPIViaApply(cmd *cobra.Command, config *types.Config, oasData map[string]interface{}, versionName string, setDefault bool) error {
	// Auto-generate x-tyk-api-gateway extensions for plain OAS documents
	if !oas.HasTykExtensions(oasData) {
		var err error
		oasData, err = oas.AddTykExtensions(oasData)
		if err != nil {
			return &ExitError{Code: 2, Message: fmt.Sprintf("failed to generate Tyk extensions: %v", err)}
		}
	}
	
	// Strip any existing ID (shouldn't be there, but be safe)
	oasData = stripExistingAPIID(oasData)
	
	// Extract version name from OAS if not provided
	if versionName == "" {
		versionName = extractVersionFromOAS(oasData)
		if versionName == "" {
			versionName = "v1" // fallback
		}
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
	api, err := c.CreateOASAPI(ctx, oasData)
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

// runAPIUpdate implements the 'tyk api update' command (explicit update)
func runAPIUpdate(cmd *cobra.Command, args []string) error {
	// Get flags
	apiIDFlag, _ := cmd.Flags().GetString("api-id")
	filePath, _ := cmd.Flags().GetString("file")
	versionName, _ := cmd.Flags().GetString("version-name")
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
	
	// Determine API ID to use
	var apiID string
	if apiIDFlag != "" {
		apiID = apiIDFlag
	} else {
		// Try to extract from file
		if id, hasID := extractAPIIDFromOAS(oasData); hasID {
			apiID = id
		} else {
			return &ExitError{Code: 2, Message: "Missing required API ID. Use --api-id flag or ensure x-tyk-api-gateway.info.id is set in file"}
		}
	}
	
	return updateExistingAPI(cmd, config, apiID, oasData, versionName, setDefault)
}

// runAPIDelete implements the 'tyk api delete' command
func runAPIDelete(cmd *cobra.Command, args []string) error {
	apiID := args[0]
	skipConfirmation, _ := cmd.Flags().GetBool("yes")
	
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
	
	// Verify API exists first
	api, err := c.GetOASAPI(ctx, apiID, "")
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return &ExitError{Code: 3, Message: fmt.Sprintf("API '%s' not found", apiID)}
		}
		return fmt.Errorf("failed to verify API exists: %w", err)
	}
	
	// Confirmation prompt unless --yes flag is provided
	if !skipConfirmation {
		fmt.Printf("Are you sure you want to delete API '%s' (%s)? [y/N]: ", apiID, api.Name)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Delete operation cancelled")
			return nil
		}
	}
	
	// Delete the API
	err = c.DeleteOASAPI(ctx, apiID)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return &ExitError{Code: 3, Message: fmt.Sprintf("API '%s' not found", apiID)}
		}
		return fmt.Errorf("failed to delete API: %w", err)
	}
	
	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())
	
	if outputFormat == types.OutputJSON {
		return outputDeletedAPIAsJSON(apiID)
	}
	
	return outputDeletedAPIAsHuman(apiID, api.Name)
}

// outputUpdatedAPIAsJSON outputs the updated API result in JSON format
func outputUpdatedAPIAsJSON(api *types.OASAPI, versionName string) error {
	result := map[string]interface{}{
		"api_id":       api.ID,
		"version_name": versionName,
		"name":         api.Name,
		"listen_path":  api.ListenPath,
		"default_version": api.DefaultVersion,
		"operation":    "updated",
	}
	
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// outputUpdatedAPIAsHuman outputs the updated API result in human-readable format
func outputUpdatedAPIAsHuman(api *types.OASAPI, versionName string) error {
	green := color.New(color.FgGreen, color.Bold)
	blue := color.New(color.FgBlue, color.Bold)
	
	green.Println("✓ API updated successfully!")
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

// outputDeletedAPIAsJSON outputs the deleted API result in JSON format
func outputDeletedAPIAsJSON(apiID string) error {
	result := map[string]interface{}{
		"api_id":    apiID,
		"operation": "deleted",
		"success":   true,
	}
	
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// outputDeletedAPIAsHuman outputs the deleted API result in human-readable format
func outputDeletedAPIAsHuman(apiID, apiName string) error {
	green := color.New(color.FgGreen, color.Bold)
	
	green.Printf("✓ Deleted API '%s'\n", apiID)
	if apiName != "" {
		fmt.Printf("  Name: %s\n", apiName)
	}
	
	return nil
}

