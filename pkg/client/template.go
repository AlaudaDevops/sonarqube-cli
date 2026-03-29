package client

import (
	"net/url"
	"strings"

	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/config"
)

// CreatePermissionTemplate creates a new permission template in SonarQube and assigns it to a group.
func (c *Client) CreatePermissionTemplate(tmpl config.PermissionTemplate, groupName string) (bool, error) {
	created := false
	params := url.Values{}
	params.Set("name", tmpl.Name)
	params.Set("description", tmpl.Description)
	params.Set("projectKeyPattern", tmpl.ProjectKeyPattern)

	if err := c.doRequest("POST", "/api/permissions/create_template", params, nil); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return false, alreadyExistsError("permission template", tmpl.Name, err)
		}
		return false, err
	}
	created = true

	for _, perm := range tmpl.Permissions {
		p := url.Values{}
		p.Set("templateName", tmpl.Name)
		p.Set("groupName", groupName)
		p.Set("permission", perm)
		if err := c.doRequest("POST", "/api/permissions/add_group_to_template", p, nil); err != nil {
			return created, err
		}
	}

	return created, nil
}

// DeletePermissionTemplate deletes a permission template from SonarQube.
func (c *Client) DeletePermissionTemplate(name string) error {
	params := url.Values{}
	params.Set("templateName", name)
	if err := c.doRequest("POST", "/api/permissions/delete_template", params, nil); err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return nil
}
