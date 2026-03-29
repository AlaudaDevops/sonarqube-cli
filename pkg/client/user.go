// Package client provides a SonarQube API client for resource management.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/config"
)

// Client is a wrapper for SonarQube API.
type Client struct {
	endpoint string
	token    string
	client   *http.Client
}

// New creates a new SonarQube Client.
func New(endpoint, token string) *Client {
	// Ensure endpoint doesn't end with a slash to avoid double slashes in paths
	endpoint = strings.TrimSuffix(endpoint, "/")

	timeout := 60 * time.Second
	if t := os.Getenv("SONARQUBE_TIMEOUT"); t != "" {
		if d, err := time.ParseDuration(t); err == nil {
			timeout = d
		}
	}

	return &Client{
		endpoint: endpoint,
		token:    token,
		client:   &http.Client{Timeout: timeout},
	}
}

func (c *Client) doRequest(method, path string, params url.Values, body interface{}) error {
	u := c.endpoint + path
	var bodyReader io.Reader
	contentType := ""

	if body != nil {
		// JSON body (typically for v2 APIs)
		jsonData, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(jsonData)
		contentType = "application/json"
		// Query params can still be added to the URL if needed
		if params != nil {
			u += "?" + params.Encode()
		}
	} else if params != nil && (method == "POST" || method == "PUT") {
		// Form data body (typically for classic v1 APIs)
		bodyReader = strings.NewReader(params.Encode())
		contentType = "application/x-www-form-urlencoded"
	} else if params != nil {
		// Query params only (for GET/DELETE)
		u += "?" + params.Encode()
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	req.SetBasicAuth(c.token, "") // SonarQube API uses token as username with empty password
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyData, _ := io.ReadAll(resp.Body)
		// Mask query params in the URL for security in error logs
		safeURL := c.endpoint + path
		return fmt.Errorf("api error %d on %s %s: %s", resp.StatusCode, method, safeURL, string(bodyData))
	}

	// Read and discard body to enable connection reuse
	io.Copy(io.Discard, resp.Body)
	return nil
}

// CreateGroup creates a new user group in SonarQube.
func (c *Client) CreateGroup(name, description string) (bool, error) {
	path := "/api/v2/authorizations/groups"
	body := map[string]string{
		"name":        name,
		"description": description,
	}
	if err := c.doRequest("POST", path, nil, body); err != nil {
		// 409 Conflict means group already exists in v2 API
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "409") {
			return false, alreadyExistsError("group", name, err)
		}
		return false, err
	}
	return true, nil
}

// DeleteGroup deletes a user group from SonarQube.
func (c *Client) DeleteGroup(name string) error {
	// v2 API uses DELETE method and path parameter
	path := fmt.Sprintf("/api/v2/authorizations/groups/%s", url.PathEscape(name))
	if err := c.doRequest("DELETE", path, nil, nil); err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return nil
}

// CreateUser creates a new user account in SonarQube.
func (c *Client) CreateUser(user config.User) (bool, error) {
	created := false
	params := url.Values{}
	params.Set("login", user.Login)
	params.Set("name", user.Name)
	params.Set("email", user.Email)
	params.Set("password", user.Password)
	// Use v1 API for users as v2 might not be available
	if err := c.doRequest("POST", "/api/users/create", params, nil); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return false, alreadyExistsError("user", user.Login, err)
		}
		return false, err
	}
	created = true

	for _, group := range user.Groups {
		params := url.Values{}
		params.Set("login", user.Login)
		params.Set("name", group)
		if err := c.doRequest("POST", "/api/user_groups/add_user", params, nil); err != nil {
			return created, err
		}
	}
	return created, nil
}

// DeleteUser deactivates a user account in SonarQube.
func (c *Client) DeleteUser(login string) error {
	params := url.Values{}
	params.Set("login", login)
	// Use v1 API (deactivate) for users
	if err := c.doRequest("POST", "/api/users/deactivate", params, nil); err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return nil
}
