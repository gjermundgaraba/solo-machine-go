package relayer

import (
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"go.uber.org/zap"
	"time"
)

func (r *Relayer) QueryConnection(chainName string, connectionID string) (*connectiontypes.ConnectionEnd, error) {
	clientCtx := r.createClientCtx(chainName)
	queryClient := connectiontypes.NewQueryClient(clientCtx)
	req := &connectiontypes.QueryConnectionRequest{
		ConnectionId: connectionID,
	}

	res, err := queryClient.Connection(clientCtx.CmdContext, req)
	if err != nil {
		return nil, err
	}

	return res.Connection, nil
}

func (r *Relayer) InitConnection(chainName string, clientID string, counterpartyClientID string) (string, error) {
	clientCtx := r.createClientCtx(chainName)
	txf := r.createTxFactory(clientCtx, chainName)

	var version *connectiontypes.Version // Can be nil? Not sure.
	merklePrefix := commitmenttypes.NewMerklePrefix([]byte(ibcexported.StoreKey))
	initMsg := connectiontypes.NewMsgConnectionOpenInit(
		clientID,
		counterpartyClientID,
		merklePrefix,
		version,
		0,
		clientCtx.From,
	)

	txResp, err := r.sendTx(clientCtx, txf, initMsg)
	if err != nil {
		return "", err
	}

	connectionID, err := parseConnectionIDFromEvents(txResp.Events)
	if err != nil {
		return "", err
	}

	return connectionID, nil
}

func (r *Relayer) ConnectionOpenAck(
	chainName string,
	connectionID string,
	counterpartyConnectionID string,
	counterpartyClient ibcexported.ClientState,
	tryProof []byte,
	clientProof []byte,
	consensusProof []byte,
	consensusHeight clienttypes.Height,
) error {
	clientCtx := r.createClientCtx(chainName)
	txf := r.createTxFactory(clientCtx, chainName)

	r.logger.Debug("sending connection open ack",
		zap.String("connection-id", connectionID),
		zap.String("counterparty-connection-id", counterpartyConnectionID),
		zap.String("counterparty-client", counterpartyClient.String()),
		zap.String("consensus-height", consensusHeight.String()),
		zap.String("version", connectiontypes.GetCompatibleVersions()[0].String()),
		zap.String("from", clientCtx.From),
	)

	// Just to make sure consensus height is not equal to the current height of the chain
	time.Sleep(5 * time.Second)

	ackMsg := connectiontypes.NewMsgConnectionOpenAck(
		connectionID,
		counterpartyConnectionID,
		counterpartyClient,
		tryProof,
		clientProof,
		consensusProof,
		clienttypes.ZeroHeight(),
		consensusHeight,
		connectiontypes.GetCompatibleVersions()[0],
		clientCtx.From,
	)

	_, err := r.sendTx(clientCtx, txf, ackMsg)
	if err != nil {
		return err
	}

	return nil
}
