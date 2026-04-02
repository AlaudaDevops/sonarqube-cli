package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveConfigOverrides_ReadsSecretFiles(t *testing.T) {
	dir := t.TempDir()
	managerTokenPath := filepath.Join(dir, "manager.token")
	tempPasswordPath := filepath.Join(dir, "temp.password")

	if err := os.WriteFile(managerTokenPath, []byte("manager-token\n"), 0600); err != nil {
		t.Fatalf("WriteFile(managerToken) error = %v", err)
	}
	if err := os.WriteFile(tempPasswordPath, []byte("TempPass123\n"), 0600); err != nil {
		t.Fatalf("WriteFile(tempPassword) error = %v", err)
	}

	overrides, err := resolveConfigOverrides(managerTokenPath, tempPasswordPath)
	if err != nil {
		t.Fatalf("resolveConfigOverrides() error = %v", err)
	}
	if overrides["SONARQUBE_MANAGER_TOKEN"] != "manager-token" {
		t.Fatalf("SONARQUBE_MANAGER_TOKEN = %q, want manager-token", overrides["SONARQUBE_MANAGER_TOKEN"])
	}
	if overrides["TEMP_USER_PASSWORD"] != "TempPass123" {
		t.Fatalf("TEMP_USER_PASSWORD = %q, want TempPass123", overrides["TEMP_USER_PASSWORD"])
	}
}

func TestReadSecretFile_RejectsEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.secret")

	if err := os.WriteFile(path, []byte("\n"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := readSecretFile(path)
	if err == nil {
		t.Fatal("readSecretFile() error = nil, want empty file error")
	}
}
