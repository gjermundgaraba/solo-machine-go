package relayer

import (
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
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

	r.config.CosmosChain.SoloMachineLightClient.IBCClientID = clientID
	err = WriteConfigToFile(r.config, "", true)
	if err != nil {
		return err
	}

	r.logger.Info("Config updated with light client ID created on the cosmos chain", zap.String("client-id", r.config.CosmosChain.SoloMachineLightClient.IBCClientID))

	return nil
}

func (r *Relayer) SoloMachineLightClientExistsOnCosmos() (bool, error) {
	clientID := r.config.CosmosChain.SoloMachineLightClient.IBCClientID
	return r.cosmosChain.ClientExists(clientID)
}

func (r *Relayer) CreateTendermintLightClientOnSoloMachine() error {
	clientState, ibcHeader, err := r.cosmosChain.GetCreateClientInfo()
	if err != nil {
		return err
	}

	msg, err := clienttypes.NewMsgCreateClient(clientState, ibcHeader.ConsensusState(), "")
	if err != nil {
		return err
	}

	return r.soloMachine.CreateTendermintLightClient(msg, ibcHeader)
}

func (r *Relayer) TendermintLightClientExistsOnSoloMachine() (bool, error) {
	return r.soloMachine.LightClientExists()
}
