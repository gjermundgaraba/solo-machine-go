package cosmoschain

import (
	"github.com/cosmos/cosmos-sdk/client"
	"go.uber.org/zap"
)

type CosmosChain struct {
	logger  *zap.Logger
	homedir string

	clientCtx     client.Context
	gasAdjustment float64
	gasPrices     string
	gas           string
}

func NewCosmosChain(
	clientCtx client.Context,
	logger *zap.Logger,
	homedir string,
	gasAdjustment float64,
	gasPrices string,
	gas string,
) *CosmosChain {
	return &CosmosChain{
		logger:        logger,
		homedir:       homedir,
		clientCtx:     clientCtx,
		gasAdjustment: gasAdjustment,
		gasPrices:     gasPrices,
		gas:           gas,
	}
}
