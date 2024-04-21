package cosmoschain

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"go.uber.org/zap"
	"time"
)

// A good amount of this is copied from cosmos sdk client code

func (cc *CosmosChain) sendTx(msgs ...sdk.Msg) (string, error) {
	for _, msg := range msgs {
		m, ok := msg.(sdk.HasValidateBasic)
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return "", err
		}
	}

	var err error
	txf := cc.newTxFactory()
	txf, err = txf.Prepare(cc.clientCtx)
	if err != nil {
		cc.logger.Error("Failed to prepare tx", zap.Error(err))
		return "", err
	}

	adjusted, err := calculateGas(cc.clientCtx, txf, msgs...)
	if err != nil {
		return "", err
	}
	txf = txf.WithGas(adjusted)
	cc.logger.Debug("Estimated gas", zap.Uint64("Gas", txf.Gas()))

	builtTx, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return "", err
	}

	if err = tx.Sign(cc.clientCtx.CmdContext, txf, cc.clientCtx.FromName, builtTx, true); err != nil {
		return "", err
	}

	txBytes, err := cc.clientCtx.TxConfig.TxEncoder()(builtTx.GetTx())
	if err != nil {
		return "", err
	}

	// broadcast to a CometBFT node
	res, err := cc.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return "", err
	}

	if res.Code != 0 {
		return "", fmt.Errorf(res.RawLog)
	}

	out, err := cc.clientCtx.Codec.MarshalJSON(res)
	if err != nil {
		return "", err
	}
	cc.logger.Info("Successfully broadcast tx", zap.String("tx", string(out)))

	return res.TxHash, cc.waitForTX(res.TxHash)
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

func (cc *CosmosChain) newTxFactory() tx.Factory {
	gasSetting, err := flags.ParseGasSetting(cc.gas)
	if err != nil {
		panic(err)
	}

	return tx.Factory{}.
		WithTxConfig(cc.clientCtx.TxConfig).
		WithAccountRetriever(cc.clientCtx.AccountRetriever).
		WithKeybase(cc.clientCtx.Keyring).
		WithChainID(cc.clientCtx.ChainID).
		WithFromName(cc.clientCtx.FromName).
		WithGas(gasSetting.Gas).
		WithGasPrices(cc.gasPrices).
		WithSimulateAndExecute(gasSetting.Simulate).
		WithGasAdjustment(cc.gasAdjustment).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)
}

func (cc *CosmosChain) waitForTX(txHash string) error {
	cc.logger.Debug("Starting to wait for tx", zap.String("tx_hash", txHash))
	try := 1
	maxTries := 25
	for {
		txResp, err := authtx.QueryTx(cc.clientCtx, txHash)
		if err != nil {
			if try == maxTries {
				err2 := fmt.Errorf("transaction with hash %s exceeded max retry limit of %d with error %s", txHash, try, err)
				cc.logger.Error("Transaction not found", zap.Error(err))
				return err2
			}

			cc.logger.Debug("Waiting for transaction", zap.String("tx_hash", txHash), zap.Int("try", try), zap.Error(err))
			time.Sleep(500 * time.Millisecond)
			try++
			continue
		}

		if txResp.Code != 0 {
			return fmt.Errorf("transaction failed: %s", txResp.RawLog)
		}

		cc.logger.Info("Transaction succeeded", zap.String("tx_hash", txHash))
		return nil
	}
}
