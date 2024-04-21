package solomachine

import (
	"encoding/hex"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	solomachineclient "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	"go.uber.org/zap"
)

type SoloMachine struct {
	logger *zap.Logger

	Sequence    uint64
	PublicKey   cryptotypes.PubKey
	Diversifier string
	Time        uint64
}

func NewSoloMachine(logger *zap.Logger) *SoloMachine {
	// TODO: Get values from some peristence layer
	seq := uint64(1)

	privKeyBytes, err := hex.DecodeString("")
	if err != nil {
		panic(err)
	}
	privKey := &secp256k1.PrivKey{
		Key: privKeyBytes,
	}
	pk := privKey.PubKey()

	div := "diverse"

	time := uint64(1)

	return &SoloMachine{
		logger:      logger,
		Sequence:    seq,
		PublicKey:   pk,
		Diversifier: div,
		Time:        time,
	}
}

// ClientState returns a new solo machine ClientState instance.
func (sm *SoloMachine) ClientState() (*solomachineclient.ClientState, error) {
	consensusState, err := sm.ConsensusState()
	if err != nil {
		return nil, err
	}
	return solomachineclient.NewClientState(sm.Sequence, consensusState), nil
}

// ConsensusState returns a new solo machine ConsensusState instance
func (sm *SoloMachine) ConsensusState() (*solomachineclient.ConsensusState, error) {
	publicKey, err := codectypes.NewAnyWithValue(sm.PublicKey)
	if err != nil {
		return nil, err
	}

	return &solomachineclient.ConsensusState{
		PublicKey:   publicKey,
		Diversifier: sm.Diversifier,
		Timestamp:   sm.Time,
	}, nil
}
