package cmd

import (
	"github.com/gjermundgaraba/solo-machine-go/relayer"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func getLogger(cmd *cobra.Command) *zap.Logger {
	logger := cmd.Context().Value(contextKeyLogger).(*zap.Logger)
	if logger == nil {
		panic("logger is nil")
	}

	return logger
}

func getHomedir(cmd *cobra.Command) string {
	homedir := cmd.Context().Value(contextKeyHomedir).(string)
	if homedir == "" {
		panic("homedir is empty")
	}

	return homedir
}

func getConfig(cmd *cobra.Command) relayer.Config {
	return cmd.Context().Value(contextKeyConfig).(relayer.Config)
}

func getChainName(cmd *cobra.Command) string {
	chainName := cmd.Context().Value(contextKeyChainName).(string)
	if chainName == "" {
		panic("chain name is empty")
	}

	return chainName
}
