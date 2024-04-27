package storage

import channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

func (cs *ChainStorage) CreateICS20Channel() string {
	nextChannelSeq := cs.parent.nextChannelNumber
	defer cs.parent.incrementNextChannelNumber()

	channelID := channeltypes.FormatChannelIdentifier(nextChannelSeq)
	cs.setICS20Channel(channelID)

	return channelID
}

func (cs *ChainStorage) ICS20ChannelExists() bool {
	return cs.ics20Channel != ""
}

func (cs *ChainStorage) CounterpartyICS20ChannelExists() bool {
	return cs.counterpartyICS20Channel != ""
}

func (cs *ChainStorage) ICS20ChannelID() string {
	return cs.ics20Channel
}

func (cs *ChainStorage) CounterpartyICS20Channel() string {
	return cs.counterpartyICS20Channel
}

func (cs *ChainStorage) setICS20Channel(channelID string) {
	cs.store.Set([]byte(ics20ChannelKey), []byte(channelID))
	cs.ics20Channel = channelID
	cs.parent.Commit()
}

func (cs *ChainStorage) SetCounterpartyICS20ChannelID(channelID string) {
	cs.store.Set([]byte(counterpartyICS20ChannelKey), []byte(channelID))
	cs.counterpartyICS20Channel = channelID
	cs.parent.Commit()
}
