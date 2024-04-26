package solomachine

import (
	"fmt"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	solomachineclient "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	tmclient "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"go.uber.org/zap"
	"time"
)

func (sm *SoloMachine) ConnectionExists(chainName string) bool {
	chainStorage := sm.storage.GetChainStorage(chainName)
	return chainStorage.ConnectionExists()
}

func (sm *SoloMachine) CreateConnection(chainName string) error {
	chainStorage := sm.storage.GetChainStorage(chainName)

	clientID := chainStorage.ClientID()
	counterpartyClientID := chainStorage.CounterpartyClientID()
	connectionID := chainStorage.ConnectionID()
	counterpartyConnectionID := chainStorage.CounterpartyConnectionID()

	if !chainStorage.CounterpartyConnectionExists() {
		if err := sm.UpdateCounterpartyLightClient(chainName); err != nil {
			return err
		}

		var err error
		counterpartyConnectionID, err = sm.r.InitConnection(chainName, counterpartyClientID, clientID)
		if err != nil {
			return err
		}
		chainStorage.SetCounterpartyConnectionID(counterpartyConnectionID)
		sm.logger.Info("Connection initialized on the chain", zap.String("chain", chainName), zap.String("connection-id", counterpartyConnectionID))
	}

	if !chainStorage.ConnectionExists() {
		if err := sm.UpdateLightClient(chainName); err != nil {
			return err
		}

		// Similar to OPEN_TRY, sort-of, except we don't keep that as a state
		// TODO: Keep OPEN_TRY as a state?
		connectionID = chainStorage.CreateConnection()
		sm.logger.Info("Connection on solo machine", zap.String("for chain chain", chainName), zap.String("connection-id", connectionID))
	}

	counterpartyConnectionEnd, err := sm.r.QueryConnection(chainName, counterpartyConnectionID)
	if err != nil {
		return err
	}
	if counterpartyConnectionEnd.State == connectiontypes.OPEN {
		return nil // All good, connection is already open
	}
	if counterpartyConnectionEnd.State != connectiontypes.INIT {
		return fmt.Errorf("unexpected connection state: wanted %s, got %s", connectiontypes.INIT, counterpartyConnectionEnd.State)
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

	tryProof, err := sm.GenerateConnOpenTryProof(chainName, sequence)
	if err != nil {
		return err
	}

	sequence++ // Because. That is how it works.

	clientProof, err := sm.GenerateClientStateProof(chainName, sequence, lightClientState)
	if err != nil {
		return err
	}

	sequence++ // Because. That is how it works.

	consensusProof, err := sm.GenerateConsensusStateProof(chainName, sequence, lightClientState)
	if err != nil {
		return err
	}

	if err := sm.r.ConnectionOpenAck(
		chainName,
		counterpartyConnectionID,
		connectionID,
		lightClientState,
		tryProof,
		clientProof,
		consensusProof,
		lightClientState.LatestHeight,
	); err != nil {
		return err
	}

	if err := sm.UpdateLightClient(chainName); err != nil {
		return err
	}
	// This is where we would normally call ConfirmOpenConnection, but it doesn't really make any sense for the solo machine

	return nil
}

// GenerateConnOpenTryProof generates the proofTry required for the connection open ack handshake step.
// The clientID, connectionID provided represent the clientID and connectionID created on the counterparty chain, that is the tendermint chain.
func (sm *SoloMachine) GenerateConnOpenTryProof(chainName string, sequence uint64) ([]byte, error) {
	chainStorage := sm.storage.GetChainStorage(chainName)

	merklePrefix := commitmenttypes.NewMerklePrefix([]byte(exported.StoreKey))

	counterparty := connectiontypes.NewCounterparty(chainStorage.CounterpartyClientID(), chainStorage.CounterpartyConnectionID(), merklePrefix)
	connection := connectiontypes.NewConnectionEnd(connectiontypes.TRYOPEN, chainStorage.ClientID(), counterparty, []*connectiontypes.Version{connectiontypes.GetCompatibleVersions()[0]}, 0)

	data, err := sm.cdc.Marshal(&connection)
	if err != nil {
		return nil, err
	}

	merklePath := commitmenttypes.NewMerklePath(host.ConnectionPath(chainStorage.ConnectionID()))
	merklePath, err = commitmenttypes.ApplyPrefix(merklePrefix, merklePath)
	if err != nil {
		return nil, err
	}
	// in a multistore context: index 0 is the key for the IBC store in the multistore, index 1 is the key in the IBC store
	key, err := merklePath.GetKey(1)
	if err != nil {
		return nil, err
	}
	signBytes := &solomachineclient.SignBytes{
		Sequence:    sequence,
		Timestamp:   uint64(time.Now().UnixMilli()),
		Diversifier: chainStorage.Diversifier(),
		Path:        key,
		Data:        data,
	}

	sm.logger.Debug("generated sign bytes", zap.Uint64("sequence", sequence), zap.Uint64("timestamp", signBytes.Timestamp), zap.String("diversifier", signBytes.Diversifier), zap.String("path", string(signBytes.Path)), zap.String("data", string(signBytes.Data)))

	return sm.GenerateProof(signBytes)
}

func (sm *SoloMachine) GenerateClientStateProof(chainName string, sequence uint64, clientState exported.ClientState) ([]byte, error) {
	chainStorage := sm.storage.GetChainStorage(chainName)

	data, err := ibcclienttypes.MarshalClientState(sm.cdc, clientState)
	if err != nil {
		return nil, err
	}

	path := host.FullClientStateKey(chainStorage.ClientID())
	signBytes := &solomachineclient.SignBytes{
		Sequence:    sequence,
		Timestamp:   uint64(time.Now().UnixMilli()),
		Diversifier: chainStorage.Diversifier(),
		Path:        path,
		Data:        data,
	}

	return sm.GenerateProof(signBytes)
}

func (sm *SoloMachine) GenerateConsensusStateProof(chainName string, sequence uint64, clientState *tmclient.ClientState) ([]byte, error) {
	chainStorage := sm.storage.GetChainStorage(chainName)

	height := clientState.LatestHeight
	consensusState, err := chainStorage.GetLightConsensusState(height)
	if err != nil {
		return nil, err
	}

	data, err := ibcclienttypes.MarshalConsensusState(sm.cdc, consensusState)
	if err != nil {
		return nil, err
	}

	path := host.FullConsensusStateKey(chainStorage.ClientID(), height)
	signBytes := &solomachineclient.SignBytes{
		Sequence:    sequence,
		Timestamp:   uint64(time.Now().UnixMilli()),
		Diversifier: chainStorage.Diversifier(),
		Path:        path,
		Data:        data,
	}

	return sm.GenerateProof(signBytes)
}
