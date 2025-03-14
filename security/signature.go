package security

import (
	"errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// DecodeEthereumSignature decodes an Ethereum signature.
func DecodeEthereumSignature(sigHex string) ([]byte, error) {
	sig, err := hexutil.Decode(sigHex)
	if err != nil {
		return nil, err
	}

	// logic from https://github.com/ethereum/go-ethereum/blob/55599ee95d4151a2502465e0afc7c47bd1acba77/internal/ethapi/api.go#L442
	if sig[64] != 27 && sig[64] != 28 {
		return nil, errors.New("invalid Ethereum signature (V is not 27 or 28)")
	}
	sig[64] -= 27

	return sig, nil
}

// EncodeEthereumSignature encodes an Ethereum signature.
func EncodeEthereumSignature(sig []byte) (string, error) {
	// logic from https://github.com/ethereum/go-ethereum/blob/55599ee95d4151a2502465e0afc7c47bd1acba77/internal/ethapi/api.go#L442
	if sig[64] != 0 && sig[64] != 1 {
		return "", errors.New("invalid Ethereum signature (V is not 27 or 28)")
	}
	sig[64] += 27

	return hexutil.Encode(sig), nil
}
