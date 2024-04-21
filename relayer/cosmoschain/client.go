package cosmoschain

import (
	"fmt"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/light"
	comethttp "github.com/cometbft/cometbft/light/provider/http"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	tmclient "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"time"
)

// DefaultUpgradePath is the default IBC upgrade path set for an on-chain light client
var defaultUpgradePath = []string{"upgrade", "upgradedIBCState"}

func (cc *CosmosChain) CreateClient(clientState ibcexported.ClientState, consensusState ibcexported.ConsensusState) (string, error) {
	address, err := cc.getAddress()
	if err != nil {
		return "", err
	}

	msg, err := clienttypes.NewMsgCreateClient(clientState, consensusState, address)
	if err != nil {
		return "", err
	}

	txHash, err := cc.sendTx(msg)
	if err != nil {
		return "", err
	}

	txResp, err := authtx.QueryTx(cc.clientCtx, txHash)
	if err != nil {
		return "", err
	}

	return parseClientIDFromEvents(txResp.Events)
}

// Repurposed from cosmos relayer
func parseClientIDFromEvents(events []abcitypes.Event) (string, error) {
	for _, event := range events {
		if event.Type == clienttypes.EventTypeCreateClient {
			for _, attr := range event.Attributes {
				if attr.Key == clienttypes.AttributeKeyClientID {
					return attr.Value, nil
				}
			}
		}
	}

	return "", fmt.Errorf("client identifier event attribute not found")
}

func (cc *CosmosChain) GetCreateClientInfo() (ibcexported.ClientState, ibcexported.ConsensusState, error) {
	height, err := cc.getLatestHeight()
	if err != nil {
		return nil, nil, err
	}

	ibcHeader, err := cc.getIBCHeader(height)
	if err != nil {
		return nil, nil, err
	}

	clientState, err := cc.getClientState(ibcHeader)
	if err != nil {
		return nil, nil, err
	}

	return clientState, ibcHeader.ConsensusState(), nil
}

func (cc *CosmosChain) getClientState(ibcHeader TendermintIBCHeader) (ibcexported.ClientState, error) {
	revisionNumber := clienttypes.ParseChainID(cc.clientCtx.ChainID)

	unbondingPeriod, err := cc.getUnbondingPeriod()
	if err != nil {
		return nil, err
	}

	return &tmclient.ClientState{
		ChainId:         cc.clientCtx.ChainID,
		TrustLevel:      tmclient.NewFractionFromTm(light.DefaultTrustLevel),
		TrustingPeriod:  time.Duration(int64(unbondingPeriod) / 100 * 85),
		UnbondingPeriod: unbondingPeriod,
		MaxClockDrift:   10 * time.Minute,
		FrozenHeight:    clienttypes.ZeroHeight(),
		LatestHeight: clienttypes.Height{
			RevisionNumber: revisionNumber,
			RevisionHeight: ibcHeader.Height(),
		},
		ProofSpecs:  commitmenttypes.GetSDKSpecs(),
		UpgradePath: defaultUpgradePath,
	}, nil
}

func (cc *CosmosChain) getLatestHeight() (int64, error) {
	stat, err := cc.clientCtx.Client.Status(cc.clientCtx.CmdContext)
	if err != nil {
		return -1, err
	} else if stat.SyncInfo.CatchingUp {
		return -1, fmt.Errorf("node at %s running chain %s not caught up", cc.clientCtx.NodeURI, cc.clientCtx.ChainID)
	}
	return stat.SyncInfo.LatestBlockHeight, nil
}

func (cc *CosmosChain) getIBCHeader(height int64) (TendermintIBCHeader, error) {
	if height <= 0 {
		return TendermintIBCHeader{}, fmt.Errorf("height cannot be 0 or less")
	}

	provider, err := comethttp.New(cc.clientCtx.ChainID, cc.clientCtx.NodeURI)
	if err != nil {
		return TendermintIBCHeader{}, err
	}

	lightBlock, err := provider.LightBlock(cc.clientCtx.CmdContext, height)
	if err != nil {
		return TendermintIBCHeader{}, err
	}

	return TendermintIBCHeader{
		SignedHeader: lightBlock.SignedHeader,
		ValidatorSet: lightBlock.ValidatorSet,
	}, nil
}

func (cc *CosmosChain) getUnbondingPeriod() (time.Duration, error) {
	queryClient := stakingtypes.NewQueryClient(cc.clientCtx)
	res, err := queryClient.Params(cc.clientCtx.CmdContext, &stakingtypes.QueryParamsRequest{})
	if err != nil {
		return 0, err
	}

	return res.Params.UnbondingTime, nil
}
