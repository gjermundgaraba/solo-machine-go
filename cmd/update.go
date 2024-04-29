package cmd

import (
	"github.com/gjermundgaraba/solo-machine/relayer"
	"github.com/gjermundgaraba/solo-machine/solomachine"
	"github.com/gjermundgaraba/solo-machine/utils"
	"github.com/spf13/cobra"
)

func UpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update --chain-name",
		Short: "Update the light clients on both sides",
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
			if err := sm.UpdateCounterpartyLightClient(chainName); err != nil {
				return err
			}
			if err := sm.UpdateLightClient(chainName); err != nil {
				return err
			}

			return nil
		},
	}
}
