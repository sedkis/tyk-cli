package cli

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
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
    "golang.org/x/term"
    "gopkg.in/yaml.v3"
)

// truncateWithEllipsis shortens s to fit max characters, adding "..." when needed
func truncateWithEllipsis(s string, max int) string {
    if max <= 0 {
        return ""
    }
    if len(s) <= max {
        return s
    }
    if max <= 3 {
        // not enough room for meaningful content
        return s[:max]
    }
    return s[:max-3] + "..."
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// computeTableLayout returns column widths for ID/Name/Path and whether to use a stacked fallback.
func computeTableLayout(termWidth int) (idW, nameW, pathW int, stacked bool) {
    if termWidth < 20 {
        return 0, 0, 0, true
    }

    const sepWidth = 6 // two " | " separators
    contentWidth := termWidth - sepWidth
    if contentWidth < 15 {
        return 0, 0, 0, true
    }

    // Minimums and pleasant defaults
    minID, minName, minPath := 12, 14, 10
    idW, nameW, pathW = 16, 20, 14
    baseTotal := idW + nameW + pathW

    if contentWidth < (minID + minName + minPath) {
        return 0, 0, 0, true
    }

    if contentWidth < baseTotal {
        // shrink in order: id -> name -> path
        over := baseTotal - contentWidth
        shrink := func(cur *int, min int, want int) {
            if over <= 0 {
                return
            }
            can := *cur - min
            if can <= 0 {
                return
            }
            delta := want
            if delta > can {
                delta = can
            }
            if delta > over {
                delta = over
            }
            *cur -= delta
            over -= delta
        }

        shrink(&idW, minID, 8)
        shrink(&nameW, minName, 8)
        shrink(&pathW, minPath, 8)

        if over > 0 {
            return 0, 0, 0, true
        }
    } else if contentWidth > baseTotal {
        // distribute extra space: Name → ID → Path
        extra := contentWidth - baseTotal
        grow := func(cur *int, cap int) {
            if extra <= 0 {
                return
            }
            take := cap
            if take > extra {
                take = extra
            }
            *cur += take
            extra -= take
        }
        grow(&nameW, 30)
        grow(&idW, 20)
        grow(&pathW, 12)
    }

    return idW, nameW, pathW, false
}

func hideCursor(w io.Writer) { fmt.Fprint(w, "\x1b[?25l") }
func showCursor(w io.Writer) { fmt.Fprint(w, "\x1b[?25h") }

// readKey reads a single key or interprets ESC [ C/D as right/left arrows.
// It returns 'R' for right, 'L' for left, or the raw byte for other keys.
func readKey(r io.Reader) (byte, error) {
    buf := make([]byte, 1)
    if _, err := os.Stdin.Read(buf); err != nil { // use stdin directly (raw mode)
        return 0, err
    }
    b := buf[0]
    if b != 27 { // not ESC
        return b, nil
    }
    time.Sleep(2 * time.Millisecond)
    tail := make([]byte, 2)
    n, _ := os.Stdin.Read(tail)
    if n == 2 && tail[0] == '[' {
        switch tail[1] {
        case 'C':
            return 'R', nil // Right
        case 'D':
            return 'L', nil // Left
        }
    }
    return 27, nil // plain ESC
}

// alPrintf writes at column 0 for the current line to avoid drift from prior content.
func alPrintf(w io.Writer, format string, a ...interface{}) {
    fmt.Fprint(w, "\x1b[0G")
    fmt.Fprintf(w, format, a...)
}

// NewAPICommand creates the 'tyk api' command and its subcommands
func NewAPICommand() *cobra.Command {
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Manage OAS APIs",
		Long:  "Commands for managing OAS-native APIs in Tyk Dashboard",
	}

	// Add API subcommands
	apiCmd.AddCommand(NewAPIListCommand())
	apiCmd.AddCommand(NewAPIGetCommand())
	apiCmd.AddCommand(NewAPICreateCommand())
	apiCmd.AddCommand(NewAPIApplyCommand()) // New declarative upsert command
	apiCmd.AddCommand(NewAPIUpdateCommand())
	apiCmd.AddCommand(NewAPIDeleteCommand())
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
		Long: `Retrieve an OAS API by its ID, optionally specifying a version.

By default, returns the full API metadata including Tyk-specific extensions.
Use --oas-only to get a clean OpenAPI specification without Tyk extensions,
suitable for use with standard OpenAPI tooling.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runAPIGet,
	}

	cmd.Flags().String("version-name", "", "Specific version name to retrieve")
	cmd.Flags().Bool("oas-only", false, "Return only the OpenAPI specification without Tyk extensions")

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
		RunE: runAPICreate,
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
		RunE: runAPIApply,
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
		RunE: runAPIUpdate,
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

// NewAPIListCommand creates the 'tyk api list' command
func NewAPIListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OAS APIs",
		Long:  "List OAS APIs in the Dashboard, paginated with optional interactive navigation",
		RunE:  runAPIList,
	}

	cmd.Flags().Int("page", 1, "Page number (10 per page)")
	cmd.Flags().BoolP("interactive", "i", false, "Enable interactive pagination with arrow key navigation")

	return cmd
}

// runAPIList implements the 'tyk api list' command
func runAPIList(cmd *cobra.Command, args []string) error {
	page, _ := cmd.Flags().GetInt("page")
	interactive, _ := cmd.Flags().GetBool("interactive")
	
	if page <= 0 {
		page = 1
	}

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

	// Get output format from context
	outputFormat := GetOutputFormatFromContext(cmd.Context())

	// If interactive mode is requested, switch to interactive pagination
	if interactive {
		if outputFormat == types.OutputJSON {
			return fmt.Errorf("interactive mode is not compatible with JSON output format")
		}
		return runInteractiveAPIList(c, page)
	}

	// Non-interactive mode (existing behavior)
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apis, err := c.ListOASAPIs(ctx, page)
	if err != nil {
		return fmt.Errorf("failed to list APIs: %w", err)
	}

	if outputFormat == types.OutputJSON {
		payload := map[string]interface{}{
			"page":  page,
			"count": len(apis),
			"apis":  apis,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(payload)
	}

	// Human readable output
	displayAPIPage(apis, page, false)
	return nil
}

// displayAPIPage displays a page of APIs in a formatted table
func displayAPIPage(apis []*types.OASAPI, page int, interactive bool) {
	if len(apis) == 0 {
		if interactive {
			fmt.Fprintf(os.Stderr, "\033[2J\033[H")
			fmt.Fprintf(os.Stderr, "No APIs found on page %d.\n", page)
			fmt.Fprintf(os.Stderr, "\nNavigation:\n")
			fmt.Fprintf(os.Stderr, "  ← → or A D    Previous/Next page\n")
			fmt.Fprintf(os.Stderr, "  q or Ctrl+C   Quit\n")
			fmt.Fprintf(os.Stderr, "  r             Refresh current page\n")
			fmt.Fprintf(os.Stderr, "\nPress a key to navigate... ")
		} else {
			fmt.Fprintf(os.Stderr, "No APIs found on page %d.\n", page)
		}
		return
	}

    if interactive {
        // Clear screen and move cursor to home
        fmt.Fprintf(os.Stderr, "\033[2J\033[H")

        // Determine terminal width (fallback to 80)
        termWidth := 80
        if w, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil && w > 0 {
            termWidth = w
        }
        idW, nameW, pathW, stacked := computeTableLayout(termWidth)

        // Header sized to table width to avoid visual overflow on very wide terminals
        tableWidth := idW + nameW + pathW + 6 // columns + separators
        if tableWidth <= 0 || tableWidth > termWidth {
            tableWidth = termWidth
        }
        alPrintf(os.Stderr, "%s\n", strings.Repeat("=", tableWidth))
        color.New(color.FgBlue, color.Bold).Fprintf(os.Stderr, "APIs (page %d)\n", page)
        alPrintf(os.Stderr, "%s\n\n", strings.Repeat("=", tableWidth))

        if stacked {
            for _, api := range apis {
                alPrintf(os.Stderr, "ID: %s\n", truncateWithEllipsis(api.ID, 48))
                alPrintf(os.Stderr, "Name: %s\n", truncateWithEllipsis(api.Name, 48))
                alPrintf(os.Stderr, "Listen Path: %s\n", truncateWithEllipsis(api.ListenPath, 48))
                alPrintf(os.Stderr, "%s\n", strings.Repeat("-", 32))
            }
        } else {
            // Table header and divider with color
            hdr := color.New(color.FgCyan, color.Bold)
            dim := color.New(color.FgHiBlack)
            headerLine := fmt.Sprintf("%-*s | %-*s | %-*s", idW, "ID", nameW, "Name", pathW, "Listen Path")
            fmt.Fprint(os.Stderr, "\x1b[0G")
            hdr.Fprintln(os.Stderr, headerLine)
            dividerLine := fmt.Sprintf("%s | %s | %s", strings.Repeat("-", idW), strings.Repeat("-", nameW), strings.Repeat("-", pathW))
            fmt.Fprint(os.Stderr, "\x1b[0G")
            dim.Fprintln(os.Stderr, dividerLine)

            // Rows
            for _, api := range apis {
                id := truncateWithEllipsis(api.ID, idW)
                name := truncateWithEllipsis(api.Name, nameW)
                listenPath := truncateWithEllipsis(api.ListenPath, pathW)
                alPrintf(os.Stderr, "%-*s | %-*s | %-*s\n", idW, id, nameW, name, pathW, listenPath)
            }
        }

        dim := color.New(color.FgHiBlack)
        alPrintf(os.Stderr, "\n%s\n", strings.Repeat("=", tableWidth))
        fmt.Fprint(os.Stderr, "\x1b[0G")
        dim.Fprintln(os.Stderr, "Navigation: [←→ or AD] Next/Prev | [R] Refresh | [Q] Quit")
        alPrintf(os.Stderr, "%s\n", strings.Repeat("=", tableWidth))
        fmt.Fprint(os.Stderr, "\x1b[0G")
        dim.Fprint(os.Stderr, "Press a key to navigate... ")
    } else {
        // Non-interactive mode with colors
        blue := color.New(color.FgBlue, color.Bold)
        green := color.New(color.FgGreen, color.Bold)
		
		blue.Fprintf(os.Stderr, "APIs (page %d):\n", page)
		fmt.Fprintf(os.Stdout, "%-36s  %-28s  %-18s  %s\n", "ID", "Name", "Listen Path", "Default Version")
		fmt.Fprintf(os.Stdout, "%s\n", strings.Repeat("-", 36+2+28+2+18+2+16))
		for _, api := range apis {
			fmt.Fprintf(os.Stdout, "%-36s  %-28s  %-18s  %s\n", api.ID, api.Name, api.ListenPath, api.DefaultVersion)
		}
		green.Fprintf(os.Stderr, "\nUse '--page %d' for next page.\n", page+1)
	}
}

// runInteractiveAPIList handles the interactive pagination mode
func runInteractiveAPIList(c *client.Client, startPage int) error {
    // Make sure we're in a terminal that supports interactive input
    if !term.IsTerminal(int(os.Stdin.Fd())) {
        return fmt.Errorf("interactive mode requires a terminal")
    }

    // Put terminal in raw mode to capture individual keystrokes
    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        return fmt.Errorf("failed to enable raw terminal mode: %w", err)
    }
    defer func() {
        term.Restore(int(os.Stdin.Fd()), oldState)
        showCursor(os.Stderr)
    }()

    // Hide cursor during interactive repainting
    hideCursor(os.Stderr)

	currentPage := startPage
	
	for {
		// Create context with timeout for each API call
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		apis, err := c.ListOASAPIs(ctx, currentPage)
		cancel()
		
		if err != nil {
			return fmt.Errorf("failed to list APIs: %w", err)
		}

		// Display current page
		displayAPIPage(apis, currentPage, true)

        // Read a single keystroke (robust arrow handling)
        key, err := readKey(os.Stdin)
        if err != nil {
            return fmt.Errorf("failed to read input: %w", err)
        }

        switch key {
        case 'q', 'Q', 3: // 'q', 'Q', or Ctrl+C
            fmt.Fprintln(os.Stderr, "\nExiting...")
            return nil
        case 'r', 'R':
            // Refresh current page (continue loop)
            continue
        case 'a', 'A', 'L': // previous page
            if currentPage > 1 {
                currentPage--
            }
        case 'd', 'D': // next page
            // Next page - check if there are APIs on current page
            if len(apis) > 0 {
                currentPage++
            }
        default:
			// Ignore other keys
			continue
		}
	}
}

// runAPIGet implements the 'tyk api get' command
func runAPIGet(cmd *cobra.Command, args []string) error {
	apiID := args[0]
	versionName, _ := cmd.Flags().GetString("version-name")
	oasOnly, _ := cmd.Flags().GetBool("oas-only")

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
		return outputAPIAsJSON(api, oasOnly)
	}

	return outputAPIAsHuman(api, versionName, oasOnly)
}

// outputAPIAsJSON outputs the API in JSON format
func outputAPIAsJSON(api *types.OASAPI, oasOnly bool) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	
	if oasOnly && api.OAS != nil {
		// Strip the x-tyk-api-gateway extension and return only the OAS
		oasData := make(map[string]interface{})
		for key, value := range api.OAS {
			if key != "x-tyk-api-gateway" {
				oasData[key] = value
			}
		}
		return encoder.Encode(oasData)
	}
	
	return encoder.Encode(api)
}

// outputAPIAsHuman outputs the API in human-readable format
func outputAPIAsHuman(api *types.OASAPI, requestedVersion string, oasOnly bool) error {
	if api == nil {
		return fmt.Errorf("API data is nil")
	}

	blue := color.New(color.FgBlue, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow)

	// Skip API summary if OAS-only mode is requested
	if !oasOnly {
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
	}

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
			if !oasOnly {
				yellow.Fprintf(os.Stderr, "Warning: Version '%s' not found, showing main OAS document\n\n", requestedVersion)
			}
		}
	} else {
		// No specific version requested, show main OAS
		oasData = api.OAS
		versionToShow = "main"
	}

	if oasData != nil {
		// Strip x-tyk-api-gateway extension if OAS-only mode is requested
		if oasOnly {
			filteredOAS := make(map[string]interface{})
			for key, value := range oasData {
				if key != "x-tyk-api-gateway" {
					filteredOAS[key] = value
				}
			}
			oasData = filteredOAS
		} else {
			// Header to stderr (only in non-OAS-only mode)
			blue.Fprintf(os.Stderr, "OpenAPI Specification")
			if versionToShow != "main" {
				blue.Fprintf(os.Stderr, " (version: %s)", versionToShow)
			}
			blue.Fprintln(os.Stderr, ":")
		}

		// Convert to YAML for better readability and output to stdout
		yamlData, err := yaml.Marshal(oasData)
		if err != nil {
			return fmt.Errorf("failed to convert OAS to YAML: %w", err)
		}

		// Output YAML to stdout (no color for clean piping)
		fmt.Print(string(yamlData))
	} else {
		if !oasOnly {
			yellow.Fprintln(os.Stderr, "No OAS document available")
		}
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
		"api_id":          api.ID,
		"version_name":    versionName,
		"name":            api.Name,
		"listen_path":     api.ListenPath,
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
					Code:    2,
					Message: "Plain OAS document detected (missing x-tyk-api-gateway extensions). Use 'tyk api create' for plain OAS files, or add --create flag to apply",
				}
			} else {
				return &ExitError{
					Code:    2,
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
		"api_id":          api.ID,
		"version_name":    versionName,
		"name":            api.Name,
		"listen_path":     api.ListenPath,
		"default_version": api.DefaultVersion,
		"operation":       "updated",
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
