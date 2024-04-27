package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gjermundgaraba/solo-machine-go/relayer"
	"github.com/gjermundgaraba/solo-machine-go/solomachine"
	"github.com/gjermundgaraba/solo-machine-go/utils"
	"github.com/spf13/cobra"
)

func TransferCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "transfer [sender] [receiver] [amount] --chain-name [chain-name]",
		Short: "Transfer (more like create, honestly) tokens from solo machine to chain over ICS20 channel",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := getLogger(cmd)
			homedir := getHomedir(cmd)
			config := getConfig(cmd)
			chainName := getChainName(cmd)
			cdc := utils.SetupCodec()

			sender := args[0]
			receiver := args[1]
			amountStr := args[2]
			coin, err := sdk.ParseCoinNormalized(amountStr)
			if err != nil {
				return err
			}

			r, err := relayer.NewRelayer(cmd.Context(), logger, cdc, config, homedir)
			if err != nil {
				return err
			}

			sm := solomachine.NewSoloMachine(logger, cdc, r, homedir)
			if err := sm.Transfer(chainName, sender, receiver, coin.Denom, coin.Amount.Uint64()); err != nil {
				return err
			}

			return nil
		},
	}
}
