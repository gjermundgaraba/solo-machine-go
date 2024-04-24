package relayer

import (
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"go.uber.org/zap"
)

func (r *Relayer) CreateSoloMachineLightClientOnCosmos() error {
	clientState, err := r.soloMachine.ClientState(1)
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
	clientState, ibcHeader, err := r.cosmosChain.GetClientInfo()
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

func (r *Relayer) UpdateClients() error {
	r.logger.Info("Updating clients...")
	_, ibcHeader, err := r.cosmosChain.GetClientInfo()
	if err != nil {
		return err
	}
	if err := r.soloMachine.UpdateClient(ibcHeader); err != nil {
		return err
	}

	clientState, err := r.cosmosChain.GetClientState(r.config.CosmosChain.SoloMachineLightClient.IBCClientID)
	if err != nil {
		return err
	}

	soloMachineHeader, err := r.soloMachine.CreateHeader(clientState.Sequence)
	if err != nil {
		return err
	}
	if err := r.cosmosChain.UpdateClient(r.config.CosmosChain.SoloMachineLightClient.IBCClientID, soloMachineHeader); err != nil {
		return err
	}

	r.logger.Info("Clients updated successfully")
	return nil
}
