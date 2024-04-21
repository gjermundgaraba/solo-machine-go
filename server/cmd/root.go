package cmd

import "github.com/spf13/cobra"

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "solo-machine-server",
		Short: "TODO", // TODO
	}

	cmd.AddCommand(KeysCmd())

	return cmd
}
