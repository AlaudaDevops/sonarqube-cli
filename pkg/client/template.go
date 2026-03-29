package client

import (
	"net/url"
	"strings"

	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/config"
)

func (c *Client) CreatePermissionTemplate(tpl config.PermissionTemplate, groupName string) error {
	params := url.Values{}
	params.Set("name", tpl.Name)
	params.Set("description", tpl.Description)
	params.Set("projectKeyPattern", tpl.ProjectKeyPattern)
	if err := c.doRequest("POST", "/api/permissions/create_template", params); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	for _, perm := range tpl.Permissions {
		params := url.Values{}
		params.Set("templateName", tpl.Name)
		params.Set("groupName", groupName)
		params.Set("permission", perm)
		if err := c.doRequest("POST", "/api/permissions/add_group_to_template", params); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}
	return nil
}

func (c *Client) DeletePermissionTemplate(name string) error {
	params := url.Values{}
	params.Set("templateName", name)
	if err := c.doRequest("POST", "/api/permissions/delete_template", params); err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return err
	}
	return nil
}
