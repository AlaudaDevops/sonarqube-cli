package config

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

func ReplaceVariables(s, taskRunID, pluginName string) (string, error) {
	if taskRunID == "" && (strings.Contains(s, "{{TASK_RUN_ID}}") || strings.Contains(s, "${TASK_RUN_ID}")) {
		return "", fmt.Errorf("taskRunID is required for template: %s", s)
	}
	if pluginName == "" && (strings.Contains(s, "{{PLUGIN_NAME}}") || strings.Contains(s, "${PLUGIN_NAME}")) {
		return "", fmt.Errorf("pluginName is required for template: %s", s)
	}

	s = strings.ReplaceAll(s, "{{TASK_RUN_ID}}", taskRunID)
	s = strings.ReplaceAll(s, "${TASK_RUN_ID}", taskRunID)
	s = strings.ReplaceAll(s, "{{PLUGIN_NAME}}", pluginName)
	s = strings.ReplaceAll(s, "${PLUGIN_NAME}", pluginName)
	return os.ExpandEnv(s), nil
}

func ApplyTemplate(res *TempResource, taskRunID, pluginName string) error {
	var err error
	res.Group.Name, err = ReplaceVariables(res.Group.Name, taskRunID, pluginName)
	if err != nil {
		return err
	}
	res.Group.Description, err = ReplaceVariables(res.Group.Description, taskRunID, pluginName)
	if err != nil {
		return err
	}

	res.User.Login, err = ReplaceVariables(res.User.Login, taskRunID, pluginName)
	if err != nil {
		return err
	}
	res.User.Name, err = ReplaceVariables(res.User.Name, taskRunID, pluginName)
	if err != nil {
		return err
	}
	res.User.Email, err = ReplaceVariables(res.User.Email, taskRunID, pluginName)
	if err != nil {
		return err
	}
	res.User.Password, err = ReplaceVariables(res.User.Password, taskRunID, pluginName)
	if err != nil {
		return err
	}

	for i := range res.User.Groups {
		res.User.Groups[i], err = ReplaceVariables(res.User.Groups[i], taskRunID, pluginName)
		if err != nil {
			return err
		}
	}

	for i := range res.Projects {
		res.Projects[i].Key, err = ReplaceVariables(res.Projects[i].Key, taskRunID, pluginName)
		if err != nil {
			return err
		}
		res.Projects[i].Name, err = ReplaceVariables(res.Projects[i].Name, taskRunID, pluginName)
		if err != nil {
			return err
		}
	}

	res.PermissionTemplate.Name, err = ReplaceVariables(res.PermissionTemplate.Name, taskRunID, pluginName)
	if err != nil {
		return err
	}
	res.PermissionTemplate.Description, err = ReplaceVariables(res.PermissionTemplate.Description, taskRunID, pluginName)
	if err != nil {
		return err
	}
	res.PermissionTemplate.ProjectKeyPattern, err = ReplaceVariables(res.PermissionTemplate.ProjectKeyPattern, taskRunID, pluginName)
	if err != nil {
		return err
	}
	return nil
}

func RenderTemplate(templatePath, outputPath string, data interface{}) error {
	tmplData, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	tmpl, err := template.New("output").Parse(string(tmplData))
	if err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}
