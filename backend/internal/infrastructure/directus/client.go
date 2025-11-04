package directus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Directus API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

// NewClient creates a new Directus client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the authentication token
func (c *Client) SetToken(token string) {
	c.Token = token
}

// AuthRequest represents authentication request
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Data struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Expires      int64  `json:"expires"`
	} `json:"data"`
}

// Authenticate authenticates with Directus and sets the token
func (c *Client) Authenticate(email, password string) error {
	authReq := AuthRequest{
		Email:    email,
		Password: password,
	}

	body, err := json.Marshal(authReq)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	resp, err := c.HTTPClient.Post(
		fmt.Sprintf("%s/auth/login", c.BaseURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	c.Token = authResp.Data.AccessToken
	return nil
}

// Request makes a generic HTTP request to Directus API
func (c *Client) Request(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.BaseURL, path), reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// GetItems retrieves items from a collection
func (c *Client) GetItems(collection string, params map[string]string) (interface{}, error) {
	path := fmt.Sprintf("/items/%s", collection)
	if len(params) > 0 {
		path += "?"
		first := true
		for k, v := range params {
			if !first {
				path += "&"
			}
			path += fmt.Sprintf("%s=%s", k, v)
			first = false
		}
	}

	var result interface{}
	err := c.Request("GET", path, nil, &result)
	return result, err
}

// GetItem retrieves a single item from a collection
func (c *Client) GetItem(collection, id string) (interface{}, error) {
	path := fmt.Sprintf("/items/%s/%s", collection, id)
	var result interface{}
	err := c.Request("GET", path, nil, &result)
	return result, err
}

// CreateItem creates a new item in a collection
func (c *Client) CreateItem(collection string, data interface{}) (interface{}, error) {
	path := fmt.Sprintf("/items/%s", collection)
	var result interface{}
	err := c.Request("POST", path, data, &result)
	return result, err
}

// UpdateItem updates an existing item in a collection
func (c *Client) UpdateItem(collection, id string, data interface{}) (interface{}, error) {
	path := fmt.Sprintf("/items/%s/%s", collection, id)
	var result interface{}
	err := c.Request("PATCH", path, data, &result)
	return result, err
}

// DeleteItem deletes an item from a collection
func (c *Client) DeleteItem(collection, id string) error {
	path := fmt.Sprintf("/items/%s/%s", collection, id)
	return c.Request("DELETE", path, nil, nil)
}

// GetFiles retrieves files
func (c *Client) GetFiles(params map[string]string) (interface{}, error) {
	path := "/files"
	if len(params) > 0 {
		path += "?"
		first := true
		for k, v := range params {
			if !first {
				path += "&"
			}
			path += fmt.Sprintf("%s=%s", k, v)
			first = false
		}
	}

	var result interface{}
	err := c.Request("GET", path, nil, &result)
	return result, err
}

// UploadFile uploads a file to Directus
func (c *Client) UploadFile(fileData io.Reader, fileName string) (interface{}, error) {
	// Note: This is a simplified version. For actual file upload,
	// you'll need to use multipart/form-data
	path := "/files"
	var result interface{}
	err := c.Request("POST", path, fileData, &result)
	return result, err
}
