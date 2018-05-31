package sync

import (
	"github.com/go-fusion/p2p"
)

// TxPoolReactor ss
type TxPoolReactor struct {
	p2p.BaseReactor
}

// NewPoolReactor ss
func NewPoolReactor() *TxPoolReactor {
	txR := &TxPoolReactor{}
	txR.BaseReactor = *p2p.NewBaseReactor("TxPoolReactor", txR)
	return txR
}
