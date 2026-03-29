package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/config"
)

type Client struct {
	endpoint string
	token    string
	client   *http.Client
}

func New(endpoint, token string) *Client {
	return &Client{
		endpoint: endpoint,
		token:    token,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) doRequest(method, path string, params url.Values) error {
	u := c.endpoint + path
	if params != nil {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.token, "") // SonarQube API uses token as username with empty password
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Read and discard body to enable connection reuse
	io.Copy(io.Discard, resp.Body)
	return nil
}

func (c *Client) CreateGroup(name, description string) error {
	params := url.Values{}
	params.Set("name", name)
	params.Set("description", description)
	if err := c.doRequest("POST", "/api/user_groups/create", params); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return err
	}
	return nil
}

func (c *Client) DeleteGroup(name string) error {
	params := url.Values{}
	params.Set("name", name)
	if err := c.doRequest("POST", "/api/user_groups/delete", params); err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return nil
}

func (c *Client) CreateUser(user config.User) error {
	params := url.Values{}
	params.Set("login", user.Login)
	params.Set("name", user.Name)
	params.Set("email", user.Email)
	params.Set("password", user.Password)
	if err := c.doRequest("POST", "/api/users/create", params); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	for _, group := range user.Groups {
		params := url.Values{}
		params.Set("login", user.Login)
		params.Set("name", group)
		if err := c.doRequest("POST", "/api/user_groups/add_user", params); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}
	return nil
}

func (c *Client) DeleteUser(login string) error {
	params := url.Values{}
	params.Set("login", login)
	if err := c.doRequest("POST", "/api/users/deactivate", params); err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return nil
}
