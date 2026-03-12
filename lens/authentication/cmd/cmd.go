package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	scmd "github.com/rhine-tech/scene/scenes/cmd"
	"github.com/spf13/cobra"
)

type app struct {
	auth  authentication.IAuthenticationService `aperture:""`
	token authentication.IAccessTokenService    `aperture:""`
}

func NewCmdApp() scmd.CmdApp {
	return &app{}
}

func (a *app) Name() scene.ImplName {
	return authentication.Lens.ImplNameNoVer("CmdApp")
}

func (a *app) Command(rootCmd *cobra.Command) error {
	root := &cobra.Command{
		Use:     "authenticate",
		Aliases: []string{"authentication", "auth"},
		Short:   "Manage users and access tokens",
	}
	root.AddCommand(
		newCreateUserCmd(a),
		newDeleteUserCmd(a),
		newGetUserCmd(a),
		newListUsersCmd(a),
		newCreateTokenCmd(a),
	)
	rootCmd.AddCommand(root)
	return nil
}

func newCreateUserCmd(app *app) *cobra.Command {
	return &cobra.Command{
		Use:     "create_user <username> <password>",
		Aliases: []string{"create-user"},
		Short:   "Create a user",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := app.auth.AddUser(args[0], args[1])
			if err != nil {
				return err
			}
			return printJSON(cmd, user)
		},
	}
}

func newDeleteUserCmd(app *app) *cobra.Command {
	return &cobra.Command{
		Use:     "delete_user <user_id>",
		Aliases: []string{"delete-user"},
		Short:   "Delete a user by user ID",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.auth.DeleteUser(args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "ok")
			return nil
		},
	}
}

func newGetUserCmd(app *app) *cobra.Command {
	return &cobra.Command{
		Use:   "get_user <user_id>",
		Short: "Get a user by user ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := app.auth.UserById(args[0])
			if err != nil {
				return err
			}
			return printJSON(cmd, user)
		},
	}
}

func newListUsersCmd(app *app) *cobra.Command {
	var offset int64
	var limit int64
	cmd := &cobra.Command{
		Use:     "list_users",
		Aliases: []string{"list-users"},
		Short:   "List users",
		RunE: func(cmd *cobra.Command, args []string) error {
			users, err := app.auth.ListUsers(offset, limit)
			if err != nil {
				return err
			}
			return printJSON(cmd, users)
		},
	}
	cmd.Flags().Int64Var(&offset, "offset", 0, "pagination offset")
	cmd.Flags().Int64Var(&limit, "limit", 20, "pagination limit")
	return cmd
}

func newCreateTokenCmd(app *app) *cobra.Command {
	var name string
	var expireAt int64
	cmd := &cobra.Command{
		Use:   "create_token <user_id>",
		Short: "Create an access token for a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := app.token.Create(args[0], name, expireAt)
			if err != nil {
				return err
			}
			return printJSON(cmd, token)
		},
	}
	cmd.Flags().StringVar(&name, "name", "cli", "token display name")
	cmd.Flags().Int64Var(&expireAt, "expire-at", -1, "token expiration unix timestamp, -1 means never")
	return cmd
}

func printJSON(cmd *cobra.Command, value any) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
