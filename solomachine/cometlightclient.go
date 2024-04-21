package solomachine

import (
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"go.uber.org/zap"
)

func (sm *SoloMachine) CreateCometLightClient(clientState exported.ClientState, consensusState exported.ConsensusState) error {
	// TODO: Persist this somehow
	// Might want to look into how the tendermint light client is made from MsgCreateClient and mimic some of that

	sm.logger.Debug("Creating Comet Light Client in the Solo Machine",
		zap.String("client state client type", clientState.ClientType()),
		zap.Uint64("client state height revision number", clientState.GetLatestHeight().GetRevisionNumber()),
		zap.Uint64("client state height revision height", clientState.GetLatestHeight().GetRevisionHeight()),
		zap.String("consensus state client type", consensusState.ClientType()),
		zap.Uint64("consensus timestampe", consensusState.GetTimestamp()),
	)

	return nil
}
