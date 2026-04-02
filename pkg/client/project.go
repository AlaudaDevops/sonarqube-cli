package client

import (
	"net/http"
	"net/url"

	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/config"
)

// AddGlobalPermission adds a global permission to a user.
func (c *Client) AddGlobalPermission(login, permission string) error {
	params := url.Values{}
	params.Set("login", login)
	params.Set("permission", permission)
	if _, err := c.doRequest("POST", "/api/permissions/add_user", params, nil); err != nil {
		if bodyContains(err, "already exists", "already been granted") {
			return alreadyExistsError("global permission", permission, err)
		}
		return err
	}
	return nil
}

// CreateProject creates a new project in SonarQube.
func (c *Client) CreateProject(proj config.Project) (bool, error) {
	params := url.Values{}
	params.Set("project", proj.Key)
	params.Set("name", proj.Name)
	params.Set("visibility", proj.Visibility)
	if _, err := c.doRequest("POST", "/api/projects/create", params, nil); err != nil {
		if bodyContains(err, "already exists") {
			return false, alreadyExistsError("project", proj.Key, err)
		}
		return false, err
	}
	return true, nil
}

// DeleteProject deletes a project from SonarQube.
func (c *Client) DeleteProject(key string) error {
	params := url.Values{}
	params.Set("project", key)
	if _, err := c.doRequest("POST", "/api/projects/delete", params, nil); err != nil {
		if hasStatus(err, http.StatusNotFound) {
			return nil
		}
		return err
	}
	return nil
}
