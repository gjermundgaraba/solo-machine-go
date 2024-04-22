package relayer

import (
	"context"
	"github.com/gjermundgaraba/solo-machine-go/relayer/cosmoschain"
	"github.com/gjermundgaraba/solo-machine-go/solomachine"
	"github.com/gjermundgaraba/solo-machine-go/utils"
	"go.uber.org/zap"
)

type Relayer struct {
	logger *zap.Logger
	config *Config

	soloMachine *solomachine.SoloMachine
	cosmosChain *cosmoschain.CosmosChain
}

func NewRelayer(cmdCtx context.Context, logger *zap.Logger, config *Config, homedir string) *Relayer {
	clientCtx, err := utils.SetupClientContext(
		cmdCtx,
		homedir,
		config.CosmosChain.AccountPrefix,
		config.CosmosChain.KeyringBackend,
		config.CosmosChain.Key,
		config.CosmosChain.RPCAddr,
		config.CosmosChain.ChainID)
	if err != nil {
		logger.Error("Error setting up client context", zap.Error(err))
		panic(err)
	}

	solo := solomachine.NewSoloMachine(logger, homedir)

	cosmos := cosmoschain.NewCosmosChain(
		clientCtx,
		logger,
		homedir,
		config.CosmosChain.GasAdjustment,
		config.CosmosChain.GasPrices,
		config.CosmosChain.Gas,
	)

	return &Relayer{
		logger: logger,
		config: config,

		soloMachine: solo,
		cosmosChain: cosmos,
	}
}
