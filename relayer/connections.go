package relayer

import "go.uber.org/zap"

func (r *Relayer) CreateConnections() error {
	tendermintClientID := r.soloMachine.TendermintClientID()

	// TODO: Check if there is a connection already
	connectionID, err := r.cosmosChain.InitConnection(
		r.config.CosmosChain.SoloMachineLightClient.IBCClientID,
		tendermintClientID,
	)
	if err != nil {
		return err
	}

	// Caller needs to save the config to disk after this (I don't like this approach very much)
	r.config.CosmosChain.SoloMachineLightClient.ConnectionID = connectionID
	if err = WriteConfigToFile(r.config, "", true); err != nil {
		return err
	}
	r.logger.Info("Config updated with connection ID created on the cosmos chain", zap.String("connection-id", r.config.CosmosChain.SoloMachineLightClient.ConnectionID))

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
