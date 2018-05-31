package types

import (
	"math/big"
	"sync/atomic"
)

// TxInput ss
type TxInput struct {
	Source   Hash
	SourceID uint64 // 0 Send 1 odd
}

// TxOutput ss
// Amount from Transaction
type TxOutput struct {
	StartTime uint64
	EndTime   uint64
}

// Transaction ss
type Transaction struct {
	Version  uint64
	AssetID  AssetID
	Nonce    uint64
	Amount   *big.Int
	GasPrice *big.Int
	GasLimit uint64
	TO       Address
	Payload  []byte
	Inputs   []*TxInput
	Output   TxOutput // just support one output auto gen a odd

	Sign []byte

	// caches
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}
