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
	OASAPIPath         = "/api/apis/oas/%s"           // {apiId}
	OASAPIVersionsPath = "/api/apis/oas/%s/versions"  // {apiId}
	
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

	baseURL, err := url.Parse(config.DashURL)
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
	fullURL.Path = path

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set(HeaderAuthorization, c.config.AuthToken)
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

	var result types.OASAPIResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.API, nil
}

// CreateOASAPI creates a new OAS API
func (c *Client) CreateOASAPI(ctx context.Context, req *types.CreateOASAPIRequest) (*types.OASAPI, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, OASAPIsPath, req)
	if err != nil {
		return nil, err
	}

	var result types.OASAPIResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.API, nil
}

// UpdateOASAPI updates an existing OAS API
func (c *Client) UpdateOASAPI(ctx context.Context, apiID string, req *types.UpdateOASAPIRequest) (*types.OASAPI, error) {
	apiPath := fmt.Sprintf(OASAPIPath, url.PathEscape(apiID))

	resp, err := c.doRequest(ctx, http.MethodPut, apiPath, req)
	if err != nil {
		return nil, err
	}

	var result types.OASAPIResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.API, nil
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