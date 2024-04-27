package solomachine

import (
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

type Status struct {
	Diversifier string

	LightClientID                   string
	LightClientLatestHeight         clienttypes.Height
	CounterpartyActualHeight        uint64
	CounterpartyLightClientID       string
	CounterpartyLightClientSequence uint64
	ConnectionID                    string
	CounterpartyConnectionID        string
	CounterpartyConnectionState     string
	ICS20ChannelID                  string
	CounterpartyICS20ChannelID      string
	CounterpartyICS20ChannelState   string
}

func (sm *SoloMachine) Status(chainName string) (Status, error) {
	chainStorage := sm.storage.GetChainStorage(chainName)

	var lightClientLatestHeight clienttypes.Height
	if chainStorage.LightClientExists() {
		lightClientState, err := chainStorage.LightClientState()
		if err != nil {
			return Status{}, err
		}

		lightClientLatestHeight = lightClientState.LatestHeight
	}

	var counterpartySequence uint64
	var counterpartyActualHeight uint64
	if sm.CounterpartyLightClientExists(chainName) {
		counterpartyLightClientState, err := sm.r.GetClientState(chainName, chainStorage.CounterpartyClientID())
		if err != nil {
			return Status{}, err
		}
		counterpartySequence = counterpartyLightClientState.Sequence

		latestIBCHeader, err := sm.r.GetLatestIBCHeader(chainName)
		if err != nil {
			return Status{}, err
		}

		counterpartyActualHeight = uint64(latestIBCHeader.Header.Height)
	}

	var counterpartyConnectionState connectiontypes.State
	if chainStorage.CounterpartyConnectionExists() {
		connectionEnd, err := sm.r.QueryConnection(chainName, chainStorage.ConnectionID())
		if err != nil {
			return Status{}, err
		}

		counterpartyConnectionState = connectionEnd.State
	}

	var counterpartyICS20ChannelState channeltypes.State
	if chainStorage.CounterpartyICS20ChannelExists() {
		channelEnd, err := sm.r.QueryChannel(chainName, transfertypes.PortID, chainStorage.CounterpartyICS20Channel())
		if err != nil {
			return Status{}, err
		}

		counterpartyICS20ChannelState = channelEnd.State
	}

	return Status{
		Diversifier:                     chainStorage.Diversifier(),
		LightClientID:                   chainStorage.ClientID(),
		LightClientLatestHeight:         lightClientLatestHeight,
		CounterpartyActualHeight:        counterpartyActualHeight,
		CounterpartyLightClientID:       chainStorage.CounterpartyClientID(),
		CounterpartyLightClientSequence: counterpartySequence,
		ConnectionID:                    chainStorage.ConnectionID(),
		CounterpartyConnectionID:        chainStorage.CounterpartyConnectionID(),
		CounterpartyConnectionState:     counterpartyConnectionState.String(),
		ICS20ChannelID:                  chainStorage.ICS20ChannelID(),
		CounterpartyICS20ChannelID:      chainStorage.CounterpartyICS20Channel(),
		CounterpartyICS20ChannelState:   counterpartyICS20ChannelState.String(),
	}, nil
}
