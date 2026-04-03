package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GenerateUserToken generates a new user token in SonarQube.
// SECURITY: The returned token value is extremely sensitive and MUST NOT be logged or displayed.
// It is intended to be passed directly to a secure storage (like a file with restricted permissions).
func (c *Client) GenerateUserToken(login, name string) (string, error) {
	params := url.Values{}
	params.Set("login", login)
	params.Set("name", name)

	bodyData, err := c.doRequest("POST", "/api/user_tokens/generate", params, nil)
	if err != nil {
		if bodyContains(err, "already exists") {
			return "", alreadyExistsError("token", name, err)
		}
		return "", err
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(bodyData, &result); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return result.Token, nil
}

// RevokeUserToken revokes a user token in SonarQube.
func (c *Client) RevokeUserToken(login, name string) error {
	params := url.Values{}
	params.Set("login", login)
	params.Set("name", name)
	if _, err := c.doRequest("POST", "/api/user_tokens/revoke", params, nil); err != nil {
		if hasStatus(err, http.StatusNotFound) {
			return nil
		}
		return err
	}
	return nil
}
