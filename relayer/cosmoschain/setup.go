package cosmoschain

import (
	"context"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	solomachineclient "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	tmclient "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	"os"
)

func setupClientContext(
	cmdCtx context.Context,
	homedir string,
	accountPrefix string,
	keyringBackend string,
	key string,
	rpc string,
	chainID string,
) (client.Context, error) {
	cfg := sdk.GetConfig()
	accountPubKeyPrefix := accountPrefix + "pub"
	validatorAddressPrefix := accountPrefix + "valoper"
	validatorPubKeyPrefix := accountPrefix + "valoperpub"
	consNodeAddressPrefix := accountPrefix + "valcons"
	consNodePubKeyPrefix := accountPrefix + "valconspub"
	cfg.SetBech32PrefixForAccount(accountPrefix, accountPubKeyPrefix)
	cfg.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	cfg.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	cfg.Seal()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	ibcclienttypes.RegisterInterfaces(interfaceRegistry)
	ibcconnectiontypes.RegisterInterfaces(interfaceRegistry)
	ibcchanneltypes.RegisterInterfaces(interfaceRegistry)
	solomachineclient.RegisterInterfaces(interfaceRegistry)
	tmclient.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	kr, err := keyring.New("solorly", keyringBackend, homedir, os.Stdin, cdc)
	if err != nil {
		return client.Context{}, err
	}

	txCfg := tx.NewTxConfig(cdc, tx.DefaultSignModes)

	from, err := kr.Key(key)
	if err != nil {
		return client.Context{}, err
	}
	fromAddr, err := from.GetAddress()
	if err != nil {
		return client.Context{}, err
	}

	rpcClient, err := client.NewClientFromNode(rpc)
	if err != nil {
		return client.Context{}, err
	}

	return client.Context{}.
		WithCmdContext(cmdCtx).
		WithCodec(cdc).
		WithInterfaceRegistry(interfaceRegistry).
		WithTxConfig(txCfg).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(homedir).
		WithChainID(chainID).
		WithKeyring(kr).
		WithOffline(false).
		WithNodeURI(rpc).
		WithFromName(key).
		WithFromAddress(fromAddr).
		WithFrom(fromAddr.String()).
		WithClient(rpcClient).
		WithBroadcastMode("sync").
		WithViper(""), nil
}
