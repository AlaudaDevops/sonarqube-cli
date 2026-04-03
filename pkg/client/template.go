package client

import (
	"net/http"
	"net/url"

	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/config"
)

// CreatePermissionTemplate creates a new permission template in SonarQube and assigns it to a group.
func (c *Client) CreatePermissionTemplate(tmpl config.PermissionTemplate, groupName string) (bool, error) {
	created := false
	params := url.Values{}
	params.Set("name", tmpl.Name)
	params.Set("description", tmpl.Description)
	params.Set("projectKeyPattern", tmpl.ProjectKeyPattern)

	if _, err := c.doRequest("POST", "/api/permissions/create_template", params, nil); err != nil {
		if bodyContains(err, "already exists") {
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
		if _, err := c.doRequest("POST", "/api/permissions/add_group_to_template", p, nil); err != nil {
			// Return created=true so rollback can remove the template when only permission binding failed.
			return created, err
		}
	}

	return created, nil
}

// DeletePermissionTemplate deletes a permission template from SonarQube.
func (c *Client) DeletePermissionTemplate(name string) error {
	params := url.Values{}
	params.Set("templateName", name)
	if _, err := c.doRequest("POST", "/api/permissions/delete_template", params, nil); err != nil {
		if hasStatus(err, http.StatusNotFound) {
			return nil
		}
		return err
	}
	return nil
}
