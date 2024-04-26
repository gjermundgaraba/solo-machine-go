package relayer

import (
	"context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/gjermundgaraba/solo-machine-go/utils"
	"go.uber.org/zap"
)

type Relayer struct {
	ctx     context.Context
	logger  *zap.Logger
	cdc     codec.Codec
	homedir string

	chains map[string]ChainConfig
}

func NewRelayer(ctx context.Context, logger *zap.Logger, cdc codec.Codec, config Config, homedir string) (*Relayer, error) {
	for chainName, chainConfig := range config.Chains {
		keyring, err := utils.GetKeyring(chainConfig.KeyringBackend, homedir, cdc)
		if err != nil {
			return nil, err
		}

		if _, err := keyring.Key(chainConfig.KeyName); err != nil {
			logger.Error("Error getting key from keyring", zap.Error(err), zap.String("chain", chainName), zap.String("key", chainConfig.KeyName))
			return nil, err
		}
	}

	return &Relayer{
		ctx:     ctx,
		logger:  logger,
		cdc:     cdc,
		homedir: homedir,

		chains: config.Chains,
	}, nil
}
