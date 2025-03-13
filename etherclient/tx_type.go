package etherclient

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// TxType is a Go type for known EVM chain tx types.
type TxType int

func (typ *TxType) UnmarshalJSON(input []byte) error {
	v, err := hexutil.DecodeBig(string(input[1 : len(input)-1]))
	if err != nil {
		return fmt.Errorf("failed to unmarshal tx type: %v", err)
	}
	*typ = TxType(v.Int64())
	return nil
}

func (typ TxType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, hexutil.EncodeUint64(uint64(typ)))), nil
}

const (
	LegacyTxType     TxType = 0x00
	AccessListTxType TxType = 0x01
	DynamicFeeTxType TxType = 0x02
	BlobTxType       TxType = 0x03
	SetCodeTxType    TxType = 0x04

	OptimismDepositTxType TxType = 0x7E

	ArbitrumDepositTxType         = 0x64
	ArbitrumUnsignedTxType        = 0x65
	ArbitrumContractTxType        = 0x66
	ArbitrumRetryTxType           = 0x68
	ArbitrumSubmitRetryableTxType = 0x69
	ArbitrumInternalTxType        = 0x6A
	ArbitrumLegacyTxType          = 0x78
)

func (typ TxType) String() string {
	switch typ {
	case LegacyTxType:
		return "legacyTx"
	case AccessListTxType:
		return "accessListTx"
	case DynamicFeeTxType:
		return "dynamicFeeTx"
	case BlobTxType:
		return "blobTx"
	case SetCodeTxType:
		return "setCodeTx"

	case OptimismDepositTxType:
		return "optimismDepositTx"

	case ArbitrumDepositTxType:
		return "arbitrumDepositTx"
	case ArbitrumUnsignedTxType:
		return "arbitrumUnsignedTx"
	case ArbitrumContractTxType:
		return "arbitrumContractTx"
	case ArbitrumRetryTxType:
		return "arbitrumRetryTx"
	case ArbitrumSubmitRetryableTxType:
		return "arbitrumSubmitRetryableTx"
	case ArbitrumInternalTxType:
		return "arbitrumInternalTx"
	case ArbitrumLegacyTxType:
		return "arbitrumLegacyTx"

	default:
		return "unknownTx"
	}
}
