package cmd

import (
	"github.com/spf13/cobra"
)

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the relayer",
		RunE: func(cmd *cobra.Command, args []string) error {
			panic("not implemented")

			return nil
		},
	}

	return cmd
}
