# SonarQube CLI Development Guide

## Build and Test Commands
- Build: `go build -o sonarqube-cli ./cmd` (Or use `make build`)
- Test: `go test -v ./...`
- Clean: `rm -f sonarqube-cli` (Or use `make clean`)

## Code Style & Standards
- Follow standard Go project layout:
    - `cmd/`: CLI entry point and command definitions.
    - `pkg/`: Library code that can be used by other projects.
- Use `spf13/cobra` for command-line interface.
- Use `spf13/pflag` for consistent flag handling.
- Keep `main` package in `cmd/` minimal; delegate logic to `pkg/`.

## Project Structure
- `pkg/client`: API client implementation for SonarQube.
- `pkg/config`: Configuration loading and type definitions.
- `templates/`: Example configuration and templates.
