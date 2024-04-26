package relayer

import (
	"fmt"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	solomachineclient "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	"time"
)

func (r *Relayer) CreateClient(chainName string, clientState ibcexported.ClientState, consensusState ibcexported.ConsensusState) (string, error) {
	clientCtx := r.createClientCtx(chainName)
	txf := r.createTxFactory(clientCtx, chainName)

	msg, err := clienttypes.NewMsgCreateClient(clientState, consensusState, clientCtx.From)
	if err != nil {
		return "", err
	}

	txResp, err := r.sendTx(clientCtx, txf, msg)
	if err != nil {
		return "", err
	}

	return parseClientIDFromEvents(txResp.Events)
}

func (r *Relayer) GetUnbondingPeriod(chainName string) (time.Duration, error) {
	clientCtx := r.createClientCtx(chainName)
	queryClient := stakingtypes.NewQueryClient(clientCtx)
	res, err := queryClient.Params(clientCtx.CmdContext, &stakingtypes.QueryParamsRequest{})
	if err != nil {
		return 0, err
	}

	return res.Params.UnbondingTime, nil
}

func (r *Relayer) UpdateClient(chainName string, clientID string, clientMsg ibcexported.ClientMessage) error {
	clientCtx := r.createClientCtx(chainName)
	txf := r.createTxFactory(clientCtx, chainName)
	msg, err := clienttypes.NewMsgUpdateClient(clientID, clientMsg, clientCtx.From)
	if err != nil {
		return err
	}

	_, err = r.sendTx(clientCtx, txf, msg)
	return err
}

func (r *Relayer) GetClientState(chainName string, clientID string) (*solomachineclient.ClientState, error) {
	clientCtx := r.createClientCtx(chainName)
	queryClient := clienttypes.NewQueryClient(clientCtx)
	res, err := queryClient.ClientState(clientCtx.CmdContext, &clienttypes.QueryClientStateRequest{ClientId: clientID})
	if err != nil {
		return nil, err
	}

	clientStateUnpacked, err := clienttypes.UnpackClientState(res.ClientState)
	if err != nil {
		return nil, err
	}

	clientState, ok := clientStateUnpacked.(*solomachineclient.ClientState)
	if !ok {
		return nil, fmt.Errorf("cannot convert %T into %T", clientStateUnpacked, clientState)
	}

	return clientState, nil
}
