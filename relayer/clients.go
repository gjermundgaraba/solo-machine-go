package relayer

import (
	"go.uber.org/zap"
)

func (r *Relayer) CreateSoloMachineLightClientOnCosmos() error {
	clientState, err := r.soloMachine.ClientState()
	if err != nil {
		return err
	}

	consensusState, err := r.soloMachine.ConsensusState()
	if err != nil {
		return err
	}

	clientID, err := r.cosmosChain.CreateClient(clientState, consensusState)
	r.logger.Info("Client on CosmosChain created successfully", zap.String("client-id", clientID))

	// Caller should save config back to disk probably
	r.config.CosmosChain.SoloMachineLightClient.IBCClientID = clientID

	return nil
}

func (r *Relayer) CreateTendermintLightClientOnSoloMachine() error {
	clientState, consensusState, err := r.cosmosChain.GetCreateClientInfo()
	if err != nil {
		return err
	}

	return r.soloMachine.CreateCometLightClient(clientState, consensusState)
}
