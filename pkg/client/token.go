package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) GenerateUserToken(login, name string) (string, error) {
	// Revoke existing token if it exists (idempotency)
	_ = c.RevokeUserToken(login, name)

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
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Token, nil
}

func (c *Client) RevokeUserToken(login, name string) error {
	params := url.Values{}
	params.Set("login", login)
	params.Set("name", name)
	if err := c.doRequest("POST", "/api/user_tokens/revoke", params); err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return nil
}
