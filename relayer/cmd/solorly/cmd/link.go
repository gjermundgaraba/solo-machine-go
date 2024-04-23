package cmd

import (
	"github.com/gjermundgaraba/solo-machine-go/relayer"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func LinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Set up all the clients, connections and channels",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := getLogger(cmd)
			config := getRelayerConfig(cmd)
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}
			r := relayer.NewRelayer(cmd.Context(), logger, getRelayerConfig(cmd), getHomedir(cmd))

			if force || config.CosmosChain.SoloMachineLightClient.IBCClientID == "" {
				if err := r.CreateSoloMachineLightClientOnCosmos(); err != nil {
					return err
				}
			} else {
				logger.Info("Skipping creation of solo machine light client on cosmos chain as we already have one configured", zap.String("client-id", config.CosmosChain.SoloMachineLightClient.IBCClientID))
			}

			existsOnSoloClient, err := r.TendermintLightClientExistsOnSoloMachine()
			if err != nil {
				return err
			}
			if force || !existsOnSoloClient {
				if err := r.CreateTendermintLightClientOnSoloMachine(); err != nil {
					return err
				}
			} else {
				logger.Info("Skipping creation of tendermint light client on solo machine as it already exists")
			}

			if force || config.CosmosChain.SoloMachineLightClient.ConnectionID != "" {
				if err := r.InitConnection(); err != nil {
					return err
				}
			} else {
				logger.Info("Skipping creation of connection on cosmos chain as we already have one configured", zap.String("connection-id", config.CosmosChain.SoloMachineLightClient.ConnectionID))
			}

			if err := r.FinishAnyRemainingConnectionHandshakes(); err != nil {
				return err
			}
			logger.Info("Connections are ready")

			// TODO: Create channel on cosmos
			// TODO: Create channel on solo machine (?)

			return nil
		},
	}

	cmd.Flags().Bool(flagForce, false, "recreate clients, connections and channels even if they already exist")

	return cmd
}
