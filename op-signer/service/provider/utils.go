package provider

import (
	"context"
	"errors"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	fmt.Printf("PEM:\n%s", result.Pem)

	key := []byte(result.Pem)
	if int64(crc32c(key)) != result.PemCrc32C.Value {
		return nil, errors.New("cloud kms public key response corrupted in transit")
	}

	uncompressed, err := decodePublicKeyPEM(key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	fmt.Printf("HEX:\n%x\n", uncompressed)
	hash := crypto.Keccak256(uncompressed[1:])
	addr := common.BytesToAddress(hash[len(hash)-20:])
	fmt.Printf("Addr:\n%s\n", addr)
	return &addr, nil
}
