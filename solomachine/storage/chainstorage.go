package storage

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	tmclient "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"go.uber.org/zap"
	"math/rand/v2"
)

const (
	lightClientPrefix = "light-clients"

	diversifierKey              = "diversifier"
	counterpartyClientIDKey     = "counterparty-client-id"
	counterpartyConnectionIDKey = "counterparty-connection-id"
	clientIDKey                 = "client-id"
	connectionIDKey             = "connection-id"

	ics20ChannelKey             = "ics20-channel"
	counterpartyICS20ChannelKey = "counterparty-ics20-channel"
)

var _ exported.ClientStoreProvider = &ChainStorage{}

type ChainStorage struct {
	parent *Storage
	logger *zap.Logger
	store  *prefix.Store

	diversifier              string
	counterpartyClientID     string
	counterpartyConnectionID string
	clientID                 string
	connectionID             string
	ics20Channel             string
	counterpartyICS20Channel string

	tmLightClientModule tmclient.LightClientModule
}

func (s *Storage) GetChainStorage(chainName string) *ChainStorage {
	rootChainsStore := s.store.GetCommitKVStore(s.rootChainsStoreKey)
	chainStore := prefix.NewStore(rootChainsStore, []byte(chainName))

	if !chainStore.Has([]byte(diversifierKey)) {
		charset := "abcdefghijklmnopqrstuvwxyz"
		randomBytes := make([]byte, 15)
		for i := range randomBytes {
			randomBytes[i] = charset[rand.IntN(len(charset))]
		}

		chainStore.Set([]byte(diversifierKey), randomBytes)
	}

	diversifier := string(chainStore.Get([]byte(diversifierKey)))

	counterpartyClientID := ""
	counterpartyClientIDBz := chainStore.Get([]byte(counterpartyClientIDKey))
	if counterpartyClientIDBz != nil {
		counterpartyClientID = string(counterpartyClientIDBz)
	}

	counterpartyConnectionID := ""
	counterpartyConnectionIDBz := chainStore.Get([]byte(counterpartyConnectionIDKey))
	if counterpartyConnectionIDBz != nil {
		counterpartyConnectionID = string(counterpartyConnectionIDBz)
	}

	clientID := ""
	clientIDBz := chainStore.Get([]byte(clientIDKey))
	if clientIDBz != nil {
		clientID = string(clientIDBz)
	}

	connectionID := ""
	connectionIDBz := chainStore.Get([]byte(connectionIDKey))
	if connectionIDBz != nil {
		connectionID = string(connectionIDBz)
	}

	ics20Channel := ""
	ics20ChannelBz := chainStore.Get([]byte(ics20ChannelKey))
	if ics20ChannelBz != nil {
		ics20Channel = string(ics20ChannelBz)
	}

	counterpartyICS20Channel := ""
	counterpartyICS20ChannelBz := chainStore.Get([]byte(counterpartyICS20ChannelKey))
	if counterpartyICS20ChannelBz != nil {
		counterpartyICS20Channel = string(counterpartyICS20ChannelBz)
	}

	cs := &ChainStorage{
		parent: s,
		logger: s.logger,
		store:  &chainStore,

		diversifier:              diversifier,
		counterpartyClientID:     counterpartyClientID,
		counterpartyConnectionID: counterpartyConnectionID,
		clientID:                 clientID,
		connectionID:             connectionID,
		ics20Channel:             ics20Channel,
		counterpartyICS20Channel: counterpartyICS20Channel,

		// Setting up the light client module below because it needs a ref to the chain storage because it implements exported.ClientStoreProvider
	}

	tmLightClient := tmclient.NewLightClientModule(s.cdc, "dummy") // Authority does not matter here (?)
	tmLightClient.RegisterStoreProvider(cs)

	cs.tmLightClientModule = tmLightClient

	return cs
}

func (cs *ChainStorage) Diversifier() string {
	return cs.diversifier
}

func (cs *ChainStorage) ClientID() string {
	return cs.clientID
}

func (cs *ChainStorage) CounterpartyClientID() string {
	return cs.counterpartyClientID
}

func (cs *ChainStorage) setClientID(clientID string) {
	cs.store.Set([]byte(clientIDKey), []byte(clientID))
	cs.clientID = clientID
	cs.parent.Commit()
}

func (cs *ChainStorage) SetCounterPartyClientID(clientID string) {
	cs.store.Set([]byte(counterpartyClientIDKey), []byte(clientID))
	cs.counterpartyClientID = clientID
	cs.parent.Commit()
}

// ClientStore implements exported.ClientStoreProvider
// Is used by ibc light client module
func (cs *ChainStorage) ClientStore(ctx sdk.Context, clientID string) storetypes.KVStore {
	lightClientStore := prefix.NewStore(cs.store, []byte(lightClientPrefix))
	return prefix.NewStore(lightClientStore, []byte(clientID))
}

func (cs *ChainStorage) LightClientExists() bool {
	return cs.clientID != ""
}

func (cs *ChainStorage) CreateLightClient(ctx sdk.Context, clientState exported.ClientState, consensusState exported.ConsensusState) error {
	if cs.LightClientExists() {
		panic("light client already exists")
	}

	anyClientState, err := clienttypes.PackClientState(clientState)
	if err != nil {
		return err
	}

	anyConsensusState, err := clienttypes.PackConsensusState(consensusState)
	if err != nil {
		return err
	}

	nextClientSeq := cs.parent.nextLightClientNumber
	defer cs.parent.incrementNextLightClientNumber()

	clientID := clienttypes.FormatClientIdentifier(exported.Tendermint, nextClientSeq)
	if err := cs.tmLightClientModule.Initialize(ctx, clientID, anyClientState.Value, anyConsensusState.Value); err != nil {
		return err
	}

	cs.setClientID(clientID)

	cs.logger.Info("Initialized tendermint light client", zap.Any("client-id", clientID))

	return nil
}

func (cs *ChainStorage) UpdateLightClient(ctx sdk.Context, ibcHeader tmclient.Header) {
	cs.tmLightClientModule.UpdateState(ctx, cs.clientID, &ibcHeader)
	cs.parent.Commit()
	cs.logger.Info("Updated tendermint light client", zap.Any("client-id", cs.clientID))
}

func (cs *ChainStorage) LightClientState() (*tmclient.ClientState, error) {
	clientStore := cs.ClientStore(sdk.Context{}, cs.clientID)
	bz := clientStore.Get(host.ClientStateKey())
	if len(bz) == 0 {
		return nil, fmt.Errorf("light client state not found")
	}

	clientStateI := clienttypes.MustUnmarshalClientState(cs.parent.cdc, bz)
	clientState, ok := clientStateI.(*tmclient.ClientState)
	if !ok {
		return nil, fmt.Errorf("cannot convert %T into %T", clientStateI, clientState)
	}

	return clientState, nil
}

func (cs *ChainStorage) GetLightConsensusState(height exported.Height) (*tmclient.ConsensusState, error) {
	clientStore := cs.ClientStore(sdk.Context{}, cs.clientID)
	bz := clientStore.Get(host.ConsensusStateKey(height))
	if len(bz) == 0 {
		return nil, fmt.Errorf("consensus state not found for height: %s", height)
	}

	consensusStateI := clienttypes.MustUnmarshalConsensusState(cs.parent.cdc, bz)
	var consensusState *tmclient.ConsensusState
	consensusState, ok := consensusStateI.(*tmclient.ConsensusState)
	if !ok {
		return nil, fmt.Errorf("cannot convert %T into %T", consensusStateI, consensusState)
	}

	return consensusState, nil
}
