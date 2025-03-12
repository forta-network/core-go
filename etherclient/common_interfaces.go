package etherclient

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// TraceCallTransaction contains the fields of the to-be-simulated transaction.
type TraceCallTransaction struct {
	From     string       `json:"from"`
	To       string       `json:"to"`
	Gas      *hexutil.Big `json:"gas,omitempty"`
	GasPrice *hexutil.Big `json:"gasPrice,omitempty"`
	Value    *hexutil.Big `json:"value,omitempty"`
	Data     string       `json:"data"`
}

// TraceCallConfig contains the tracer configuration to be used while simulating the transaction.
type TraceCallConfig struct {
	Tracer         string                 `json:"tracer,omitempty"`
	TracerConfig   *TracerConfig          `json:"tracerConfig,omitempty"`
	StateOverrides map[string]interface{} `json:"stateOverrides,omitempty"`
}

// TracerConfig contains some extra tracer parameters.
type TracerConfig struct {
	WithLog     bool `json:"withLog,omitempty"`
	OnlyTopCall bool `json:"onlyTopCall,omitempty"`
}

// TracedCall contains traced call data. This also represents the top level object
// in the debug_traceCall response.
type TracedCall struct {
	From     common.Address  `json:"from"`
	To       common.Address  `json:"to"`
	CallType string          `json:"type"`
	GasUsed  *hexutil.Big    `json:"gasUsed"`
	Input    string          `json:"input"`
	Output   string          `json:"output"`
	Error    string          `json:"error"`
	Calls    []*TracedCall   `json:"calls"`
	Logs     []*TracedLog    `json:"logs"`
	Raw      json.RawMessage `json:"-"`
	Value    *hexutil.Big    `json:"value"`
}

// TracedLog contains log data from trace.
type TracedLog struct {
	Index   int            `json:"index"`
	Address common.Address `json:"address"`
	Topics  []string       `json:"topics"`
	Data    hexutil.Bytes  `json:"data"`
}
