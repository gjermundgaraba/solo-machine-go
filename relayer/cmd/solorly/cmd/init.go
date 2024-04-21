package cmd

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/gjermundgaraba/solo-machine-go/relayer"
	"github.com/gjermundgaraba/solo-machine-go/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the relayer",
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, err := cmd.Flags().GetBool(flagVerbose)
			if err != nil {
				return err
			}
			logger := utils.CreateLogger(verbose)

			makeExampleConfig, err := cmd.Flags().GetBool("example")
			if err != nil {
				return err
			}

			var config *relayer.Config
			if makeExampleConfig {
				config = &relayer.Config{
					SoloMachine: relayer.SoloMachineConfig{},
					CosmosChain: relayer.CosmosChainConfig{
						RPCAddr:        "http://0.0.0.0:26657",
						AccountPrefix:  "gg",
						ChainID:        "gg",
						GasAdjustment:  1.5,
						GasPrices:      "0.025stake",
						Gas:            "auto",
						KeyringBackend: "test",
						Key:            "tt",
						SoloMachineLightClient: relayer.SoloMachineLightClientConfig{
							// No point in setting these up unless we also hardcode this into the gg chain
							IBCClientID:  "",
							ConnectionID: "",
							ChannelID:    "",
						},
					},
				}
			}

			force, err := cmd.Flags().GetBool("force")

			homedir, err := cmd.Flags().GetString(flags.FlagHome)
			if err != nil {
				return err
			}
			configPath := getConfigPath(homedir)
			if err := relayer.WriteConfigToFile(config, configPath, force); err != nil {
				return err
			}

			logger.Info("Initialized relayer config file", zap.String("path", configPath))

			return nil
		},
	}

	cmd.Flags().Bool("example", false, "init an example config instead of an empty one")
	cmd.Flags().Bool("force", false, "forcefully overwrite, even if an existing one exists")

	return cmd
}
