package solomachine

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	tmclient "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/gjermundgaraba/solo-machine-go/utils"
	"go.uber.org/zap"
)

const clientID = "07-tendermint-1"

func (sm *SoloMachine) TendermintClientID() string {
	return clientID
}

func (sm *SoloMachine) LightClientState() (*tmclient.ClientState, error) {
	clientStore := sm.ClientStore(sdk.Context{}, clientID)
	bz := clientStore.Get(host.ClientStateKey())
	if len(bz) == 0 {
		return nil, fmt.Errorf("light client state not found")
	}

	clientStateI := clienttypes.MustUnmarshalClientState(sm.cdc, bz)
	var clientState *tmclient.ClientState
	clientState, ok := clientStateI.(*tmclient.ClientState)
	if !ok {
		return nil, fmt.Errorf("cannot convert %T into %T", clientStateI, clientState)
	}

	return clientState, nil
}

// GetConsensusState retrieves the consensus state from the client prefixed store.
// If the ConsensusState does not exist in state for the provided height a nil value and false boolean flag is returned
func (sm *SoloMachine) GetLightConsensusState(height ibcexported.Height) (*tmclient.ConsensusState, error) {
	clientStore := sm.ClientStore(sdk.Context{}, clientID)
	bz := clientStore.Get(host.ConsensusStateKey(height))
	if len(bz) == 0 {
		return nil, fmt.Errorf("consensus state not found for height: %s", height)
	}

	consensusStateI := clienttypes.MustUnmarshalConsensusState(sm.cdc, bz)
	var consensusState *tmclient.ConsensusState
	consensusState, ok := consensusStateI.(*tmclient.ConsensusState)
	if !ok {
		return nil, fmt.Errorf("cannot convert %T into %T", consensusStateI, consensusState)
	}

	return consensusState, nil
}

func (sm *SoloMachine) LightClientExists() (bool, error) {
	latestHeight := sm.tmLightClient.LatestHeight(sdk.Context{}, clientID)
	return latestHeight.GetRevisionHeight() != 0, nil
}

func (sm *SoloMachine) CreateTendermintLightClient(msgCreateClient *clienttypes.MsgCreateClient, ibcHeader utils.TendermintIBCHeader) error {
	// TODO: Should we maybe not allow this if the client is already created?
	header := *ibcHeader.SignedHeader.Header.ToProto()
	ctx := sdk.NewContext(sm.store, header, false, sm.sdkLogger)

	sm.logger.Debug("Creating tendermint light client", zap.Uint64("height", ibcHeader.Height()))

	if err := sm.tmLightClient.Initialize(ctx, clientID, msgCreateClient.ClientState.Value, msgCreateClient.ConsensusState.Value); err != nil {
		return err
	}

	sm.logger.Debug("Initialized tendermint light client", zap.Any("client-id", clientID))

	sm.store.Commit()

	sm.logger.Debug("store version", zap.Int64("version", sm.store.LatestVersion()))
	return nil
}
