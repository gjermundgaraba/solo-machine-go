package cmd

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/gjermundgaraba/solo-machine-go/relayer"
	"github.com/gjermundgaraba/solo-machine-go/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

const (
	// Persistent Flags
	flagVerbose   = "verbose"
	flagChainName = "chain-name"

	contextKeyLogger    = "logger"
	contextKeyConfig    = "config"
	contextKeyHomedir   = "homedir"
	contextKeyChainName = "chain-name"
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "solo-machine",
		Short: "TODO", // TODO
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Context() == nil {
				cmd.SetContext(context.Background())
			}

			verbose, err := cmd.Flags().GetBool(flagVerbose)
			if err != nil {
				return err
			}
			logger := utils.CreateLogger(verbose)
			cmd.SetContext(context.WithValue(cmd.Context(), contextKeyLogger, logger))

			homedir, err := cmd.Flags().GetString(flags.FlagHome)
			if err != nil {
				return err
			}
			cmd.SetContext(context.WithValue(cmd.Context(), contextKeyHomedir, homedir))

			chainName, err := cmd.Flags().GetString(flagChainName)
			if err != nil {
				return err
			}
			cmd.SetContext(context.WithValue(cmd.Context(), contextKeyChainName, chainName))
			if chainName == "" {
				return fmt.Errorf("chain name cannot be empty")
			}

			configPath := relayer.GetConfigPath(homedir)
			if !relayer.ConfigExists(configPath) {
				if err := relayer.WriteConfigToFile(relayer.Config{
					Chains: map[string]relayer.ChainConfig{
						chainName: {},
					},
				}, configPath); err != nil {
					return err
				}

				logger.Error("The relayer config does not exist, creating a new empty example one and exiting the command", zap.String("path", configPath))
				return fmt.Errorf("please fill in the config file with the necessary information and run the init command again if you want to set up clients, connections and channels")
			}
			config, err := relayer.ReadConfigFromFile(logger, configPath)
			if err != nil {
				return err
			}
			if err := config.Validate(); err != nil {
				return err
			}
			cmd.SetContext(context.WithValue(cmd.Context(), contextKeyConfig, config))

			chainConfig, ok := config.Chains[chainName]
			if !ok {
				return fmt.Errorf("chain with name %s does not exist in the config file", chainName)
			}

			// The rest here is mostly just needed for the keys command (cosmos sdk stuff)
			cfg := sdk.GetConfig()
			accountPubKeyPrefix := chainConfig.AccountPrefix + "pub"
			validatorAddressPrefix := chainConfig.AccountPrefix + "valoper"
			validatorPubKeyPrefix := chainConfig.AccountPrefix + "valoperpub"
			consNodeAddressPrefix := chainConfig.AccountPrefix + "valcons"
			consNodePubKeyPrefix := chainConfig.AccountPrefix + "valconspub"
			cfg.SetBech32PrefixForAccount(chainConfig.AccountPrefix, accountPubKeyPrefix)
			cfg.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
			cfg.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
			//cfg.Seal()
			cdc := utils.SetupCodec()
			kr, err := utils.GetKeyring(chainConfig.KeyringBackend, homedir, cdc)
			if err != nil {
				panic(err)
			}
			clientCtx := client.Context{}.
				WithCmdContext(cmd.Context()).
				WithCodec(cdc).
				WithInput(os.Stdin).
				WithAccountRetriever(authtypes.AccountRetriever{}).
				WithHomeDir(homedir).
				WithKeyring(kr).
				WithViper("")
			cmd.SetContext(context.WithValue(cmd.Context(), client.ClientContextKey, &clientCtx))

			return nil
		},
	}

	cmd.AddCommand(keys.Commands())
	cmd.AddCommand(StartCmd())
	cmd.AddCommand(InitCmd())

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	defaultHomedir := filepath.Join(userHomeDir, ".solo-machine")

	cmd.PersistentFlags().String(flags.FlagHome, defaultHomedir, "solo-machine home directory")
	cmd.PersistentFlags().Bool(flagVerbose, false, "Enable verbose output")
	cmd.PersistentFlags().String(flagChainName, "", "Chain name")

	if err := cmd.MarkPersistentFlagRequired(flagChainName); err != nil {
		panic(err)
	}

	return cmd
}
