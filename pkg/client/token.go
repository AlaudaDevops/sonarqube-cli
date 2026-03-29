package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// GenerateUserToken generates a new user token in SonarQube.
// SECURITY: The returned token value is extremely sensitive and MUST NOT be logged or displayed.
// It is intended to be passed directly to a secure storage (like a file with restricted permissions).
func (c *Client) GenerateUserToken(login, name string) (string, error) {
	params := url.Values{}
	params.Set("login", login)
	params.Set("name", name)

	u := c.endpoint + "/api/user_tokens/generate"
	req, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return "", err
	}

	req.URL.RawQuery = params.Encode()
	req.SetBasicAuth(c.token, "")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyData, _ := io.ReadAll(resp.Body)
		// SECURITY: Ensure bodyData doesn't contain the token before logging
		err := fmt.Errorf("api error %d: %s", resp.StatusCode, string(bodyData))
		if strings.Contains(err.Error(), "already exists") {
			return "", alreadyExistsError("token", name, err)
		}
		return "", err
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return result.Token, nil
}

// RevokeUserToken revokes a user token in SonarQube.
func (c *Client) RevokeUserToken(login, name string) error {
	params := url.Values{}
	params.Set("login", login)
	params.Set("name", name)
	if err := c.doRequest("POST", "/api/user_tokens/revoke", params, nil); err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return nil
}
