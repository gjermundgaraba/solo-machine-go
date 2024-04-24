package relayer

import (
	"fmt"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"go.uber.org/zap"
)

func (r *Relayer) InitChannel() error {
	if err := r.UpdateClients(); err != nil {
		return err
	}

	connectionID := r.config.CosmosChain.SoloMachineLightClient.ConnectionID

	channelID, err := r.cosmosChain.InitICS20Channel(connectionID)
	if err != nil {
		return err
	}

	r.logger.Info("Channel initialized on the cosmos chain", zap.String("channel-id", channelID))

	r.config.CosmosChain.SoloMachineLightClient.ChannelID = channelID
	if err = WriteConfigToFile(r.config, "", true); err != nil {
		return err
	}
	r.logger.Info("Config updated with channel ID created on the cosmos chain", zap.String("channel-id", r.config.CosmosChain.SoloMachineLightClient.ChannelID))

	return nil
}

func (r *Relayer) FinishAnyRemainingChannelHandshakes() error {
	r.logger.Debug("Starting channel handshake")
	if err := r.UpdateClients(); err != nil {
		return err
	}

	//connectionID := r.config.CosmosChain.SoloMachineLightClient.ConnectionID
	channelID := r.config.CosmosChain.SoloMachineLightClient.ChannelID

	channel, err := r.cosmosChain.QueryChannel(transfertypes.PortID, channelID)
	if err != nil {
		return err
	}
	if channel.State == channeltypes.OPEN {
		r.logger.Debug("Channel already open", zap.String("channel-id", channelID))
		return nil // All good, channel is already open
	}
	if channel.State != channeltypes.INIT {
		return fmt.Errorf("unexpected channel state: wanted %s, got %s", channeltypes.INIT, channel.State)
	}

	r.logger.Debug("Channel is in INIT state, starting handshake", zap.String("channel-id", channelID))

	clientState, err := r.cosmosChain.GetClientState(r.config.CosmosChain.SoloMachineLightClient.IBCClientID)
	if err != nil {
		return err
	}

	tryProof, err := r.soloMachine.GenerateChanOpenTryProof(
		clientState.Sequence,
		transfertypes.PortID,
		transfertypes.Version,
		channelID)
	if err != nil {
		return err
	}

	tendermintLightClientState, err := r.soloMachine.LightClientState()
	if err != nil {
		return err
	}

	soloMachineChannelID := r.soloMachine.ChannelID()

	if err := r.cosmosChain.AckOpenICS20Channel(channelID, soloMachineChannelID, tryProof, tendermintLightClientState.LatestHeight); err != nil {
		return err
	}

	r.logger.Info("Channel handshake completed", zap.String("channel-id", channelID))

	return nil
}
