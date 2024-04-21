package cosmoschain

import (
	"context"
	"errors"
	comettypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	tendermint "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"go.uber.org/zap"
)

type TendermintIBCHeader struct {
	SignedHeader      *comettypes.SignedHeader
	ValidatorSet      *comettypes.ValidatorSet
	TrustedValidators *comettypes.ValidatorSet
	TrustedHeight     clienttypes.Height
}

type CosmosChain struct {
	logger  *zap.Logger
	homedir string

	clientCtx     client.Context
	gasAdjustment float64
	gasPrices     string
	gas           string
}

func NewCosmosChain(
	cmdCtx context.Context,
	logger *zap.Logger,
	homedir string,
	accountPrefix string,
	keyringBackend string,
	key string,
	rpc string,
	chainID string,
	gasAdjustment float64,
	gasPrices string,
	gas string,
) *CosmosChain {
	clientCtx, err := setupClientContext(
		cmdCtx,
		homedir,
		accountPrefix,
		keyringBackend,
		key,
		rpc,
		chainID)

	if err != nil {
		logger.Error("Error setting up client context", zap.Error(err))
		panic(err)
	}

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

func (h TendermintIBCHeader) Height() uint64 {
	return uint64(h.SignedHeader.Height)
}

func (h TendermintIBCHeader) ConsensusState() ibcexported.ConsensusState {
	return &tendermint.ConsensusState{
		Timestamp:          h.SignedHeader.Time,
		Root:               commitmenttypes.NewMerkleRoot(h.SignedHeader.AppHash),
		NextValidatorsHash: h.SignedHeader.NextValidatorsHash,
	}
}
