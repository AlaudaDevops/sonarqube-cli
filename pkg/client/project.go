package client

import (
	"net/url"
	"strings"

	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/config"
)

func (c *Client) AddGlobalPermission(login, permission string) error {
	params := url.Values{}
	params.Set("login", login)
	params.Set("permission", permission)
	if err := c.doRequest("POST", "/api/permissions/add_user", params); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}

func (c *Client) CreateProject(proj config.Project) error {
	params := url.Values{}
	params.Set("project", proj.Key)
	params.Set("name", proj.Name)
	params.Set("visibility", proj.Visibility)
	if err := c.doRequest("POST", "/api/projects/create", params); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}

func (c *Client) DeleteProject(key string) error {
	params := url.Values{}
	params.Set("project", key)
	if err := c.doRequest("POST", "/api/projects/delete", params); err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return nil
}
