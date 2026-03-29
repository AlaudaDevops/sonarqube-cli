package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
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

func (c *Config) Validate() error {
	if c.SonarQube.Endpoint == "" {
		return fmt.Errorf("sonarqube.endpoint is required")
	}

	u, err := url.Parse(c.SonarQube.Endpoint)
	if err != nil {
		return fmt.Errorf("invalid sonarqube.endpoint: %w", err)
	}
	if u.Scheme == "" {
		return fmt.Errorf("sonarqube.endpoint must include scheme (http:// or https://)")
	}
	if strings.ToLower(u.Scheme) != "https" {
		fmt.Fprintf(os.Stderr, "Warning: sonarqube.endpoint should use https for security\n")
	}

	if c.SonarQube.Manager.Token == "" {
		return fmt.Errorf("sonarqube.manager.token is required")
	}
	return nil
}
