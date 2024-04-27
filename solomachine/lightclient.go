package solomachine

import (
	"github.com/cometbft/cometbft/light"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	tmclient "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"go.uber.org/zap"
	"time"
)

// DefaultUpgradePath is the default IBC upgrade path set for an on-chain light client
var defaultUpgradePath = []string{"upgrade", "upgradedIBCState"}

func (sm *SoloMachine) LightClientExists(chainName string) bool {
	chainStorage := sm.storage.GetChainStorage(chainName)
	return chainStorage.LightClientExists()
}

func (sm *SoloMachine) CreateLightClient(chainName string) error {
	ibcHeader, err := sm.r.GetLatestIBCHeader(chainName)
	if err != nil {
		return err
	}

	revisionNumber := clienttypes.ParseChainID(ibcHeader.Header.ChainID)
	unbondingPeriod, err := sm.r.GetUnbondingPeriod(chainName)
	if err != nil {
		return err
	}

	clientState := &tmclient.ClientState{
		ChainId:         ibcHeader.Header.ChainID,
		TrustLevel:      tmclient.NewFractionFromTm(light.DefaultTrustLevel),
		TrustingPeriod:  time.Duration(int64(unbondingPeriod) / 100 * 85),
		UnbondingPeriod: unbondingPeriod,
		MaxClockDrift:   10 * time.Minute,
		FrozenHeight:    clienttypes.ZeroHeight(),
		LatestHeight: clienttypes.Height{
			RevisionNumber: revisionNumber,
			RevisionHeight: uint64(ibcHeader.SignedHeader.Header.Height),
		},
		ProofSpecs:  commitmenttypes.GetSDKSpecs(),
		UpgradePath: defaultUpgradePath,
	}

	chainStorage := sm.storage.GetChainStorage(chainName)
	ctx := sdk.NewContext(sm.storage.GetRootStore(), *ibcHeader.Header, false, sm.sdkLogger)

	sm.logger.Debug("Creating tendermint light client", zap.Int64("height", ibcHeader.SignedHeader.Header.Height))

	if err != nil {
		return err
	}

	return chainStorage.CreateLightClient(ctx, clientState, ibcHeader.ConsensusState())
}

func (sm *SoloMachine) UpdateLightClient(chainName string) error {
	ibcHeader, err := sm.r.GetLatestIBCHeader(chainName)
	if err != nil {
		return err
	}
	chainStorage := sm.storage.GetChainStorage(chainName)
	ctx := sdk.NewContext(sm.storage.GetRootStore(), *ibcHeader.Header, false, sm.sdkLogger)

	chainStorage.UpdateLightClient(ctx, ibcHeader)

	sm.logger.Info("Updated tendermint light client", zap.String("chain-name", chainName))

	return nil
}
