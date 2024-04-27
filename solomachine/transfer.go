package solomachine

import (
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"strconv"
	"time"
)

func (sm *SoloMachine) Transfer(chainName string, sender string, receiver string, denom string, amount uint64) error {
	chainStorage := sm.storage.GetChainStorage(chainName)

	if err := sm.UpdateCounterpartyLightClient(chainName); err != nil {
		return err
	}
	if err := sm.UpdateLightClient(chainName); err != nil {
		return err
	}

	// Give some time for stuff to update
	time.Sleep(5 * time.Second)

	ics20ChannelID := chainStorage.ICS20ChannelID()
	counterpartyICS20ChannelID := chainStorage.CounterpartyICS20Channel()

	lightClientState, err := chainStorage.LightClientState()
	if err != nil {
		return err
	}
	timeoutHeight := clienttypes.NewHeight(lightClientState.LatestHeight.RevisionNumber, lightClientState.LatestHeight.RevisionHeight+1000)

	amountStr := strconv.FormatInt(int64(amount), 10)
	fungibleTokenPacket := transfertypes.NewFungibleTokenPacketData(
		denom,
		amountStr,
		sender,
		receiver,
		"",
	)

	counterpartyLightClientState, err := sm.r.GetClientState(chainName, chainStorage.CounterpartyClientID())
	if err != nil {
		return err
	}
	sequence := counterpartyLightClientState.Sequence

	packet := channeltypes.NewPacket(
		fungibleTokenPacket.GetBytes(),
		sequence,
		transfertypes.PortID,
		ics20ChannelID,
		transfertypes.PortID,
		counterpartyICS20ChannelID,
		timeoutHeight,
		0,
	)

	commitmentProof, err := sm.GenerateCommitmentProof(chainName, packet, sequence)
	if err != nil {
		return err
	}

	return sm.r.SendMsgRecvPacket(chainName, packet, commitmentProof, lightClientState.LatestHeight)
}
