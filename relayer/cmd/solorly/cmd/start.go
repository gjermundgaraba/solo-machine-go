package cmd

import (
	"github.com/gjermundgaraba/solo-machine-go/relayer"
	"github.com/spf13/cobra"
)

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the relayer",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := getLogger(cmd)

			logger.Info("Starting relayer...")
			r := relayer.NewRelayer(cmd.Context(), logger, getRelayerConfig(cmd), getHomedir(cmd))
			if err := r.CreateSoloMachineLightClientOnCosmos(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
