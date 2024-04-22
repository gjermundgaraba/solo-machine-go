package cosmoschain

import (
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

// TODO: Make this function check if the connection exists, and if it is not completed, complete it
// Or make another function that does that part
func (cc *CosmosChain) InitConnection(
	clientID string,
	counterPartyClientID string,
) (string, error) {
	var version *connectiontypes.Version // Can be nil? Not sure.
	merklePrefix := commitmenttypes.NewMerklePrefix([]byte(ibcexported.StoreKey))
	initMsg := connectiontypes.NewMsgConnectionOpenInit(
		clientID,
		counterPartyClientID,
		merklePrefix,
		version,
		0,
		cc.clientCtx.From,
	)

	txResp, err := cc.sendTx(initMsg)
	if err != nil {
		return "", err
	}

	connectionID, err := parseConnectionIDFromEvents(txResp.Events)
	if err != nil {
		return "", err
	}

	return connectionID, nil
}

func (cc *CosmosChain) AckOpenConnection(
	connectionID string,
	counterpartyClientID string,
	counterpartyClient ibcexported.ClientState,
	tryProof, clientProof, consensusProof []byte,
	consensusHeight clienttypes.Height,
) error {
	ackMsg := connectiontypes.NewMsgConnectionOpenAck(
		connectionID,
		counterpartyClientID,
		counterpartyClient,
		tryProof,
		clientProof,
		consensusProof,
		clienttypes.ZeroHeight(),
		consensusHeight,
		connectiontypes.GetCompatibleVersions()[0],
		cc.clientCtx.From,
	)

	_, err := cc.sendTx(ackMsg)
	if err != nil {
		return err
	}

	return nil
}
