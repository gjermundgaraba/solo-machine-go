package relayer

import (
	"fmt"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	"go.uber.org/zap"
)

func (r *Relayer) InitConnection() error {
	if err := r.UpdateClients(); err != nil {
		return err
	}

	tendermintClientID := r.soloMachine.TendermintClientID()

	connectionID, err := r.cosmosChain.InitConnection(
		r.config.CosmosChain.SoloMachineLightClient.IBCClientID,
		tendermintClientID,
	)
	if err != nil {
		return err
	}

	r.logger.Info("Connection initialized on the cosmos chain", zap.String("connection-id", connectionID))

	r.config.CosmosChain.SoloMachineLightClient.ConnectionID = connectionID
	if err = WriteConfigToFile(r.config, "", true); err != nil {
		return err
	}
	r.logger.Info("Config updated with connection ID created on the cosmos chain", zap.String("connection-id", r.config.CosmosChain.SoloMachineLightClient.ConnectionID))

	return nil
}

func (r *Relayer) FinishAnyRemainingConnectionHandshakes() error {
	if err := r.UpdateClients(); err != nil {
		return err
	}

	connectionID := r.config.CosmosChain.SoloMachineLightClient.ConnectionID
	soloMachineConnectionID := r.soloMachine.ConnectionID()

	connection, err := r.cosmosChain.QueryConnection(connectionID)
	if err != nil {
		return err
	}
	if connection.State == connectiontypes.OPEN {
		return nil // All good, connection is already open
	}
	if connection.State != connectiontypes.INIT {
		return fmt.Errorf("unexpected connection state: wanted %s, got %s", connectiontypes.INIT, connection.State)
	}

	clientState, err := r.cosmosChain.GetClientState(r.config.CosmosChain.SoloMachineLightClient.IBCClientID)
	if err != nil {
		return err
	}

	sequence := clientState.Sequence
	tryProof, err := r.soloMachine.GenerateConnOpenTryProof(sequence, r.config.CosmosChain.SoloMachineLightClient.IBCClientID, connectionID)
	if err != nil {
		return err
	}
	sequence++

	tendermintLightClientState, err := r.soloMachine.LightClientState()
	if err != nil {
		return err
	}

	clientProof, err := r.soloMachine.GenerateClientStateProof(sequence, tendermintLightClientState)
	if err != nil {
		return err
	}
	sequence++

	consensusProof, err := r.soloMachine.GenerateConsensusStateProof(sequence, tendermintLightClientState)
	if err != nil {
		return err
	}

	if err := r.cosmosChain.AckOpenConnection(
		connectionID,
		soloMachineConnectionID,
		tendermintLightClientState,
		tryProof,
		clientProof,
		consensusProof,
		tendermintLightClientState.LatestHeight,
	); err != nil {
		return err
	}

	r.logger.Info("Connection handshake completed", zap.String("connection-id", connectionID))

	return nil
}
