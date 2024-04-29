package solomachine

import (
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	solomachineclient "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	"github.com/gjermundgaraba/solo-machine/relayer"
	smstorage "github.com/gjermundgaraba/solo-machine/solomachine/storage"
	"github.com/gjermundgaraba/solo-machine/utils"
	"go.uber.org/zap"
	"path/filepath"
	"time"
)

type SoloMachine struct {
	logger    *zap.Logger
	sdkLogger utils.SDKLoggerWrapper
	cdc       codec.Codec

	r *relayer.Relayer

	storage *smstorage.Storage
}

func NewSoloMachine(logger *zap.Logger, cdc codec.Codec, r *relayer.Relayer, homedir string) *SoloMachine {
	dbName := "solo-machine"
	db, err := dbm.NewGoLevelDB(dbName, homedir, nil)
	if err != nil {
		panic(err)
	}
	logger.Debug("opened database", zap.String("path", filepath.Join(homedir, dbName+dbm.DBFileSuffix)))

	storage := smstorage.NewStorage(db, logger, cdc)

	sm := &SoloMachine{
		logger:    logger,
		sdkLogger: utils.NewSDKLoggerWrapper(logger),
		cdc:       cdc,

		r: r,

		storage: storage,
	}

	return sm
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
	sig, err := sm.storage.Sign(bz)
	if err != nil {
		return nil, err
	}
	signatureData := &signing.SingleSignatureData{
		Signature: sig,
	}
	protoSigData := signing.SignatureDataToProto(signatureData)
	return sm.cdc.Marshal(protoSigData)
}

// GenerateCommitmentProof generates a commitment proof for the provided packet.
func (sm *SoloMachine) GenerateCommitmentProof(chainName string, packet channeltypes.Packet, sequence uint64) ([]byte, error) {
	chainStorage := sm.storage.GetChainStorage(chainName)
	commitment := channeltypes.CommitPacket(sm.cdc, packet)

	path := host.PacketCommitmentKey(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	signBytes := &solomachineclient.SignBytes{
		Sequence:    sequence,
		Timestamp:   uint64(time.Now().UnixMilli()),
		Diversifier: chainStorage.Diversifier(),
		Path:        path,
		Data:        commitment,
	}

	return sm.GenerateProof(signBytes)
}
