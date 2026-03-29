package config

type Config struct {
	SonarQube SonarQubeConfig `yaml:"sonarqube"`
}

type SonarQubeConfig struct {
	Endpoint      string         `yaml:"endpoint"`
	Manager       Manager        `yaml:"manager"`
	TempResources []TempResource `yaml:"temp_resources"`
}

type Manager struct {
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
}

type TempResource struct {
	PluginName         string             `yaml:"plugin_name"`
	Group              Group              `yaml:"group"`
	User               User               `yaml:"user"`
	GlobalPermissions  []string           `yaml:"global_permissions"`
	Projects           []Project          `yaml:"projects"`
	PermissionTemplate PermissionTemplate `yaml:"permission_template"`
}

type Group struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type User struct {
	Login    string   `yaml:"login"`
	Name     string   `yaml:"name"`
	Email    string   `yaml:"email"`
	Password string   `yaml:"password"`
	Groups   []string `yaml:"groups"`
}

type Project struct {
	Key        string `yaml:"key"`
	Name       string `yaml:"name"`
	Visibility string `yaml:"visibility"`
}

type PermissionTemplate struct {
	Name              string   `yaml:"name"`
	Description       string   `yaml:"description"`
	ProjectKeyPattern string   `yaml:"project_key_pattern"`
	Permissions       []string `yaml:"permissions"`
}
