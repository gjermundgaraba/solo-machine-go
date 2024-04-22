package cosmoschain

import (
	"errors"
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

func (cc *CosmosChain) getAddress() (string, error) {
	address := cc.clientCtx.From
	if address == "" {
		return "", errors.New("cosmos chain address is empty in clientCtx")
	}

	return address, nil
}
