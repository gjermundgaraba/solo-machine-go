package storage

import connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"

func (cs *ChainStorage) ConnectionID() string {
	return cs.connectionID
}

func (cs *ChainStorage) CounterpartyConnectionID() string {
	return cs.counterpartyConnectionID
}

func (cs *ChainStorage) CounterpartyConnectionExists() bool {
	return cs.counterpartyConnectionID != ""
}

func (cs *ChainStorage) ConnectionExists() bool {
	return cs.connectionID != ""
}

func (cs *ChainStorage) SetCounterpartyConnectionID(connectionID string) {
	cs.store.Set([]byte(counterpartyConnectionIDKey), []byte(connectionID))
	cs.counterpartyConnectionID = connectionID
	cs.parent.Commit()
}

// Similar to OPEN_TRY essentially
func (cs *ChainStorage) CreateConnection() string {
	nextConnectionSeq := cs.parent.nextConnectionNumber
	defer cs.parent.incrementNextConnectionNumber()

	connectionID := connectiontypes.FormatConnectionIdentifier(nextConnectionSeq)
	cs.setConnectionID(connectionID)

	return connectionID
}

func (cs *ChainStorage) setConnectionID(connectionID string) {
	cs.store.Set([]byte(connectionIDKey), []byte(connectionID))
	cs.connectionID = connectionID
	cs.parent.Commit()
}
