package cmd

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/gjermundgaraba/solo-machine-go/relayer"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func LinkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "link",
		Short: "Set up all the clients, connections and channels",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := getLogger(cmd)
			config := getRelayerConfig(cmd)
			r := relayer.NewRelayer(cmd.Context(), logger, getRelayerConfig(cmd), getHomedir(cmd))

			homedir, err := cmd.Flags().GetString(flags.FlagHome)
			if err != nil {
				return err
			}
			configPath := getConfigPath(homedir)

			if config.CosmosChain.SoloMachineLightClient.IBCClientID == "" {
				if err := r.CreateSoloMachineLightClientOnCosmos(); err != nil {
					return err
				}

				if err := relayer.WriteConfigToFile(config, configPath, true); err != nil {
					return err
				}
				logger.Info("Config updated with light client ID created on the cosmos chain", zap.String("client-id", config.CosmosChain.SoloMachineLightClient.IBCClientID))
			} else {
				logger.Info("Skipping creation of solo machine light client on cosmos chain as we already have one configured", zap.String("client-id", config.CosmosChain.SoloMachineLightClient.IBCClientID))
			}

			if err := r.CreateTendermintLightClientOnSoloMachine(); err != nil {
				return err
			}

			// TODO: Set up client on solo machine
			// TODO: Create connection on cosmos
			// TODO: Create connection on solo machine (?)
			// TODO: Create channel on cosmos
			// TODO: Create channel on solo machine (?)

			return nil
		},
	}
}
