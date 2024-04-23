package relayer

import (
	"fmt"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	"go.uber.org/zap"
)

func (r *Relayer) InitConnection() error {
	tendermintClientID := r.soloMachine.TendermintClientID()

	connectionID, err := r.cosmosChain.InitConnection(
		r.config.CosmosChain.SoloMachineLightClient.IBCClientID,
		tendermintClientID,
	)
	if err != nil {
		return err
	}

	r.config.CosmosChain.SoloMachineLightClient.ConnectionID = connectionID
	if err = WriteConfigToFile(r.config, "", true); err != nil {
		return err
	}
	r.logger.Info("Config updated with connection ID created on the cosmos chain", zap.String("connection-id", r.config.CosmosChain.SoloMachineLightClient.ConnectionID))

	return nil
}

func (r *Relayer) FinishAnyRemainingConnectionHandshakes() error {
	connectionID := r.config.CosmosChain.SoloMachineLightClient.ConnectionID
	tendermintClientID := r.soloMachine.TendermintClientID()

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

	tendermintLightClientState, err := r.soloMachine.LightClientState()
	if err != nil {
		return err
	}

	tryProof, err := r.soloMachine.GenerateConnOpenTryProof(r.config.CosmosChain.SoloMachineLightClient.IBCClientID, connectionID)
	if err != nil {
		return err
	}

	clientProof, err := r.soloMachine.GenerateClientStateProof(tendermintLightClientState)
	if err != nil {
		return err
	}

	consensusProof, err := r.soloMachine.GenerateConsensusStateProof(tendermintLightClientState)
	if err != nil {
		return err
	}

	if err := r.cosmosChain.AckOpenConnection(
		connectionID,
		tendermintClientID,
		tendermintLightClientState,
		tryProof,
		clientProof,
		consensusProof,
		tendermintLightClientState.LatestHeight,
	); err != nil {
		return err
	}

	return nil
}
