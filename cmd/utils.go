package cmd

import (
	"github.com/gjermundgaraba/solo-machine-go/relayer"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"path"
)

func getConfigPath(homedir string) string {
	return path.Join(homedir, "config.yaml")
}

func getLogger(cmd *cobra.Command) *zap.Logger {
	return cmd.Context().Value(contextKeyLogger).(*zap.Logger)
}

func getRelayerConfig(cmd *cobra.Command) *relayer.Config {
	return cmd.Context().Value(contextKeyConfig).(*relayer.Config)
}

func getHomedir(cmd *cobra.Command) string {
	return cmd.Context().Value(contextKeyHomedir).(string)
}
