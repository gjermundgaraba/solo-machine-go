package storage

import (
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gjermundgaraba/solo-machine-go/utils"
	"go.uber.org/zap"
)

const (
	soloMachineStorePrefix = "solo-machine-store"
	chainsStoragePrefix    = "chains"

	privateKeyKey            = "private-key"
	nextLightClientNumberKey = "next-light-client-number"
	nextConnectionNumberKey  = "next-connection-number"
	nextChannelNumberKey     = "next-channel-number"
)

type Storage struct {
	logger *zap.Logger
	store  *rootmulti.Store
	cdc    codec.Codec

	privateKey            cryptotypes.PrivKey
	publicKey             cryptotypes.PubKey
	nextLightClientNumber uint64
	nextConnectionNumber  uint64
	nextChannelNumber     uint64

	soloMachineStoreKey storetypes.StoreKey
	rootChainsStoreKey  storetypes.StoreKey
}

func NewStorage(db dbm.DB, logger *zap.Logger, cdc codec.Codec) *Storage {
	sdkLogger := utils.NewSDKLoggerWrapper(logger)
	store := rootmulti.NewStore(db, sdkLogger, metrics.NewNoOpMetrics())

	soloMachineStoreKey := storetypes.NewKVStoreKey(soloMachineStorePrefix)
	rootChainsStoreKey := storetypes.NewKVStoreKey(chainsStoragePrefix)

	store.MountStoreWithDB(soloMachineStoreKey, storetypes.StoreTypeIAVL, nil)
	store.MountStoreWithDB(rootChainsStoreKey, storetypes.StoreTypeIAVL, nil)

	if err := store.LoadLatestVersion(); err != nil {
		panic(err)
	}

	// We just need to make a temporary instance to get the storage portion out
	// TODO: Find a better way to do this
	soloMachineStorage := (&Storage{
		store:               store,
		soloMachineStoreKey: soloMachineStoreKey,
	}).getSoloMachineStorage()

	privKeyBz := soloMachineStorage.Get([]byte(privateKeyKey))
	if privKeyBz == nil {
		privKeyBz = secp256k1.GenPrivKey().Bytes()
		soloMachineStorage.Set([]byte(privateKeyKey), privKeyBz)
	}
	privateKey := &secp256k1.PrivKey{
		Key: privKeyBz,
	}
	publicKey := privateKey.PubKey()

	nextLightClientNumber := uint64(0)
	nextLightClientNumberBz := soloMachineStorage.Get([]byte(nextLightClientNumberKey))
	if nextLightClientNumberBz != nil {
		nextLightClientNumber = sdk.BigEndianToUint64(nextLightClientNumberBz)
	}

	nextConnectionNumber := uint64(0)
	nextConnectionNumberBz := soloMachineStorage.Get([]byte(nextConnectionNumberKey))
	if nextConnectionNumberBz != nil {
		nextConnectionNumber = sdk.BigEndianToUint64(nextConnectionNumberBz)
	}

	nextChannelNumber := uint64(0)
	nextChannelNumberBz := soloMachineStorage.Get([]byte(nextChannelNumberKey))
	if nextChannelNumberBz != nil {
		nextChannelNumber = sdk.BigEndianToUint64(nextChannelNumberBz)
	}

	return &Storage{
		logger: logger,
		store:  store,
		cdc:    cdc,

		privateKey:            privateKey,
		publicKey:             publicKey,
		nextLightClientNumber: nextLightClientNumber,
		nextConnectionNumber:  nextConnectionNumber,
		nextChannelNumber:     nextChannelNumber,

		soloMachineStoreKey: soloMachineStoreKey,
		rootChainsStoreKey:  rootChainsStoreKey,
	}
}

func (s *Storage) Commit() {
	s.store.Commit()
}

func (s *Storage) GetPublicKey() cryptotypes.PubKey {
	return s.publicKey
}

func (s *Storage) Sign(data []byte) ([]byte, error) {
	return s.privateKey.Sign(data)
}

func (s *Storage) getSoloMachineStorage() storetypes.CommitKVStore {
	return s.store.GetCommitKVStore(s.soloMachineStoreKey)
}

// GetRootStore returns the root store of the storage
// Do not use directly for any other purpose than creating sdk.Context
func (s *Storage) GetRootStore() storetypes.MultiStore {
	return s.store
}

func (s *Storage) incrementNextLightClientNumber() {
	s.nextLightClientNumber++
	s.getSoloMachineStorage().Set([]byte(nextLightClientNumberKey), sdk.Uint64ToBigEndian(s.nextLightClientNumber))
	s.Commit()
}

func (s *Storage) incrementNextConnectionNumber() {
	s.nextConnectionNumber++
	s.getSoloMachineStorage().Set([]byte(nextConnectionNumberKey), sdk.Uint64ToBigEndian(s.nextConnectionNumber))
	s.Commit()
}

func (s *Storage) incrementNextChannelNumber() {
	s.nextChannelNumber++
	s.getSoloMachineStorage().Set([]byte(nextChannelNumberKey), sdk.Uint64ToBigEndian(s.nextChannelNumber))
	s.Commit()
}
