package cmd

import (
	"context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/gjermundgaraba/solo-machine-go/relayer"
	"github.com/gjermundgaraba/solo-machine-go/utils"
	"github.com/spf13/cobra"
	"os"
	"path"
	"path/filepath"
)

const (
	// Persistent Flags
	flagVerbose = "verbose"
	flagForce   = "force"

	contextKeyLogger  = "logger"
	contextKeyConfig  = "config"
	contextKeyHomedir = "homedir"
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "solo-machine",
		Short: "TODO", // TODO
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Use == "init" {
				return nil
			}

			if cmd.Context() == nil {
				cmd.SetContext(context.Background())
			}

			verbose, err := cmd.Flags().GetBool(flagVerbose)
			if err != nil {
				return err
			}
			logger := utils.CreateLogger(verbose)
			cmd.SetContext(context.WithValue(cmd.Context(), contextKeyLogger, logger))

			solorlyDir, err := cmd.Flags().GetString(flags.FlagHome)
			if err != nil {
				return err
			}

			cmd.SetContext(context.WithValue(cmd.Context(), contextKeyHomedir, solorlyDir))

			configPath := path.Join(solorlyDir, "config.yaml")

			config, err := relayer.ReadConfigFromFile(logger, configPath)
			if err != nil {
				return err
			}
			cmd.SetContext(context.WithValue(cmd.Context(), contextKeyConfig, config))

			return nil
		},
	}

	cmd.AddCommand(keys.Commands())
	cmd.AddCommand(StartCmd())
	cmd.AddCommand(InitCmd())
	cmd.AddCommand(LinkCmd())

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	solorlyDir := filepath.Join(userHomeDir, ".solorly")

	cmd.PersistentFlags().String(flags.FlagHome, solorlyDir, "solorly home directory")
	cmd.PersistentFlags().Bool(flagVerbose, false, "Enable verbose output")

	return cmd
}
