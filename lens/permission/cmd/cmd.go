package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/permission"
	scmd "github.com/rhine-tech/scene/scenes/cmd"
	"github.com/spf13/cobra"
)

type app struct {
	srv permission.PermissionService `aperture:""`
}

func NewCmdApp() scmd.CmdApp {
	return &app{}
}

func (a *app) Name() scene.ImplName {
	return permission.Lens.ImplNameNoVer("CmdApp")
}

func (a *app) Command(rootCmd *cobra.Command) error {
	root := &cobra.Command{
		Use:   "permission",
		Short: "Manage owner permissions",
	}
	root.AddCommand(
		newAddCmd(a),
		newRemoveCmd(a),
		newListCmd(a),
		newListPermissionTreeCmd(a),
		newHasCmd(a),
	)
	rootCmd.AddCommand(root)
	return nil
}

func newAddCmd(app *app) *cobra.Command {
	return &cobra.Command{
		Use:   "add <owner> <permission>",
		Short: "Grant a permission to an owner",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.srv.AddPermission(args[0], args[1]); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "ok")
			return nil
		},
	}
}

func newRemoveCmd(app *app) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <owner> <permission>",
		Short: "Remove a permission from an owner",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.srv.RemovePermission(args[0], args[1]); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "ok")
			return nil
		},
	}
}

func newListCmd(app *app) *cobra.Command {
	return &cobra.Command{
		Use:   "list <owner>",
		Short: "List all permissions for an owner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			perms := app.srv.ListPermissions(args[0])
			values := make([]string, 0, len(perms))
			for _, perm := range perms {
				values = append(values, perm.String())
			}
			return printJSON(cmd, values)
		},
	}
}

func newListPermissionTreeCmd(app *app) *cobra.Command {
	return &cobra.Command{
		Use:     "list_permissiontree <owner>",
		Aliases: []string{"list-permissiontree", "permissiontree"},
		Short:   "List permissions as a permission tree for an owner",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tree := permission.BuildTree(app.srv.ListPermissions(args[0])...)
			rendered := renderPermissionTree(tree)
			_, err := fmt.Fprint(cmd.OutOrStdout(), rendered)
			return err
		},
	}
}

func newHasCmd(app *app) *cobra.Command {
	return &cobra.Command{
		Use:   "has <owner> <permission>",
		Short: "Check whether an owner has a permission",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), app.srv.HasPermissionStr(args[0], args[1]))
			return nil
		},
	}
}

func printJSON(cmd *cobra.Command, value any) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func renderPermissionTree(tree *permission.PermissionTree) string {
	if tree == nil || tree.Root == nil || len(tree.Root.Children) == 0 {
		return "(empty)\n"
	}
	var b strings.Builder
	renderPermissionNode(&b, tree.Root, "", true)
	return b.String()
}

func renderPermissionNode(b *strings.Builder, node *permission.PermissionNode, prefix string, isRoot bool) {
	keys := make([]string, 0, len(node.Children))
	for key := range node.Children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for idx, key := range keys {
		child := node.Children[key]
		last := idx == len(keys)-1
		connector := "|-- "
		nextPrefix := prefix + "|   "
		if last {
			connector = "`-- "
			nextPrefix = prefix + "    "
		}
		if isRoot {
			connector = ""
			nextPrefix = ""
		}
		_, _ = b.WriteString(prefix + connector + key + "\n")
		renderPermissionNode(b, child, nextPrefix, false)
	}
}
