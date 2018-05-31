package sync

import (
	"github.com/go-fusion/p2p"
)

// BlockReactor ss
type BlockReactor struct {
	p2p.BaseReactor
}

// NewBlockReactor ss
func NewBlockReactor() *BlockReactor {
	blR := &BlockReactor{}
	blR.BaseReactor = *p2p.NewBaseReactor("BlockReactor", blR)
	return blR
}
