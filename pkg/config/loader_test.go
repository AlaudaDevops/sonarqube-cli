package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	content := `sonarqube:
  endpoint: https://sonarqube.example.com
  manager:
    username: admin
    token: test-token
  temp_resources:
    - plugin_name: tektoncd
      group:
        name: test-group
        description: Test group
      user:
        login: test-user
        name: Test User
        email: test@example.com
        password: password
        groups:
          - test-group
      global_permissions:
        - scan
      projects:
        - key: test-project
          name: Test Project
          visibility: private
      permission_template:
        name: test-template
        description: Test template
        project_key_pattern: "test-.*"
        permissions:
          - user
`

	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.SonarQube.Endpoint != "https://sonarqube.example.com" {
		t.Errorf("endpoint = %v, want https://sonarqube.example.com", cfg.SonarQube.Endpoint)
	}

	if len(cfg.SonarQube.TempResources) != 1 {
		t.Errorf("len(TempResources) = %v, want 1", len(cfg.SonarQube.TempResources))
	}
}

func TestLoad_PreservesTaskRunIDTemplateWhenEnvSet(t *testing.T) {
	t.Setenv("TASK_RUN_ID", "env-task")

	content := `sonarqube:
  endpoint: https://sonarqube.example.com
  manager:
    username: admin
    token: test-token
  temp_resources:
    - plugin_name: tektoncd
      group:
        name: test-${TASK_RUN_ID}
        description: Test group
      user:
        login: test-user-${TASK_RUN_ID}
        name: Test User
        email: test@example.com
        password: password
        groups:
          - test-group
      global_permissions:
        - scan
      projects:
        - key: test-project-${TASK_RUN_ID}
          name: Test Project
          visibility: private
      permission_template:
        name: test-template-${TASK_RUN_ID}
        description: Test template
        project_key_pattern: "test-${TASK_RUN_ID}.*"
        permissions:
          - user
`

	tmpfile, err := os.CreateTemp("", "config-template-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	res := cfg.SonarQube.TempResources[0]
	if res.Group.Name != "test-${TASK_RUN_ID}" {
		t.Fatalf("group.name = %q, want template placeholder preserved", res.Group.Name)
	}
	if res.User.Login != "test-user-${TASK_RUN_ID}" {
		t.Fatalf("user.login = %q, want template placeholder preserved", res.User.Login)
	}
	if res.Projects[0].Key != "test-project-${TASK_RUN_ID}" {
		t.Fatalf("projects[0].key = %q, want template placeholder preserved", res.Projects[0].Key)
	}
}

func TestLoadWithVariables_OverridesSecretPlaceholders(t *testing.T) {
	content := `sonarqube:
  endpoint: https://sonarqube.example.com
  manager:
    username: admin
    token: ${SONARQUBE_MANAGER_TOKEN}
  temp_resources:
    - plugin_name: tektoncd
      group:
        name: test-group
        description: Test group
      user:
        login: test-user
        name: Test User
        email: test@example.com
        password: ${TEMP_USER_PASSWORD}
        groups:
          - test-group
      global_permissions:
        - scan
      projects:
        - key: test-project
          name: Test Project
          visibility: private
      permission_template:
        name: test-template
        description: Test template
        project_key_pattern: "test-.*"
        permissions:
          - user
`

	tmpfile, err := os.CreateTemp("", "config-overrides-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cfg, err := LoadWithVariables(tmpfile.Name(), map[string]string{
		"SONARQUBE_MANAGER_TOKEN": "file-token",
		"TEMP_USER_PASSWORD":      "file-password",
	})
	if err != nil {
		t.Fatalf("LoadWithVariables() error = %v", err)
	}

	if cfg.SonarQube.Manager.Token != "file-token" {
		t.Fatalf("manager.token = %q, want file-token", cfg.SonarQube.Manager.Token)
	}
	if cfg.SonarQube.TempResources[0].User.Password != "file-password" {
		t.Fatalf("user.password = %q, want file-password", cfg.SonarQube.TempResources[0].User.Password)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		token    string
		wantErr  bool
		setup    func()
		cleanup  func()
	}{
		{
			name:     "valid https",
			endpoint: "https://sonarqube.example.com",
			token:    "token",
			wantErr:  false,
		},
		{
			name:     "http by default fail",
			endpoint: "http://sonarqube.example.com",
			token:    "token",
			wantErr:  true,
		},
		{
			name:     "http allowed with env",
			endpoint: "http://sonarqube.example.com",
			token:    "token",
			wantErr:  false,
			setup: func() {
				os.Setenv("SONARQUBE_ALLOW_HTTP", "true")
			},
			cleanup: func() {
				os.Unsetenv("SONARQUBE_ALLOW_HTTP")
			},
		},
		{
			name:     "missing scheme",
			endpoint: "sonarqube.example.com",
			token:    "token",
			wantErr:  true,
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			token:    "token",
			wantErr:  true,
		},
		{
			name:     "missing token",
			endpoint: "https://sonarqube.example.com",
			token:    "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			cfg := &Config{
				SonarQube: SonarQubeConfig{
					Endpoint: tt.endpoint,
					Manager: Manager{
						Token: tt.token,
					},
				},
			}
			if err := cfg.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConnection_IgnoresInvalidTempResources(t *testing.T) {
	content := `sonarqube:
  endpoint: https://sonarqube.example.com
  manager:
    username: admin
    token: test-token
  temp_resources:
    - plugin_name: broken
      group:
        name: ""
      user:
        login: ""
`

	tmpfile, err := os.CreateTemp("", "config-connection-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	endpoint, token, err := LoadConnection(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConnection() error = %v", err)
	}
	if endpoint != "https://sonarqube.example.com" {
		t.Fatalf("LoadConnection() endpoint = %q, want https://sonarqube.example.com", endpoint)
	}
	if token != "test-token" {
		t.Fatalf("LoadConnection() token = %q, want test-token", token)
	}
}
