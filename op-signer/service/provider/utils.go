package provider

import (
	"context"
	"errors"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

func PEMPublicKeyToAddress(keyName string) (*common.Address, error) {
	ctx := context.Background()
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, err
	}

	request := kmspb.GetPublicKeyRequest{
		Name: keyName,
	}

	result, err := client.GetPublicKey(ctx, &request)
	if err != nil {
		return nil, fmt.Errorf("kms get public key request failed: %w", err)
	}

	key := []byte(result.Pem)
	if int64(crc32c(key)) != result.PemCrc32C.Value {
		return nil, errors.New("cloud kms public key response corrupted in transit")
	}

	publicKeyBytes, err := decodePublicKeyPEM(key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	hash := crypto.Keccak256(publicKeyBytes)
	addr := common.BytesToAddress(hash[len(hash)-20:])
	return &addr, nil
}

func ToAddr() cli.ActionFunc {
	return func(cliCtx *cli.Context) error {
		keyName := cliCtx.String("key-name")
		addr, err := PEMPublicKeyToAddress(keyName)
		if err != nil {
			return err
		}
		fmt.Println(addr)
		return nil
	}
}
