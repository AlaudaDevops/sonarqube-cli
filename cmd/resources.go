package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/client"
	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/config"
	"gopkg.in/yaml.v3"
)

var resourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Manage SonarQube temporary resources",
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create temporary resources",
	RunE:  runCreate,
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup temporary resources",
	RunE:  runCleanup,
}

var (
	configFile     string
	taskRunID      string
	plugin         string
	resolvedConfig string
	outputTemplate string
	outputFile     string
	tokenFile      string
)

func init() {
	resourcesCmd.AddCommand(createCmd, cleanupCmd)

	for _, cmd := range []*cobra.Command{createCmd, cleanupCmd} {
		cmd.Flags().StringVar(&configFile, "config", "", "Config file path")
		cmd.Flags().StringVar(&taskRunID, "task-run-id", "", "Tekton TaskRun ID (only needed if config is template)")
		cmd.Flags().StringVar(&plugin, "plugin", "", "Plugin name")
		cmd.Flags().StringVar(&resolvedConfig, "resolved-config", "", "Path to save resolved config")
		cmd.Flags().StringVar(&outputTemplate, "output-template", "", "Path to output template")
		cmd.Flags().StringVar(&outputFile, "output-file", "", "Path to output file")
		cmd.Flags().StringVar(&tokenFile, "token-file", "", "Write token to file (recommended for security)")
		cmd.MarkFlagRequired("config")
		cmd.MarkFlagRequired("plugin")
	}
}

func runCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configFile)
	if err != nil {
		return err
	}

	c := client.New(cfg.SonarQube.Endpoint, cfg.SonarQube.Manager.Token)

	for i, res := range cfg.SonarQube.TempResources {
		if res.PluginName != plugin {
			continue
		}

		// Apply variables to config
		if err := config.ApplyTemplate(&res, taskRunID, plugin); err != nil {
			return err
		}
		cfg.SonarQube.TempResources[i] = res // Update back to config for saving later

		if err := c.CreateGroup(res.Group.Name, res.Group.Description); err != nil {
			return err
		}

		if err := c.CreateUser(res.User); err != nil {
			cleanupResources(c, res, taskRunID)
			return err
		}

		for _, perm := range res.GlobalPermissions {
			if err := c.AddGlobalPermission(res.User.Login, perm); err != nil {
				cleanupResources(c, res, taskRunID)
				return err
			}
		}

		if err := c.CreatePermissionTemplate(res.PermissionTemplate, res.Group.Name); err != nil {
			cleanupResources(c, res, taskRunID)
			return err
		}

		for _, proj := range res.Projects {
			if err := c.CreateProject(proj); err != nil {
				cleanupResources(c, res, taskRunID)
				return err
			}
		}

		tokenName := fmt.Sprintf("test-token-%s", taskRunID)
		if taskRunID == "" {
			tokenName = "test-token"
		}
		token, err := c.GenerateUserToken(res.User.Login, tokenName)
		if err != nil {
			cleanupResources(c, res, taskRunID)
			return err
		}

		// SECURITY: Token must be written to file with restricted permissions
		if tokenFile == "" {
			cleanupResources(c, res, taskRunID)
			return fmt.Errorf("--token-file is required for security (prevents token leakage to logs)")
		}

		tokenData := fmt.Sprintf("SONARQUBE_TOKEN=%s\nSONARQUBE_USER=%s\n", token, res.User.Login)
		if err := os.WriteFile(tokenFile, []byte(tokenData), 0600); err != nil {
			cleanupResources(c, res, taskRunID)
			return fmt.Errorf("failed to write token file: %w", err)
		}
		fmt.Printf("Token written to %s\n", tokenFile)

		// Save resolved config
		if resolvedConfig != "" {
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			if err := os.WriteFile(resolvedConfig, data, 0600); err != nil {
				return fmt.Errorf("failed to write resolved config: %w", err)
			}
		}

		// Render output template
		if outputTemplate != "" && outputFile != "" {
			data := map[string]interface{}{
				"Endpoint": cfg.SonarQube.Endpoint,
				"Users": []map[string]string{
					{
						"Login":    res.User.Login,
						"Password": res.User.Password,
						"Token":    token,
					},
				},
				"Projects": res.Projects,
			}
			if err := config.RenderTemplate(outputTemplate, outputFile, data); err != nil {
				cleanupResources(c, res, taskRunID)
				return err
			}
		}

		fmt.Printf("Created resources for plugin: %s\n", plugin)
		return nil
	}

	return fmt.Errorf("plugin not found: %s", plugin)
}

func cleanupResources(c *client.Client, res config.TempResource, taskRunID string) error {
	var errs []error
	tokenName := fmt.Sprintf("test-token-%s", taskRunID)
	if taskRunID == "" {
		tokenName = "test-token"
	}
	// Revoke token if it exists
	if err := c.RevokeUserToken(res.User.Login, tokenName); err != nil {
		// Ignore error if token doesn't exist (which is likely if creation failed before token)
	}

	for _, proj := range res.Projects {
		if err := c.DeleteProject(proj.Key); err != nil {
			errs = append(errs, fmt.Errorf("delete project %s: %w", proj.Key, err))
		}
	}

	if err := c.DeletePermissionTemplate(res.PermissionTemplate.Name); err != nil {
		errs = append(errs, fmt.Errorf("delete template: %w", err))
	}
	if err := c.DeleteUser(res.User.Login); err != nil {
		errs = append(errs, fmt.Errorf("delete user: %w", err))
	}
	if err := c.DeleteGroup(res.Group.Name); err != nil {
		errs = append(errs, fmt.Errorf("delete group: %w", err))
	}

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", e)
		}
		return fmt.Errorf("cleanup failed")
	}
	return nil
}

func runCleanup(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configFile)
	if err != nil {
		return err
	}

	c := client.New(cfg.SonarQube.Endpoint, cfg.SonarQube.Manager.Token)

	for _, res := range cfg.SonarQube.TempResources {
		if res.PluginName != plugin {
			continue
		}

		// Only apply template if taskRunID is provided (backward compatibility)
		if taskRunID != "" {
			if err := config.ApplyTemplate(&res, taskRunID, plugin); err != nil {
				return err
			}
		}

		if err := cleanupResources(c, res, taskRunID); err != nil {
			return fmt.Errorf("cleanup failed for plugin %s: %w", plugin, err)
		}

		fmt.Printf("Cleaned up resources for plugin: %s\n", plugin)
		return nil
	}

	return fmt.Errorf("plugin not found: %s", plugin)
}
