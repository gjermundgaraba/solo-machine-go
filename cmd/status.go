package cmd

import (
	"github.com/gjermundgaraba/solo-machine/relayer"
	"github.com/gjermundgaraba/solo-machine/solomachine"
	"github.com/gjermundgaraba/solo-machine/utils"
	"github.com/spf13/cobra"
)

func StatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status --chain-name [chain-name]",
		Short: "Print the status of the clients, connections, and channels on the solo machine and chain",
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
			status, err := sm.Status(chainName)
			if err != nil {
				return err
			}

			cmd.Println("Status:")
			cmd.Println("Diversifier:", status.Diversifier)
			cmd.Println("LightClientID:", status.LightClientID)
			cmd.Println("LightClientLatestHeight:", status.LightClientLatestHeight)
			cmd.Println("CounterpartyActualHeight:", status.CounterpartyActualHeight)
			cmd.Println("CounterpartyLightClientID:", status.CounterpartyLightClientID)
			cmd.Println("CounterpartyLightClientSequence:", status.CounterpartyLightClientSequence)
			cmd.Println("ConnectionID:", status.ConnectionID)
			cmd.Println("CounterpartyConnectionID:", status.CounterpartyConnectionID)
			cmd.Println("CounterpartyConnectionState:", status.CounterpartyConnectionState)
			cmd.Println("ICS20ChannelID:", status.ICS20ChannelID)
			cmd.Println("CounterpartyICS20ChannelID:", status.CounterpartyICS20ChannelID)
			cmd.Println("CounterpartyICS20ChannelState:", status.CounterpartyICS20ChannelState)

			return nil
		},
	}
}
