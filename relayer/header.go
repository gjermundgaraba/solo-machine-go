package relayer

import (
	"fmt"
	comethttp "github.com/cometbft/cometbft/light/provider/http"
	"github.com/cosmos/cosmos-sdk/client"
	tmclient "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
)

func (r *Relayer) GetLatestIBCHeader(chainName string) (tmclient.Header, error) {
	clientCtx := r.createClientCtx(chainName)

	height, err := r.getLatestHeight(clientCtx)
	if err != nil {
		return tmclient.Header{}, err
	}

	provider, err := comethttp.New(clientCtx.ChainID, clientCtx.NodeURI)
	if err != nil {
		return tmclient.Header{}, err
	}

	lightBlock, err := provider.LightBlock(clientCtx.CmdContext, height)
	if err != nil {
		return tmclient.Header{}, err
	}

	protoSignedHeader := lightBlock.SignedHeader.ToProto()
	protoValidatorSet, err := lightBlock.ValidatorSet.ToProto()
	if err != nil {
		return tmclient.Header{}, err
	}

	return tmclient.Header{
		SignedHeader: protoSignedHeader,
		ValidatorSet: protoValidatorSet,
	}, nil
}

func (r *Relayer) getLatestHeight(clientCtx client.Context) (int64, error) {
	stat, err := clientCtx.Client.Status(clientCtx.CmdContext)
	if err != nil {
		return -1, err
	} else if stat.SyncInfo.CatchingUp {
		return -1, fmt.Errorf("node at %s running chain %s not caught up", clientCtx.NodeURI, clientCtx.ChainID)
	}

	return stat.SyncInfo.LatestBlockHeight, nil
}
