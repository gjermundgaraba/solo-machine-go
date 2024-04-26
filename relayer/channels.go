package relayer

import (
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"go.uber.org/zap"
)

func (r *Relayer) QueryChannel(chainName string, portID string, channelID string) (*channeltypes.Channel, error) {
	clientCtx := r.createClientCtx(chainName)
	queryClient := channeltypes.NewQueryClient(clientCtx)
	req := &channeltypes.QueryChannelRequest{
		PortId:    portID,
		ChannelId: channelID,
	}

	res, err := queryClient.Channel(clientCtx.CmdContext, req)
	if err != nil {
		return nil, err
	}

	return res.Channel, nil
}

func (r *Relayer) InitChannel(chainName string, connectionID string, portID string, version string, counterpartyPortID string) (string, error) {
	clientCtx := r.createClientCtx(chainName)
	txf := r.createTxFactory(clientCtx, chainName)

	initMsg := channeltypes.NewMsgChannelOpenInit(
		portID,
		version,
		channeltypes.UNORDERED,
		[]string{connectionID},
		counterpartyPortID,
		clientCtx.From,
	)
	txResp, err := r.sendTx(clientCtx, txf, initMsg)
	if err != nil {
		return "", err
	}

	channelID, err := parseChannelIDFromEvents(txResp.Events)
	if err != nil {
		return "", err
	}

	r.logger.Info("Channel initialized on the cosmos chain", zap.String("channel-id", channelID))

	return channelID, nil
}

func (r *Relayer) ChannelOpenAck(chainName string, channelID string, counterpartyChannelID string, tryProof []byte, proofHeight clienttypes.Height) error {
	clientCtx := r.createClientCtx(chainName)
	txf := r.createTxFactory(clientCtx, chainName)

	ackMsg := channeltypes.NewMsgChannelOpenAck(
		transfertypes.PortID,
		channelID,
		counterpartyChannelID,
		transfertypes.Version,
		tryProof,
		proofHeight,
		clientCtx.From,
	)

	_, err := r.sendTx(clientCtx, txf, ackMsg)
	return err
}
