package cmd

import (
	"fmt"
	"github.com/gjermundgaraba/solo-machine-go/relayer"
	"github.com/spf13/cobra"
)

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the relayer",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := getLogger(cmd)
			config := getRelayerConfig(cmd)
			homedir := getHomedir(cmd)

			logger.Info("Starting relayer...")
			r := relayer.NewRelayer(cmd.Context(), logger, config, homedir)

			// Check if light client exists
			if config.CosmosChain.SoloMachineLightClient.IBCClientID == "" {
				return fmt.Errorf("empty light client in the config, run the link command to create a new one or update the config with an existing one")
			}
			existsOnCosmos, err := r.SoloMachineLightClientExistsOnCosmos()
			if err != nil {
				return err
			}
			if !existsOnCosmos {
				return fmt.Errorf("light client with ID %s not found on CosmosChain", config.CosmosChain.SoloMachineLightClient.IBCClientID)
			}

			existsOnSoloClient, err := r.TendermintLightClientExistsOnSoloMachine()
			if err != nil {
				return err
			}
			if !existsOnSoloClient {
				return fmt.Errorf("light client not found on SoloMachine, run the link command to create a new one")
			}

			return nil
		},
	}

	return cmd
}
