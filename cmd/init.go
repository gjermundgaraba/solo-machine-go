package cmd

import (
	"github.com/gjermundgaraba/solo-machine-go/relayer"
	"github.com/gjermundgaraba/solo-machine-go/solomachine"
	"github.com/gjermundgaraba/solo-machine-go/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize everything (light clients, connections, channels) for all chains in the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := getLogger(cmd)
			homedir := getHomedir(cmd)
			config := getConfig(cmd)
			chainName := getChainName(cmd)
			cdc := utils.SetupCodec()

			r, err := relayer.NewRelayer(cmd.Context(), logger, cdc, config, homedir)
			if err != nil {
				return err
			}

			sm := solomachine.NewSoloMachine(logger, cdc, r, homedir)
			if !sm.CounterpartyLightClientExists(chainName) {
				if err := sm.CreateCounterpartyLightClient(chainName); err != nil {
					return err
				}
				logger.Info("Counterparty light client created", zap.String("chain", chainName))
			} else {
				logger.Info("Counterparty light client already exists", zap.String("chain", chainName))
				if err := sm.UpdateCounterpartyLightClient(chainName); err != nil {
					return err
				}
			}

			if !sm.LightClientExists(chainName) {
				if err := sm.CreateLightClient(chainName); err != nil {
					return err
				}
				logger.Info("Light client created", zap.String("chain", chainName))
			} else {
				logger.Info("Light client already exists", zap.String("chain", chainName))
				if err := sm.UpdateLightClient(chainName); err != nil {
					return err
				}
			}

			// Connection and channel creation is safe to call multiple times as it checks if it exists and the states
			// It will also continue if the handshake is started but not completed
			if err := sm.CreateConnection(chainName); err != nil {
				return err
			}

			if err := sm.CreateICS20Channel(chainName); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
