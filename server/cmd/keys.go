package cmd

import (
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/spf13/cobra"
)

func KeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "TODO", // TODO
	}

	cmd.AddCommand(generateKeysCmd())
	cmd.AddCommand(signTextCmd())
	cmd.AddCommand(verifySignatureCmd())

	return cmd
}

func generateKeysCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "TODO", // TODO
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey := secp256k1.GenPrivKey()
			pubKey := privKey.PubKey()

			privKeyHex := hex.EncodeToString(privKey.Bytes())
			pubKeyHex := hex.EncodeToString(pubKey.Bytes())

			cmd.Println("Private Key:", privKeyHex)
			cmd.Println("Public Key:", pubKeyHex)

			return nil
		},
	}
}

func signTextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sign [text] [private key]",
		Short: "TODO", // TODO,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[0]
			privKeyHex := args[1]

			privKeyBytes, err := hex.DecodeString(privKeyHex)
			if err != nil {
				return err
			}

			privKey := secp256k1.PrivKey{
				Key: privKeyBytes,
			}
			sign, err := privKey.Sign([]byte(text))
			if err != nil {
				return err
			}

			signHex := hex.EncodeToString(sign)
			cmd.Println("Signature:", signHex)

			return nil
		},
	}
}

func verifySignatureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify [text] [signature] [public key]",
		Short: "TODO", // TODO,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[0]
			signHex := args[1]
			pubKeyHex := args[2]

			sign, err := hex.DecodeString(signHex)
			if err != nil {
				return err
			}

			pubKeyBytes, err := hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			pubKey := secp256k1.PubKey{
				Key: pubKeyBytes,
			}
			isValid := pubKey.VerifySignature([]byte(text), sign)
			cmd.Println("Is valid:", isValid)

			return nil
		},
	}
}
