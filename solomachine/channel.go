package solomachine

import (
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	solomachineclient "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	"go.uber.org/zap"
	"time"
)

func (sm *SoloMachine) ICS20ChannelExists(chainName string) bool {
	chainStorage := sm.storage.GetChainStorage(chainName)
	return chainStorage.ICS20ChannelExists()
}

func (sm *SoloMachine) CreateICS20Channel(chainName string) error {
	chainStorage := sm.storage.GetChainStorage(chainName)

	counterpartyClientID := chainStorage.CounterpartyClientID()
	connectionID := chainStorage.ConnectionID()
	counterpartyConnectionID := chainStorage.CounterpartyConnectionID()
	ics20ChannelID := chainStorage.ICS20ChannelID()
	counterpartyICS20ChannelID := chainStorage.CounterpartyICS20Channel()

	if !chainStorage.CounterpartyICS20ChannelExists() {
		if err := sm.UpdateCounterpartyLightClient(chainName); err != nil {
			return err
		}

		var err error
		counterpartyICS20ChannelID, err = sm.r.InitChannel(
			chainName,
			counterpartyConnectionID,
			transfertypes.PortID,
			transfertypes.Version,
			transfertypes.PortID,
		)
		if err != nil {
			return err
		}
		chainStorage.SetCounterpartyICS20ChannelID(counterpartyICS20ChannelID)
		sm.logger.Info("ICS20 channel initialized on the chain", zap.String("chain", chainName), zap.String("channel-id", counterpartyICS20ChannelID))
	}

	if !chainStorage.ICS20ChannelExists() {
		if err := sm.UpdateLightClient(chainName); err != nil {
			return err
		}

		// Similar to OPEN_TRY, sort-of, except we don't keep that as a state
		ics20ChannelID = chainStorage.CreateICS20Channel()
		sm.logger.Info("Created ICS20 channel on solo machine", zap.String("for chain", chainName), zap.String("channel-id", ics20ChannelID))
	}

	counterpartyChannel, err := sm.r.QueryChannel(chainName, transfertypes.PortID, counterpartyICS20ChannelID)
	if err != nil {
		return err
	}
	if counterpartyChannel.State == channeltypes.OPEN {
		return nil // All good, channel is already open
	}

	counterpartyClientState, err := sm.r.GetClientState(chainName, counterpartyClientID)
	if err != nil {
		return err
	}
	sequence := counterpartyClientState.Sequence

	lightClientState, err := chainStorage.LightClientState()
	if err != nil {
		return err
	}

	tryProof, err := sm.generateChanOpenTryProof(
		chainName,
		sequence,
		connectionID,
		transfertypes.PortID,
		ics20ChannelID,
		transfertypes.Version,
		transfertypes.PortID,
		counterpartyICS20ChannelID,
	)

	if err := sm.r.ChannelOpenAck(
		chainName,
		counterpartyICS20ChannelID,
		ics20ChannelID,
		tryProof,
		lightClientState.LatestHeight); err != nil {
		return err
	}

	return nil
}

// generateChanOpenTryProof generates the proofTry required for the channel open ack handshake step.
func (sm *SoloMachine) generateChanOpenTryProof(
	chainName string,
	sequence uint64,
	connectionID string,
	portID string,
	channelID string,
	version string,
	counterpartyPortID string,
	counterpartyChannelID string,
) ([]byte, error) {
	chainStorage := sm.storage.GetChainStorage(chainName)

	counterparty := channeltypes.NewCounterparty(counterpartyPortID, counterpartyChannelID)
	channel := channeltypes.NewChannel(channeltypes.TRYOPEN, channeltypes.UNORDERED, counterparty, []string{connectionID}, version)

	data, err := sm.cdc.Marshal(&channel)
	if err != nil {
		return nil, err
	}

	path := host.ChannelKey(portID, channelID)
	signBytes := &solomachineclient.SignBytes{
		Sequence:    sequence,
		Timestamp:   uint64(time.Now().UnixMilli()),
		Diversifier: chainStorage.Diversifier(),
		Path:        path,
		Data:        data,
	}

	return sm.GenerateProof(signBytes)
}
