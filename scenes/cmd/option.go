package cmd

import "github.com/spf13/cobra"

type RootOption func(*cobra.Command) error

func WithVersion(version string) RootOption {
	return func(root *cobra.Command) error {
		root.Version = version
		return nil
	}
}
