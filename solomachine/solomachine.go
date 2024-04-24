package solomachine

import (
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	"cosmossdk.io/store/types"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	solomachineclient "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	tmclient "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"go.uber.org/zap"
	"path/filepath"
	"time"
)

const (
	lightClientStorePrefix = "light-client-store"
	soloMachineStorePrefix = "solo-machine-store"

	privateKeyKey  = "private-key"
	diversifierKey = "diversifier"
)

type SoloMachine struct {
	logger    *zap.Logger
	sdkLogger sdkloggerwrapper

	store               *rootmulti.Store
	cdc                 *codec.ProtoCodec
	lightClientStoreKey storetypes.StoreKey
	soloMachineStoreKey storetypes.StoreKey

	PrivateKey  cryptotypes.PrivKey
	PublicKey   cryptotypes.PubKey
	Diversifier string

	tmLightClient tmclient.LightClientModule
}

var _ exported.ClientStoreProvider = &SoloMachine{}

// ClientStore implement exported.ClientStoreProvider
func (sm *SoloMachine) ClientStore(ctx sdk.Context, clientID string) types.KVStore {
	return sm.store.GetCommitKVStore(sm.lightClientStoreKey)
}

func NewSoloMachine(logger *zap.Logger, homedir string) *SoloMachine {
	sdkLogger := sdkloggerwrapper{
		zapLogger: logger,
	}
	dbName := "solo-machine"
	db, err := dbm.NewGoLevelDB(dbName, homedir, nil)
	if err != nil {
		panic(err)
	}
	logger.Debug("opened database", zap.String("path", filepath.Join(homedir, dbName+dbm.DBFileSuffix)))

	store := rootmulti.NewStore(db, sdkLogger, metrics.NewNoOpMetrics())
	lightClientStoreKey := storetypes.NewKVStoreKey(lightClientStorePrefix)
	soloMachineStoreKey := storetypes.NewKVStoreKey(soloMachineStorePrefix)
	store.MountStoreWithDB(lightClientStoreKey, storetypes.StoreTypeIAVL, nil)
	store.MountStoreWithDB(soloMachineStoreKey, storetypes.StoreTypeIAVL, nil)
	if err := store.LoadLatestVersion(); err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)

	smStore := store.GetCommitKVStore(soloMachineStoreKey)

	changesMade := false
	if !smStore.Has([]byte(privateKeyKey)) {
		privKey := secp256k1.GenPrivKey()
		privKeyBytes := privKey.Bytes()
		smStore.Set([]byte(privateKeyKey), privKeyBytes)
		changesMade = true
	}
	if !smStore.Has([]byte(diversifierKey)) {
		smStore.Set([]byte(diversifierKey), []byte("diversestuff"))
		changesMade = true
	}
	if changesMade {
		store.Commit()
	}

	privKeyBytes := smStore.Get([]byte(privateKeyKey))
	privKey := secp256k1.PrivKey{
		Key: privKeyBytes,
	}
	pk := privKey.PubKey()

	divBytes := smStore.Get([]byte(diversifierKey))
	div := string(divBytes)

	logger.Debug("loaded solo machine state",
		zap.String("diversifier", div),
		zap.Binary("public-key binary", pk.Bytes()),
		zap.String("public-key string", pk.String()),
	)

	sm := &SoloMachine{
		logger:    logger,
		sdkLogger: sdkLogger,

		store:               store,
		lightClientStoreKey: lightClientStoreKey,
		soloMachineStoreKey: soloMachineStoreKey,

		PrivateKey:  &privKey,
		PublicKey:   pk,
		Diversifier: div,
		// light client is set below because we need a reference to the SoloMachine
	}

	cdc := createCodec()
	sm.cdc = cdc
	tmLightClient := tmclient.NewLightClientModule(cdc, "dummy")
	tmLightClient.RegisterStoreProvider(sm)

	latestHeight := tmLightClient.LatestHeight(sdk.Context{}, clientID)
	if latestHeight.GetRevisionHeight() == 0 {
		logger.Info("tendermint light client is not initialized yet")
	} else {
		logger.Info("tendermint light client loaded", zap.String("client-id", clientID), zap.Uint64("height", latestHeight.GetRevisionHeight()))
	}

	sm.tmLightClient = tmLightClient

	return sm
}

func createCodec() *codec.ProtoCodec {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	ibcclienttypes.RegisterInterfaces(interfaceRegistry)
	solomachineclient.RegisterInterfaces(interfaceRegistry)
	tmclient.RegisterInterfaces(interfaceRegistry)
	return codec.NewProtoCodec(interfaceRegistry)
}

func (sm *SoloMachine) GetPublicKey() cryptotypes.PubKey {
	return sm.PublicKey
}

// ClientState returns a new solo machine ClientState instance.
func (sm *SoloMachine) ClientState(sequence uint64) (*solomachineclient.ClientState, error) {
	consensusState, err := sm.ConsensusState()
	if err != nil {
		return nil, err
	}
	return solomachineclient.NewClientState(sequence, consensusState), nil
}

// ConsensusState returns a new solo machine ConsensusState instance
func (sm *SoloMachine) ConsensusState() (*solomachineclient.ConsensusState, error) {
	publicKey, err := codectypes.NewAnyWithValue(sm.PublicKey)
	if err != nil {
		return nil, err
	}

	return &solomachineclient.ConsensusState{
		PublicKey:   publicKey,
		Diversifier: sm.Diversifier,
		Timestamp:   uint64(time.Now().UnixMilli()),
	}, nil
}

// GenerateConnOpenTryProof generates the proofTry required for the connection open ack handshake step.
// The clientID, connectionID provided represent the clientID and connectionID created on the counterparty chain, that is the tendermint chain.
func (sm *SoloMachine) GenerateConnOpenTryProof(sequence uint64, counterpartyClientID, counterpartyConnectionID string) ([]byte, error) {
	merklePrefix := commitmenttypes.NewMerklePrefix([]byte(exported.StoreKey))

	counterparty := connectiontypes.NewCounterparty(counterpartyClientID, counterpartyConnectionID, merklePrefix)
	connection := connectiontypes.NewConnectionEnd(connectiontypes.TRYOPEN, clientID, counterparty, []*connectiontypes.Version{connectiontypes.GetCompatibleVersions()[0]}, 0)

	data, err := sm.cdc.Marshal(&connection)
	if err != nil {
		return nil, err
	}

	merklePath := commitmenttypes.NewMerklePath(host.ConnectionPath(connectionID))
	merklePath, err = commitmenttypes.ApplyPrefix(merklePrefix, merklePath)
	if err != nil {
		return nil, err
	}
	// in a multistore context: index 0 is the key for the IBC store in the multistore, index 1 is the key in the IBC store
	key, err := merklePath.GetKey(1)
	if err != nil {
		return nil, err
	}
	signBytes := &solomachineclient.SignBytes{
		Sequence:    sequence,
		Timestamp:   uint64(time.Now().UnixMilli()),
		Diversifier: sm.Diversifier,
		Path:        key,
		Data:        data,
	}

	sm.logger.Debug("generated sign bytes", zap.Uint64("sequence", sequence), zap.Uint64("timestamp", signBytes.Timestamp), zap.String("diversifier", signBytes.Diversifier), zap.String("path", string(signBytes.Path)), zap.String("data", string(signBytes.Data)))

	return sm.GenerateProof(signBytes)
}

// GenerateProof takes in solo machine sign bytes, generates a signature and marshals it as a proof.
func (sm *SoloMachine) GenerateProof(signBytes *solomachineclient.SignBytes) ([]byte, error) {
	bz, err := sm.cdc.Marshal(signBytes)
	if err != nil {
		return nil, err
	}

	sig, err := sm.GenerateSignature(bz)
	if err != nil {
		return nil, err
	}

	signatureDoc := &solomachineclient.TimestampedSignatureData{
		SignatureData: sig,
		Timestamp:     signBytes.Timestamp,
	}
	proof, err := sm.cdc.Marshal(signatureDoc)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

func (sm *SoloMachine) GenerateSignature(bz []byte) ([]byte, error) {
	sig, err := sm.PrivateKey.Sign(bz)
	if err != nil {
		return nil, err
	}
	signatureData := &signing.SingleSignatureData{
		Signature: sig,
	}
	protoSigData := signing.SignatureDataToProto(signatureData)
	return sm.cdc.Marshal(protoSigData)
}

func (sm *SoloMachine) GenerateClientStateProof(sequence uint64, clientState exported.ClientState) ([]byte, error) {
	data, err := ibcclienttypes.MarshalClientState(sm.cdc, clientState)
	if err != nil {
		return nil, err
	}

	path := host.FullClientStateKey(clientID)
	signBytes := &solomachineclient.SignBytes{
		Sequence:    sequence,
		Timestamp:   uint64(time.Now().UnixMilli()),
		Diversifier: sm.Diversifier,
		Path:        path,
		Data:        data,
	}

	return sm.GenerateProof(signBytes)
}

func (sm *SoloMachine) GenerateConsensusStateProof(sequence uint64, clientState *tmclient.ClientState) ([]byte, error) {
	height := clientState.LatestHeight
	consensusState, err := sm.GetLightConsensusState(height)
	if err != nil {
		return nil, err
	}

	data, err := ibcclienttypes.MarshalConsensusState(sm.cdc, consensusState)
	if err != nil {
		return nil, err
	}

	path := host.FullConsensusStateKey(clientID, height)
	signBytes := &solomachineclient.SignBytes{
		Sequence:    sequence,
		Timestamp:   uint64(time.Now().UnixMilli()),
		Diversifier: sm.Diversifier,
		Path:        path,
		Data:        data,
	}

	return sm.GenerateProof(signBytes)
}

// GenerateChanOpenTryProof generates the proofTry required for the channel open ack handshake step.
// The channelID provided represents the channelID created on the counterparty chain, that is the tendermint chain.
func (sm *SoloMachine) GenerateChanOpenTryProof(sequence uint64, portID, version, counterpartyChannelID string) ([]byte, error) {
	counterparty := channeltypes.NewCounterparty(portID, counterpartyChannelID)
	channel := channeltypes.NewChannel(channeltypes.TRYOPEN, channeltypes.UNORDERED, counterparty, []string{connectionID}, version)

	data, err := sm.cdc.Marshal(&channel)
	if err != nil {
		return nil, err
	}

	path := host.ChannelKey(portID, channelID)
	signBytes := &solomachineclient.SignBytes{
		Sequence:    sequence,
		Timestamp:   uint64(time.Now().UnixMilli()),
		Diversifier: sm.Diversifier,
		Path:        path,
		Data:        data,
	}

	return sm.GenerateProof(signBytes)
}

func (sm *SoloMachine) CreateHeader(sequence uint64) (*solomachineclient.Header, error) {
	publicKey, err := codectypes.NewAnyWithValue(sm.PublicKey)
	if err != nil {
		return nil, err
	}

	data := &solomachineclient.HeaderData{
		NewPubKey:      publicKey,
		NewDiversifier: sm.Diversifier,
	}

	dataBz, err := sm.cdc.Marshal(data)
	if err != nil {
		return nil, err
	}

	timestamp := uint64(time.Now().UnixMilli())

	signBytes := &solomachineclient.SignBytes{
		Sequence:    sequence,
		Timestamp:   timestamp,
		Diversifier: sm.Diversifier,
		Path:        []byte(solomachineclient.SentinelHeaderPath),
		Data:        dataBz,
	}

	bz, err := sm.cdc.Marshal(signBytes)
	if err != nil {
		return nil, err
	}

	sig, err := sm.GenerateSignature(bz)
	if err != nil {
		return nil, err
	}

	header := &solomachineclient.Header{
		Timestamp:      timestamp,
		Signature:      sig,
		NewPublicKey:   publicKey,
		NewDiversifier: sm.Diversifier,
	}

	return header, nil
}
