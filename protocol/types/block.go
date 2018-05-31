package types

// BlockHeader ss
type BlockHeader struct {
	Version                uint64
	Height                 uint64
	Timestamp              uint64
	PreviousBlockHash      Hash
	TransactionsMerkleRoot Hash
	TransactionsStatusHash Hash
	Validator              Address
}

// Block ss
type Block struct {
	BlockHeader
	Transactions []*Transaction
}
