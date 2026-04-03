package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ReplaceVariables replaces template variables (TASK_RUN_ID, PLUGIN_NAME) in a string.
// It returns an error if a required variable is missing in the template string.
func ReplaceVariables(s, taskRunID, pluginName string) (string, error) {
	// Fail only when the template actually references a missing variable, so plain strings still work without both inputs.
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
	return s, nil
}

// ApplyTemplate resolves template variables in a TempResource.
func ApplyTemplate(res *TempResource, taskRunID, pluginName string) error {
	var err error
	res.PluginName, err = ReplaceVariables(res.PluginName, taskRunID, pluginName)
	if err != nil {
		return err
	}

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

// RenderTemplate renders a template file with the given data and writes it to outputPath.
func RenderTemplate(templatePath, outputPath string, data interface{}) error {
	tmplData, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	tmpl, err := template.New("output").Parse(string(tmplData))
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.CreateTemp(filepath.Dir(outputPath), ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := f.Name()
	removeTmp := true
	defer func() {
		_ = f.Close()
		if removeTmp {
			_ = os.Remove(tmpPath)
		}
	}()

	if err := f.Chmod(0600); err != nil {
		return err
	}
	if err := tmpl.Execute(f, data); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, outputPath); err != nil {
		return err
	}
	removeTmp = false
	return nil
}
