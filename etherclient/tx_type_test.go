package etherclient

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTxType(t *testing.T) {
	r := require.New(t)

	origTxTypeStr := `"0x7e"` // OpDepositTxType
	var txType TxType
	err := (&txType).UnmarshalJSON([]byte(origTxTypeStr))
	r.NoError(err)
	r.Equal(OptimismDepositTxType, txType)
	b, err := txType.MarshalJSON()
	r.NoError(err)
	r.Equal(origTxTypeStr, string(b))
}
