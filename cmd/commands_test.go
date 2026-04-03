package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConnectionOptionsResolve_PrefersFlagsOverEnv(t *testing.T) {
	t.Setenv("SONARQUBE_URL", "https://env.example.com")
	t.Setenv("SONARQUBE_MANAGER_TOKEN", "env-token")

	endpoint, token, err := (connectionOptions{
		endpoint: "https://flag.example.com",
		token:    "flag-token",
	}).resolve()
	if err != nil {
		t.Fatalf("resolve() error = %v", err)
	}
	if endpoint != "https://flag.example.com" {
		t.Fatalf("endpoint = %q, want flag value", endpoint)
	}
	if token != "flag-token" {
		t.Fatalf("token = %q, want flag value", token)
	}
}

func TestConnectionOptionsResolve_UsesEnvFallback(t *testing.T) {
	t.Setenv("SONARQUBE_URL", "https://env.example.com")
	t.Setenv("SONARQUBE_MANAGER_TOKEN", "env-token")

	endpoint, token, err := (connectionOptions{}).resolve()
	if err != nil {
		t.Fatalf("resolve() error = %v", err)
	}
	if endpoint != "https://env.example.com" {
		t.Fatalf("endpoint = %q, want env value", endpoint)
	}
	if token != "env-token" {
		t.Fatalf("token = %q, want env value", token)
	}
}

func TestConnectionOptionsResolve_UsesManagerTokenFileFallback(t *testing.T) {
	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "manager.token")
	if err := os.WriteFile(tokenPath, []byte("file-token\n"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv("SONARQUBE_URL", "https://env.example.com")
	t.Setenv("SONARQUBE_MANAGER_TOKEN", "env-token")

	endpoint, token, err := (connectionOptions{
		tokenFile: tokenPath,
	}).resolve()
	if err != nil {
		t.Fatalf("resolve() error = %v", err)
	}
	if endpoint != "https://env.example.com" {
		t.Fatalf("endpoint = %q, want env value", endpoint)
	}
	if token != "file-token" {
		t.Fatalf("token = %q, want file-token", token)
	}
}

func TestConnectionOptionsResolve_RejectsEmptyManagerTokenFile(t *testing.T) {
	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "manager.token")
	if err := os.WriteFile(tokenPath, []byte("\n"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv("SONARQUBE_URL", "https://env.example.com")
	t.Setenv("SONARQUBE_MANAGER_TOKEN", "env-token")

	_, _, err := (connectionOptions{
		tokenFile: tokenPath,
	}).resolve()
	if err == nil {
		t.Fatal("resolve() error = nil, want empty manager token file error")
	}
	if !strings.Contains(err.Error(), "is empty") {
		t.Fatalf("resolve() error = %v, want empty file error", err)
	}
}

func TestConnectionOptionsResolve_RequiresEndpoint(t *testing.T) {
	t.Setenv("SONARQUBE_URL", "")
	t.Setenv("SONARQUBE_MANAGER_TOKEN", "env-token")

	_, _, err := (connectionOptions{}).resolve()
	if err == nil {
		t.Fatal("resolve() error = nil, want endpoint validation error")
	}
	if !strings.Contains(err.Error(), "--endpoint or SONARQUBE_URL is required") {
		t.Fatalf("resolve() error = %v, want endpoint validation error", err)
	}
}

func TestConnectionOptionsResolve_RequiresToken(t *testing.T) {
	t.Setenv("SONARQUBE_URL", "https://env.example.com")
	t.Setenv("SONARQUBE_MANAGER_TOKEN", "")

	_, _, err := (connectionOptions{}).resolve()
	if err == nil {
		t.Fatal("resolve() error = nil, want token validation error")
	}
	if !strings.Contains(err.Error(), "--token, --manager-token-file, or SONARQUBE_MANAGER_TOKEN is required") {
		t.Fatalf("resolve() error = %v, want token validation error", err)
	}
}

func TestNewProjectDeleteCmd_ExecutesWithBoundOptions(t *testing.T) {
	var deletedProject string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects/delete" {
			t.Fatalf("path = %s, want /api/projects/delete", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		deletedProject = r.FormValue("project")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cmd := newProjectDeleteCmd()
	cmd.SetArgs([]string{
		"--endpoint", server.URL,
		"--token", "manager-token",
		"--key", "demo-project",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if deletedProject != "demo-project" {
		t.Fatalf("deletedProject = %q, want demo-project", deletedProject)
	}
}

func TestNewTokenRevokeCmd_ExecutesWithBoundOptions(t *testing.T) {
	var login, tokenName string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user_tokens/revoke" {
			t.Fatalf("path = %s, want /api/user_tokens/revoke", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		login = r.FormValue("login")
		tokenName = r.FormValue("name")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cmd := newTokenRevokeCmd()
	cmd.SetArgs([]string{
		"--endpoint", server.URL,
		"--token", "manager-token",
		"--login", "demo-user",
		"--name", "demo-token",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if login != "demo-user" || tokenName != "demo-token" {
		t.Fatalf("got login=%q tokenName=%q, want demo-user/demo-token", login, tokenName)
	}
}
