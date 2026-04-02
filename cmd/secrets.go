package main

import (
	"fmt"
	"os"
	"strings"
)

// readSecretFile reads a single secret value from disk and trims trailing line breaks.
func readSecretFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read secret file %s: %w", path, err)
	}
	// Secret files often end with a trailing newline from Kubernetes or shell tooling.
	secret := strings.TrimRight(string(data), "\r\n")
	if secret == "" {
		return "", fmt.Errorf("secret file %s is empty", path)
	}
	return secret, nil
}

// resolveConfigOverrides loads supported secret placeholders from files for config expansion.
func resolveConfigOverrides(managerTokenPath, tempPasswordPath string) (map[string]string, error) {
	overrides := make(map[string]string, 2)

	if managerTokenPath != "" {
		// Limit file-based overrides to the supported placeholders so config expansion stays predictable.
		token, err := readSecretFile(managerTokenPath)
		if err != nil {
			return nil, err
		}
		overrides["SONARQUBE_MANAGER_TOKEN"] = token
	}

	if tempPasswordPath != "" {
		// Load the temporary password from file for the same reason as manager tokens: avoid secret values in env or args.
		password, err := readSecretFile(tempPasswordPath)
		if err != nil {
			return nil, err
		}
		overrides["TEMP_USER_PASSWORD"] = password
	}

	return overrides, nil
}
