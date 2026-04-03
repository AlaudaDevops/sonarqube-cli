package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/client"
)

type connectionOptions struct {
	endpoint  string
	token     string
	tokenFile string
}

// resolve merges connection inputs from flags, secret files, and env vars with deterministic precedence.
func (o connectionOptions) resolve() (string, string, error) {
	endpoint := o.endpoint
	if endpoint == "" {
		// Fall back to env vars so non-interactive callers do not need to repeat shared connection flags.
		endpoint = os.Getenv("SONARQUBE_URL")
	}
	if endpoint == "" {
		return "", "", fmt.Errorf("--endpoint or SONARQUBE_URL is required")
	}

	token := o.token
	if token == "" && o.tokenFile != "" {
		var err error
		// Prefer the file when provided to avoid exposing long-lived secrets in process lists or shell history.
		token, err = readSecretFile(o.tokenFile)
		if err != nil {
			return "", "", err
		}
	}
	if token == "" {
		// Keep env fallback last so explicit flags and mounted secret files always win.
		token = os.Getenv("SONARQUBE_MANAGER_TOKEN")
	}
	if token == "" {
		return "", "", fmt.Errorf("--token, --manager-token-file, or SONARQUBE_MANAGER_TOKEN is required")
	}

	return endpoint, token, nil
}

// bindConnectionFlags attaches common SonarQube connection flags to a command.
func bindConnectionFlags(cmd *cobra.Command, opts *connectionOptions) {
	cmd.Flags().StringVar(&opts.endpoint, "endpoint", "", "SonarQube endpoint (fallback to SONARQUBE_URL)")
	cmd.Flags().StringVar(&opts.token, "token", "", "Manager token (fallback to SONARQUBE_MANAGER_TOKEN)")
	cmd.Flags().StringVar(&opts.tokenFile, "manager-token-file", "", "Path to file containing manager token (preferred over SONARQUBE_MANAGER_TOKEN)")
}

type projectDeleteOptions struct {
	connectionOptions
	key string
}

type userDeleteOptions struct {
	connectionOptions
	login string
}

type groupDeleteOptions struct {
	connectionOptions
	name string
}

type tokenRevokeOptions struct {
	connectionOptions
	login string
	name  string
}

var (
	projectCmd = newProjectCmd()
	userCmd    = newUserCmd()
	groupCmd   = newGroupCmd()
	tokenCmd   = newTokenCmd()
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
	}
	cmd.AddCommand(newProjectDeleteCmd())
	return cmd
}

func newProjectDeleteCmd() *cobra.Command {
	opts := &projectDeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectDelete(*opts)
		},
	}
	bindConnectionFlags(cmd, &opts.connectionOptions)
	cmd.Flags().StringVar(&opts.key, "key", "", "Project key")
	cmd.MarkFlagRequired("key")
	return cmd
}

func newUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
	}
	cmd.AddCommand(newUserDeleteCmd())
	return cmd
}

func newUserDeleteCmd() *cobra.Command {
	opts := &userDeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserDelete(*opts)
		},
	}
	bindConnectionFlags(cmd, &opts.connectionOptions)
	cmd.Flags().StringVar(&opts.login, "login", "", "User login")
	cmd.MarkFlagRequired("login")
	return cmd
}

func newGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage groups",
	}
	cmd.AddCommand(newGroupDeleteCmd())
	return cmd
}

func newGroupDeleteCmd() *cobra.Command {
	opts := &groupDeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a group",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGroupDelete(*opts)
		},
	}
	bindConnectionFlags(cmd, &opts.connectionOptions)
	cmd.Flags().StringVar(&opts.name, "name", "", "Group name")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage tokens",
	}
	cmd.AddCommand(newTokenRevokeCmd())
	return cmd
}

func newTokenRevokeCmd() *cobra.Command {
	opts := &tokenRevokeOptions{}
	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a token",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTokenRevoke(*opts)
		},
	}
	bindConnectionFlags(cmd, &opts.connectionOptions)
	cmd.Flags().StringVar(&opts.login, "login", "", "User login")
	cmd.Flags().StringVar(&opts.name, "name", "", "Token name")
	cmd.MarkFlagRequired("login")
	cmd.MarkFlagRequired("name")
	return cmd
}

func runProjectDelete(opts projectDeleteOptions) error {
	endpoint, token, err := opts.connectionOptions.resolve()
	if err != nil {
		return err
	}
	c := client.New(endpoint, token)
	if err := c.DeleteProject(opts.key); err != nil {
		return err
	}
	fmt.Printf("Deleted project: %s\n", opts.key)
	return nil
}

func runUserDelete(opts userDeleteOptions) error {
	endpoint, token, err := opts.connectionOptions.resolve()
	if err != nil {
		return err
	}
	c := client.New(endpoint, token)
	if err := c.DeleteUser(opts.login); err != nil {
		return err
	}
	fmt.Printf("Deleted user: %s\n", opts.login)
	return nil
}

func runGroupDelete(opts groupDeleteOptions) error {
	endpoint, token, err := opts.connectionOptions.resolve()
	if err != nil {
		return err
	}
	c := client.New(endpoint, token)
	if err := c.DeleteGroup(opts.name); err != nil {
		return err
	}
	fmt.Printf("Deleted group: %s\n", opts.name)
	return nil
}

func runTokenRevoke(opts tokenRevokeOptions) error {
	endpoint, token, err := opts.connectionOptions.resolve()
	if err != nil {
		return err
	}
	c := client.New(endpoint, token)
	if err := c.RevokeUserToken(opts.login, opts.name); err != nil {
		return err
	}
	fmt.Printf("Revoked token: %s for user: %s\n", opts.name, opts.login)
	return nil
}
