package solomachine

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	solomachineclient "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	"go.uber.org/zap"
	"time"
)

func (sm *SoloMachine) CreateCounterpartyLightClient(chainName string) error {
	chainStorage := sm.storage.GetChainStorage(chainName)

	publicKey, err := codectypes.NewAnyWithValue(sm.storage.GetPublicKey())
	if err != nil {
		return err
	}

	consensusState := &solomachineclient.ConsensusState{
		PublicKey:   publicKey,
		Diversifier: chainStorage.Diversifier(),
		Timestamp:   uint64(time.Now().UnixMilli()),
	}

	clientState := solomachineclient.NewClientState(1, consensusState)

	clientID, err := sm.r.CreateClient(chainName, clientState, consensusState)
	if err != nil {
		return err
	}

	chainStorage.SetCounterPartyClientID(clientID)

	return nil
}

func (sm *SoloMachine) CounterpartyLightClientExists(chainName string) bool {
	chainStorage := sm.storage.GetChainStorage(chainName)
	return chainStorage.CounterpartyClientID() != ""
}

func (sm *SoloMachine) UpdateCounterpartyLightClient(chainName string) error {
	chainStorage := sm.storage.GetChainStorage(chainName)
	clientID := chainStorage.CounterpartyClientID()

	soloMachineHeader, err := sm.createSoloMachineHeader(chainName)
	if err != nil {
		return err
	}
	if err := sm.r.UpdateClient(chainName, clientID, soloMachineHeader); err != nil {
		return err
	}

	sm.logger.Debug("Updated counterparty light client", zap.String("chainName", chainName))

	return nil
}

func (sm *SoloMachine) createSoloMachineHeader(chainName string) (*solomachineclient.Header, error) {
	chainStorage := sm.storage.GetChainStorage(chainName)
	clientState, err := sm.r.GetClientState(chainName, chainStorage.CounterpartyClientID())
	if err != nil {
		return nil, err
	}

	diversifier := chainStorage.Diversifier()
	publicKey, err := codectypes.NewAnyWithValue(sm.storage.GetPublicKey())
	if err != nil {
		return nil, err
	}

	data := &solomachineclient.HeaderData{
		NewPubKey:      publicKey,
		NewDiversifier: diversifier,
	}

	dataBz, err := sm.cdc.Marshal(data)
	if err != nil {
		return nil, err
	}

	timestamp := uint64(time.Now().UnixMilli())

	signBytes := &solomachineclient.SignBytes{
		Sequence:    clientState.Sequence,
		Timestamp:   timestamp,
		Diversifier: diversifier,
		Path:        []byte(solomachineclient.SentinelHeaderPath),
		Data:        dataBz,
	}

	bz, err := sm.cdc.Marshal(signBytes)
	if err != nil {
		return nil, err
	}

	sig, err := sm.GenerateSignature(bz)
	if err != nil {
		return nil, err
	}

	header := &solomachineclient.Header{
		Timestamp:      timestamp,
		Signature:      sig,
		NewPublicKey:   publicKey,
		NewDiversifier: diversifier,
	}

	return header, nil
}
