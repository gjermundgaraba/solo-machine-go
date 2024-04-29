package relayer

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/gjermundgaraba/solo-machine/utils"
	"go.uber.org/zap"
	"os"
	"time"
)

func (r *Relayer) createClientCtx(chainName string) client.Context {
	chainConfig, ok := r.chains[chainName]
	if !ok {
		panic(fmt.Sprintf("Chain %s not found in relayer config", chainName))
	}

	cfg := sdk.GetConfig()
	accountPubKeyPrefix := chainConfig.AccountPrefix + "pub"
	validatorAddressPrefix := chainConfig.AccountPrefix + "valoper"
	validatorPubKeyPrefix := chainConfig.AccountPrefix + "valoperpub"
	consNodeAddressPrefix := chainConfig.AccountPrefix + "valcons"
	consNodePubKeyPrefix := chainConfig.AccountPrefix + "valconspub"
	cfg.SetBech32PrefixForAccount(chainConfig.AccountPrefix, accountPubKeyPrefix)
	cfg.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	cfg.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	//cfg.Seal()

	kr, err := utils.GetKeyring(chainConfig.KeyringBackend, r.homedir, r.cdc)
	if err != nil {
		panic(err)
	}

	txCfg := authtx.NewTxConfig(r.cdc, authtx.DefaultSignModes)

	from, err := kr.Key(chainConfig.KeyName)
	if err != nil {
		panic(err)
	}
	fromAddr, err := from.GetAddress()
	if err != nil {
		panic(err)
	}

	rpcClient, err := client.NewClientFromNode(chainConfig.RPCAddr)
	if err != nil {
		panic(err)
	}

	return client.Context{}.
		WithCmdContext(r.ctx).
		WithCodec(r.cdc).
		WithInterfaceRegistry(r.cdc.InterfaceRegistry()).
		WithTxConfig(txCfg).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithHomeDir(r.homedir).
		WithChainID(chainConfig.ChainID).
		WithKeyring(kr).
		WithOffline(false).
		WithNodeURI(chainConfig.RPCAddr).
		WithFromName(chainConfig.KeyName).
		WithFromAddress(fromAddr).
		WithFrom(fromAddr.String()).
		WithClient(rpcClient).
		WithBroadcastMode("sync").
		WithViper("")
}

func (r *Relayer) createTxFactory(clientCtx client.Context, chainName string) tx.Factory {
	chainConfig := r.chains[chainName]

	gasSetting, err := flags.ParseGasSetting(chainConfig.Gas)
	if err != nil {
		panic(err)
	}

	return tx.Factory{}.
		WithTxConfig(clientCtx.TxConfig).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithKeybase(clientCtx.Keyring).
		WithChainID(clientCtx.ChainID).
		WithFromName(clientCtx.FromName).
		WithGas(gasSetting.Gas).
		WithGasPrices(chainConfig.GasPrices).
		WithSimulateAndExecute(gasSetting.Simulate).
		WithGasAdjustment(chainConfig.GasAdjustment).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)
}

// A good amount of this is copied from cosmos sdk client code
func (r *Relayer) sendTx(clientCtx client.Context, txf tx.Factory, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	for _, msg := range msgs {
		m, ok := msg.(sdk.HasValidateBasic)
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return nil, err
		}
	}

	var err error
	txf, err = txf.Prepare(clientCtx)
	if err != nil {
		r.logger.Error("Failed to prepare tx", zap.Error(err))
		return nil, err
	}

	adjusted, err := calculateGas(clientCtx, txf, msgs...)
	if err != nil {
		return nil, err
	}
	txf = txf.WithGas(adjusted)
	r.logger.Debug("Estimated gas", zap.Uint64("Gas", txf.Gas()))

	builtTx, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	if err = tx.Sign(clientCtx.CmdContext, txf, clientCtx.FromName, builtTx, true); err != nil {
		return nil, err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(builtTx.GetTx())
	if err != nil {
		return nil, err
	}

	// broadcast to a CometBFT node
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	if res.Code != 0 {
		return nil, fmt.Errorf(res.RawLog)
	}

	if _, err = clientCtx.Codec.MarshalJSON(res); err != nil {
		return nil, err
	}
	var msgTypes []string
	for _, msg := range msgs {
		msgTypes = append(msgTypes, sdk.MsgTypeURL(msg))

	}
	r.logger.Info("Successfully broadcast tx", zap.Strings("msgs", msgTypes))

	return r.waitForTX(clientCtx, res.TxHash)
}

func calculateGas(clientCtx gogogrpc.ClientConn, txf tx.Factory, msgs ...sdk.Msg) (uint64, error) {
	txBytes, err := txf.BuildSimTx(msgs...)
	if err != nil {
		return 0, err
	}

	txSvcClient := txtypes.NewServiceClient(clientCtx)
	simRes, err := txSvcClient.Simulate(context.Background(), &txtypes.SimulateRequest{
		TxBytes: txBytes,
	})
	if err != nil {
		return 0, err
	}

	return uint64(txf.GasAdjustment() * float64(simRes.GasInfo.GasUsed)), nil
}

func (r *Relayer) waitForTX(clientCtx client.Context, txHash string) (*sdk.TxResponse, error) {
	r.logger.Debug("Starting to wait for tx", zap.String("tx_hash", txHash))
	try := 1
	maxTries := 25
	for {
		txResp, err := authtx.QueryTx(clientCtx, txHash)
		if err != nil {
			if try == maxTries {
				err2 := fmt.Errorf("transaction with hash %s exceeded max retry limit of %d with error %s", txHash, try, err)
				r.logger.Error("Transaction not found", zap.Error(err))
				return nil, err2
			}

			r.logger.Debug("Waiting for transaction", zap.String("tx_hash", txHash), zap.Int("try", try), zap.Error(err))
			time.Sleep(500 * time.Millisecond)
			try++
			continue
		}

		if txResp.Code != 0 {
			return nil, fmt.Errorf("transaction failed: %s", txResp.RawLog)
		}

		r.logger.Info("Transaction succeeded", zap.String("tx_hash", txHash))
		return txResp, nil
	}
}
