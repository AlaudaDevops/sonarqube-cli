// Package config provides configuration loading and template rendering for SonarQube CLI.
package config

// Config represents the top-level configuration for SonarQube CLI.
type Config struct {
	SonarQube SonarQubeConfig `yaml:"sonarqube"`
}

// SonarQubeConfig contains SonarQube connection and resource management settings.
type SonarQubeConfig struct {
	Endpoint      string         `yaml:"endpoint"`
	Manager       Manager        `yaml:"manager"`
	TempResources []TempResource `yaml:"temp_resources"`
}

// Manager represents the SonarQube administrator credentials.
type Manager struct {
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
}

// TempResource defines a set of temporary SonarQube resources associated with a plugin.
type TempResource struct {
	PluginName         string             `yaml:"plugin_name"`
	Group              Group              `yaml:"group"`
	User               User               `yaml:"user"`
	GlobalPermissions  []string           `yaml:"global_permissions"`
	Projects           []Project          `yaml:"projects"`
	PermissionTemplate PermissionTemplate `yaml:"permission_template"`
}

// Group represents a SonarQube user group.
type Group struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// User represents a SonarQube user account.
type User struct {
	Login    string   `yaml:"login"`
	Name     string   `yaml:"name"`
	Email    string   `yaml:"email"`
	Password string   `yaml:"password"`
	Groups   []string `yaml:"groups"`
}

// Project represents a SonarQube project.
type Project struct {
	Key        string `yaml:"key"`
	Name       string `yaml:"name"`
	Visibility string `yaml:"visibility"`
}

// PermissionTemplate represents a SonarQube permission template.
type PermissionTemplate struct {
	Name              string   `yaml:"name"`
	Description       string   `yaml:"description"`
	ProjectKeyPattern string   `yaml:"project_key_pattern"`
	Permissions       []string `yaml:"permissions"`
}
