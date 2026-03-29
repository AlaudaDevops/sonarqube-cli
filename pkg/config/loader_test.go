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

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		token    string
		wantErr  bool
	}{
		{
			name:     "valid https",
			endpoint: "https://sonarqube.example.com",
			token:    "token",
			wantErr:  false,
		},
		{
			name:     "http with warning",
			endpoint: "http://sonarqube.example.com",
			token:    "token",
			wantErr:  false,
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
