package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// buildAllowedEnvs returns the supported placeholders and applies optional file-based overrides.
func buildAllowedEnvs(overrides map[string]string) map[string]string {
	allowedEnvs := map[string]string{
		"SONARQUBE_URL":           os.Getenv("SONARQUBE_URL"),
		"SONARQUBE_MANAGER_TOKEN": os.Getenv("SONARQUBE_MANAGER_TOKEN"),
		"TEMP_USER_PASSWORD":      os.Getenv("TEMP_USER_PASSWORD"),
	}
	for key, value := range overrides {
		if _, ok := allowedEnvs[key]; ok {
			// Ignore unknown override keys so callers cannot expand arbitrary host environment variables.
			allowedEnvs[key] = value
		}
	}
	return allowedEnvs
}

// readExpandedConfig reads YAML and expands only whitelisted placeholders.
func readExpandedConfig(path string, overrides map[string]string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// SECURITY: Use a whitelist for environment variable substitution
	// This prevents leaking sensitive system-wide environment variables (e.g., AWS keys)
	allowedEnvs := buildAllowedEnvs(overrides)

	expanded := os.Expand(string(data), func(s string) string {
		if val, ok := allowedEnvs[s]; ok && val != "" {
			return val
		}
		// Preserve template variables for later processing during resource selection.
		if s == "TASK_RUN_ID" || s == "PLUGIN_NAME" {
			return fmt.Sprintf("${%s}", s)
		}
		// Unknown placeholders collapse to empty so YAML parsing sees the same behavior as missing env vars.
		return allowedEnvs[s]
	})

	return []byte(expanded), nil
}

// Load reads and parses the configuration file from the given path.
// It performs environment variable substitution based on a whitelist.
func Load(path string) (*Config, error) {
	return LoadWithVariables(path, nil)
}

// LoadWithVariables reads and parses the configuration file from the given path.
// It performs environment variable substitution based on a whitelist and applies
// the provided variables as overrides for the supported placeholders.
func LoadWithVariables(path string, overrides map[string]string) (*Config, error) {
	data, err := readExpandedConfig(path, overrides)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadConnection reads only the SonarQube endpoint and manager token.
// It skips temp_resources validation so cleanup can proceed with a saved state file.
func LoadConnection(path string) (string, string, error) {
	return LoadConnectionWithVariables(path, nil)
}

// LoadConnectionWithVariables reads only the SonarQube endpoint and manager token.
// It skips temp_resources validation so cleanup can proceed with a saved state file.
func LoadConnectionWithVariables(path string, overrides map[string]string) (string, string, error) {
	data, err := readExpandedConfig(path, overrides)
	if err != nil {
		return "", "", err
	}

	var cfg struct {
		SonarQube struct {
			Endpoint string  `yaml:"endpoint"`
			Manager  Manager `yaml:"manager"`
		} `yaml:"sonarqube"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", "", err
	}

	if err := validateConnection(cfg.SonarQube.Endpoint, cfg.SonarQube.Manager.Token); err != nil {
		return "", "", err
	}
	return cfg.SonarQube.Endpoint, cfg.SonarQube.Manager.Token, nil
}

// validateConnection validates endpoint format and manager token presence.
func validateConnection(endpoint, token string) error {
	if endpoint == "" {
		return fmt.Errorf("sonarqube.endpoint is required")
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid sonarqube.endpoint: %w", err)
	}
	if u.Scheme == "" {
		return fmt.Errorf("sonarqube.endpoint must include scheme (http:// or https://)")
	}
	if strings.ToLower(u.Scheme) != "https" {
		// HTTP is allowed only as an explicit opt-in for local development or isolated test environments.
		if os.Getenv("SONARQUBE_ALLOW_HTTP") != "true" {
			return fmt.Errorf("sonarqube.endpoint must use https for security (set SONARQUBE_ALLOW_HTTP=true to override for development)")
		}
		fmt.Fprintf(os.Stderr, "Warning: sonarqube.endpoint is using http (allowed by SONARQUBE_ALLOW_HTTP=true)\n")
	}

	if token == "" {
		return fmt.Errorf("sonarqube.manager.token is required")
	}
	return nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if err := validateConnection(c.SonarQube.Endpoint, c.SonarQube.Manager.Token); err != nil {
		return err
	}

	for i, res := range c.SonarQube.TempResources {
		if res.PluginName == "" {
			return fmt.Errorf("temp_resources[%d].plugin_name is required", i)
		}
		if res.Group.Name == "" {
			return fmt.Errorf("group name for plugin %s is required", res.PluginName)
		}
		if res.User.Login == "" {
			return fmt.Errorf("user login for plugin %s is required", res.PluginName)
		}
		for j, proj := range res.Projects {
			if proj.Key == "" {
				return fmt.Errorf("project[%d].key for plugin %s is required", j, res.PluginName)
			}
		}
	}
	return nil
}
