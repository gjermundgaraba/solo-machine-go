package cosmoschain

import (
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

func (cc *CosmosChain) QueryChannel(
	portID, channelID string,
) (*channeltypes.Channel, error) {
	queryClient := channeltypes.NewQueryClient(cc.clientCtx)
	req := &channeltypes.QueryChannelRequest{
		PortId:    portID,
		ChannelId: channelID,
	}

	res, err := queryClient.Channel(cc.clientCtx.CmdContext, req)
	if err != nil {
		return nil, err
	}

	return res.Channel, nil
}

func (cc *CosmosChain) InitICS20Channel(connectionID string) (string, error) {
	initMsg := channeltypes.NewMsgChannelOpenInit(
		transfertypes.PortID,
		transfertypes.Version,
		channeltypes.UNORDERED,
		[]string{connectionID},
		transfertypes.PortID,
		cc.clientCtx.From,
	)

	txResp, err := cc.sendTx(initMsg)
	if err != nil {
		return "", err
	}

	channelID, err := parseChannelIDFromEvents(txResp.Events)
	if err != nil {
		return "", err
	}

	return channelID, nil
}

func (cc *CosmosChain) AckOpenICS20Channel(
	channelID string,
	counterpartyChannelID string,
	tryProof []byte,
	proofHeight clienttypes.Height,
) error {
	ackMsg := channeltypes.NewMsgChannelOpenAck(
		transfertypes.PortID,
		channelID,
		counterpartyChannelID,
		transfertypes.Version,
		tryProof,
		proofHeight,
		cc.clientCtx.From,
	)

	_, err := cc.sendTx(ackMsg)
	return err
}
