package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/tyktech/tyk-cli/pkg/types"
)

const (
	// API endpoints
	OASAPIsPath        = "/api/apis/oas"
	OASAPIPath         = "/api/apis/oas/%s"          // {apiId}
	OASAPIVersionsPath = "/api/apis/oas/%s/versions" // {apiId}

	// Default timeout
	DefaultTimeout = 30 * time.Second

	// Headers
	HeaderAuthorization = "authorization"
	HeaderContentType   = "content-type"
	HeaderAccept        = "accept"

	// Content types
	ContentTypeJSON = "application/json"
	ContentTypeYAML = "application/x-yaml"
)

// Client represents a Tyk Dashboard API client
type Client struct {
	config     *types.Config
	httpClient *http.Client
	baseURL    *url.URL
}

// NewClient creates a new Tyk Dashboard API client
func NewClient(config *types.Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Get the active environment
	activeEnv, err := config.GetActiveEnvironment()
	if err != nil {
		return nil, fmt.Errorf("no active environment: %w", err)
	}

	baseURL, err := url.Parse(activeEnv.DashboardURL)
	if err != nil {
		return nil, fmt.Errorf("invalid dashboard URL: %w", err)
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL: baseURL,
	}, nil
}

// SetTimeout sets the HTTP client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// doRequest performs an HTTP request with proper headers and error handling
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	var contentType string

	if body != nil {
		switch v := body.(type) {
		case []byte:
			reqBody = bytes.NewReader(v)
			contentType = ContentTypeJSON
		case string:
			reqBody = bytes.NewReader([]byte(v))
			contentType = ContentTypeJSON
		default:
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewReader(jsonBody)
			contentType = ContentTypeJSON
		}
	}

	// Build URL
	fullURL := *c.baseURL
	// Support optional query string embedded in path
	if u, err := url.Parse(path); err == nil {
		fullURL.Path = u.Path
		fullURL.RawQuery = u.RawQuery
	} else {
		fullURL.Path = path
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get active environment for auth token
	activeEnv, err := c.config.GetActiveEnvironment()
	if err != nil {
		return nil, fmt.Errorf("no active environment for auth: %w", err)
	}

	// Set headers
	req.Header.Set(HeaderAuthorization, activeEnv.AuthToken)
	req.Header.Set(HeaderAccept, ContentTypeJSON)
	if contentType != "" {
		req.Header.Set(HeaderContentType, contentType)
	}

	return c.httpClient.Do(req)
}

// handleResponse processes HTTP response and handles errors
func (c *Client) handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error status codes
	if resp.StatusCode >= 400 {
		var errorResp types.ErrorResponse
		errorResp.Status = resp.StatusCode
		errorResp.Message = string(body)

		// Try to parse as JSON error response
		if err := json.Unmarshal(body, &errorResp); err != nil {
			// If not JSON, use status text and body as message
			errorResp.Message = fmt.Sprintf("%s: %s", resp.Status, string(body))
		}

		return &errorResp
	}

	// Parse successful response
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// GetOASAPI retrieves an OAS API by ID
func (c *Client) GetOASAPI(ctx context.Context, apiID string, versionName string) (*types.OASAPI, error) {
	apiPath := fmt.Sprintf(OASAPIPath, url.PathEscape(apiID))

	// Add version parameter if specified
	if versionName != "" {
		values := url.Values{}
		values.Set("version_name", versionName)
		apiPath += "?" + values.Encode()
	}

	resp, err := c.doRequest(ctx, http.MethodGet, apiPath, nil)
	if err != nil {
		return nil, err
	}

	// Read the response body directly since it's a raw OAS document
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error status codes
	if resp.StatusCode >= 400 {
		var errorResp types.ErrorResponse
		errorResp.Status = resp.StatusCode
		errorResp.Message = string(body)

		// Try to parse as JSON error response
		if err := json.Unmarshal(body, &errorResp); err != nil {
			errorResp.Message = fmt.Sprintf("%s: %s", resp.Status, string(body))
		}

		return nil, &errorResp
	}

	// Parse the OAS document
	var oasDoc map[string]interface{}
	if err := json.Unmarshal(body, &oasDoc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OAS document: %w", err)
	}

	// Extract API metadata from x-tyk-api-gateway extension
	api, err := c.parseOASDocumentToAPI(oasDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API metadata: %w", err)
	}

	return api, nil
}

// CreateOASAPI creates a new OAS API
func (c *Client) CreateOASAPI(ctx context.Context, oasDocument map[string]interface{}) (*types.OASAPI, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, OASAPIsPath, oasDocument)
	if err != nil {
		return nil, err
	}

	var result types.APIResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	// Create response only returns basic info, need to get full API details
	if result.ID == "" {
		return nil, fmt.Errorf("create response missing API ID")
	}

	// Retrieve the full API details
	return c.GetOASAPI(ctx, result.ID, "")
}

// UpdateOASAPI updates an existing OAS API
func (c *Client) UpdateOASAPI(ctx context.Context, apiID string, oasDocument map[string]interface{}) (*types.OASAPI, error) {
	apiPath := fmt.Sprintf(OASAPIPath, url.PathEscape(apiID))

	resp, err := c.doRequest(ctx, http.MethodPut, apiPath, oasDocument)
	if err != nil {
		return nil, err
	}

	var result types.APIResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	// Update response only returns basic info, need to get full API details
	// Retrieve the full API details using the provided API ID
	return c.GetOASAPI(ctx, apiID, "")
}

// DeleteOASAPI deletes an OAS API by ID
func (c *Client) DeleteOASAPI(ctx context.Context, apiID string) error {
	apiPath := fmt.Sprintf(OASAPIPath, url.PathEscape(apiID))

	resp, err := c.doRequest(ctx, http.MethodDelete, apiPath, nil)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, nil)
}

// ListOASAPIs retrieves a paginated list of OAS APIs from the OAS endpoint. Page numbers are 1-based.
func (c *Client) ListOASAPIs(ctx context.Context, page int) ([]*types.OASAPI, error) {
    listPath := OASAPIsPath
    if page > 0 {
        values := url.Values{}
        values.Set("p", fmt.Sprintf("%d", page))
        listPath += "?" + values.Encode()
    }

    resp, err := c.doRequest(ctx, http.MethodGet, listPath, nil)
    if err != nil {
        return nil, err
    }

    var result types.OASAPIListResponse
    if err := c.handleResponse(resp, &result); err != nil {
        return nil, err
    }
    return result.APIs, nil
}

// ListAPIsDashboard retrieves a paginated list of APIs from the Dashboard aggregate endpoint and maps them.
func (c *Client) ListAPIsDashboard(ctx context.Context, page int) ([]*types.OASAPI, error) {
    listPath := "/api/apis"
    if page > 0 {
        values := url.Values{}
        values.Set("p", fmt.Sprintf("%d", page))
        listPath += "?" + values.Encode()
    }

    resp, err := c.doRequest(ctx, http.MethodGet, listPath, nil)
    if err != nil {
        return nil, err
    }

    // Read the response body directly
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    if resp.StatusCode >= 400 {
        var errorResp types.ErrorResponse
        errorResp.Status = resp.StatusCode
        errorResp.Message = string(body)
        if err := json.Unmarshal(body, &errorResp); err != nil {
            errorResp.Message = fmt.Sprintf("%s: %s", resp.Status, string(body))
        }
        return nil, &errorResp
    }

    var dashboardResponse map[string]interface{}
    if err := json.Unmarshal(body, &dashboardResponse); err != nil {
        return nil, fmt.Errorf("failed to unmarshal dashboard API response: %w", err)
    }

    apisArray, ok := dashboardResponse["apis"].([]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid response format: 'apis' field not found or not an array")
    }

    var apis []*types.OASAPI
    for _, apiItemInterface := range apisArray {
        apiItem, ok := apiItemInterface.(map[string]interface{})
        if !ok {
            continue
        }
        apiDefInterface, ok := apiItem["api_definition"]
        if !ok {
            continue
        }
        apiDef, ok := apiDefInterface.(map[string]interface{})
        if !ok {
            continue
        }

        apiID, _ := apiDef["api_id"].(string)
        name, _ := apiDef["name"].(string)
        var listenPath string
        if proxyInterface, ok := apiDef["proxy"]; ok {
            if proxy, ok := proxyInterface.(map[string]interface{}); ok {
                if path, ok := proxy["listen_path"].(string); ok {
                    listenPath = path
                }
            }
        }

        if apiID != "" {
            apis = append(apis, &types.OASAPI{
                ID:             apiID,
                Name:           name,
                ListenPath:     listenPath,
                DefaultVersion: "v1",
            })
        }
    }
    return apis, nil
}

// ListOASAPIVersions lists all versions for an OAS API
func (c *Client) ListOASAPIVersions(ctx context.Context, apiID string) ([]string, string, error) {
	versionsPath := fmt.Sprintf(OASAPIVersionsPath, url.PathEscape(apiID))

	resp, err := c.doRequest(ctx, http.MethodGet, versionsPath, nil)
	if err != nil {
		return nil, "", err
	}

	var result types.VersionListResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, "", err
	}

	return result.Versions, result.Default, nil
}

// SwitchDefaultVersion switches the default version of an API
func (c *Client) SwitchDefaultVersion(ctx context.Context, apiID string, versionName string) error {
	apiPath := fmt.Sprintf(OASAPIPath, url.PathEscape(apiID))

	req := map[string]interface{}{
		"set_default_version": versionName,
	}

	resp, err := c.doRequest(ctx, http.MethodPatch, apiPath, req)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, nil)
}

// Health checks the health of the Tyk Dashboard
func (c *Client) Health(ctx context.Context) error {
	resp, err := c.doRequest(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dashboard health check failed: %s", resp.Status)
	}

	return nil
}

// parseOASDocumentToAPI extracts API metadata from an OAS document with Tyk extensions
func (c *Client) parseOASDocumentToAPI(oasDoc map[string]interface{}) (*types.OASAPI, error) {
	// Extract basic OAS info
	info, ok := oasDoc["info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid OAS document: missing info section")
	}

	// Extract x-tyk-api-gateway extension
	tykExt, ok := oasDoc["x-tyk-api-gateway"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid Tyk OAS document: missing x-tyk-api-gateway extension")
	}

	// Extract API info from Tyk extension
	apiInfo, ok := tykExt["info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid Tyk OAS document: missing info in x-tyk-api-gateway")
	}

	// Extract server info to get listen path
	var listenPath string
	if server, ok := tykExt["server"].(map[string]interface{}); ok {
		if listenPathInfo, ok := server["listenPath"].(map[string]interface{}); ok {
			if path, ok := listenPathInfo["value"].(string); ok {
				listenPath = path
			}
		}
	}

	// Extract upstream URL
	var upstreamURL string
	if upstream, ok := tykExt["upstream"].(map[string]interface{}); ok {
		if url, ok := upstream["url"].(string); ok {
			upstreamURL = url
		}
	}

	// Build the API object
	api := &types.OASAPI{
		ID:          getString(apiInfo, "id"),
		Name:        getString(apiInfo, "name"),
		ListenPath:  listenPath,
		UpstreamURL: upstreamURL,
		OAS:         oasDoc,
		// For now, we'll set these to empty since they might not be in this format
		DefaultVersion: "v1",
		VersionData:    make(map[string]*types.APIVersion),
		CreatedAt:      "",
		UpdatedAt:      "",
	}

	// Extract title from OAS info if name is empty
	if api.Name == "" {
		api.Name = getString(info, "title")
	}

	return api, nil
}

// getString safely extracts a string value from a map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
