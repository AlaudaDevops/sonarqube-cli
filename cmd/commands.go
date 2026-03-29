package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/client"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a project",
	RunE:  runProjectDelete,
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	RunE:  runUserDelete,
}

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage groups",
}

var groupDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a group",
	RunE:  runGroupDelete,
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage tokens",
}

var tokenRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke a token",
	RunE:  runTokenRevoke,
}

var (
	endpoint string
	token    string
	key      string
	login    string
	name     string
)

func init() {
	projectCmd.AddCommand(projectDeleteCmd)
	userCmd.AddCommand(userDeleteCmd)
	groupCmd.AddCommand(groupDeleteCmd)
	tokenCmd.AddCommand(tokenRevokeCmd)

	for _, cmd := range []*cobra.Command{projectDeleteCmd, userDeleteCmd, groupDeleteCmd, tokenRevokeCmd} {
		cmd.Flags().StringVar(&endpoint, "endpoint", "", "SonarQube endpoint")
		cmd.Flags().StringVar(&token, "token", "", "Manager token")
		cmd.MarkFlagRequired("endpoint")
		cmd.MarkFlagRequired("token")
	}

	projectDeleteCmd.Flags().StringVar(&key, "key", "", "Project key")
	projectDeleteCmd.MarkFlagRequired("key")

	userDeleteCmd.Flags().StringVar(&login, "login", "", "User login")
	userDeleteCmd.MarkFlagRequired("login")

	groupDeleteCmd.Flags().StringVar(&name, "name", "", "Group name")
	groupDeleteCmd.MarkFlagRequired("name")

	tokenRevokeCmd.Flags().StringVar(&login, "login", "", "User login")
	tokenRevokeCmd.Flags().StringVar(&name, "name", "", "Token name")
	tokenRevokeCmd.MarkFlagRequired("login")
	tokenRevokeCmd.MarkFlagRequired("name")
}

func runProjectDelete(cmd *cobra.Command, args []string) error {
	c := client.New(endpoint, token)
	if err := c.DeleteProject(key); err != nil {
		return err
	}
	fmt.Printf("Deleted project: %s\n", key)
	return nil
}

func runUserDelete(cmd *cobra.Command, args []string) error {
	c := client.New(endpoint, token)
	if err := c.DeleteUser(login); err != nil {
		return err
	}
	fmt.Printf("Deleted user: %s\n", login)
	return nil
}

func runGroupDelete(cmd *cobra.Command, args []string) error {
	c := client.New(endpoint, token)
	if err := c.DeleteGroup(name); err != nil {
		return err
	}
	fmt.Printf("Deleted group: %s\n", name)
	return nil
}

func runTokenRevoke(cmd *cobra.Command, args []string) error {
	c := client.New(endpoint, token)
	if err := c.RevokeUserToken(login, name); err != nil {
		return err
	}
	fmt.Printf("Revoked token: %s for user: %s\n", name, login)
	return nil
}
