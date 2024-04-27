package relayer

import (
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

func (r *Relayer) SendMsgRecvPacket(
	chainName string,
	packet channeltypes.Packet,
	commitmentProof []byte,
	proofHeight clienttypes.Height,
) error {
	clientCtx := r.createClientCtx(chainName)
	txf := r.createTxFactory(clientCtx, chainName)

	msg := channeltypes.NewMsgRecvPacket(
		packet,
		commitmentProof,
		proofHeight,
		clientCtx.From,
	)

	_, err := r.sendTx(clientCtx, txf, msg)
	return err
}
