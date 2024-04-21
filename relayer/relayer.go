package relayer

import (
	"context"
	"github.com/gjermundgaraba/solo-machine-go/relayer/cosmoschain"
	"github.com/gjermundgaraba/solo-machine-go/solomachine"
	"go.uber.org/zap"
)

type Relayer struct {
	logger *zap.Logger
	config *Config

	soloMachine *solomachine.SoloMachine
	cosmosChain *cosmoschain.CosmosChain
}

func NewRelayer(cmdCtx context.Context, logger *zap.Logger, config *Config, homedir string) *Relayer {
	solo := solomachine.NewSoloMachine(logger)

	cosmos := cosmoschain.NewCosmosChain(
		cmdCtx,
		logger,
		homedir,
		config.CosmosChain.AccountPrefix,
		config.CosmosChain.KeyringBackend,
		config.CosmosChain.Key,
		config.CosmosChain.RPCAddr,
		config.CosmosChain.ChainID,
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
