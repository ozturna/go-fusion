package node

import (
	amino "github.com/tendermint/go-amino"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	cfg "github.com/go-fusion/config"
	"github.com/go-fusion/p2p"
	"github.com/go-fusion/p2p/pex"
	"github.com/go-fusion/sync"
	"github.com/go-fusion/version"
)

//------------------------------------------------------------------------------

// DBContext specifies config information for loading a new DB.
type DBContext struct {
	ID     string
	Config *cfg.Config
}

// DBProvider takes a DBContext and returns an instantiated DB.
type DBProvider func(*DBContext) (dbm.DB, error)

// DefaultDBProvider returns a database using the DBBackend and DBDir
// specified in the ctx.Config.
func DefaultDBProvider(ctx *DBContext) (dbm.DB, error) {
	dbType := dbm.DBBackendType(ctx.Config.DBBackend)
	return dbm.NewDB(ctx.ID, dbType, ctx.Config.DBDir()), nil
}

// Provider takes a config and a logger and returns a ready to go Node.
type Provider func(*cfg.Config, log.Logger) (*Node, error)

// DefaultNewNode returns a Tendermint node with default settings for the
// PrivValidator, ClientCreator, GenesisDoc, and DBProvider.
// It implements NodeProvider.
func DefaultNewNode(config *cfg.Config, logger log.Logger) (*Node, error) {
	return NewNode(config, DefaultDBProvider, logger)
}

//------------------------------------------------------------------------------

// Node is the highest level interface to a full Tendermint node.
// It includes all configuration information and running services.
type Node struct {
	cmn.BaseService

	// config
	config *cfg.Config

	// network
	sw       *p2p.Switch  // p2p connections
	addrBook pex.AddrBook // known peers

}

// NewNode returns a new, ready to go, Tendermint Node.
func NewNode(config *cfg.Config, dbProvider DBProvider, logger log.Logger) (*Node, error) {

	txpoolLogger := logger.With("module", "txpool")
	txpoolReactor := sync.NewPoolReactor()
	txpoolReactor.SetLogger(txpoolLogger)

	blockLogger := logger.With("module", "block")
	blockReactor := sync.NewBlockReactor()
	blockReactor.SetLogger(blockLogger)

	p2pLogger := logger.With("module", "p2p")

	sw := p2p.NewSwitch(config.P2P)
	sw.SetLogger(p2pLogger)

	sw.AddReactor("TXPOOL", txpoolReactor)
	sw.AddReactor("BLOCK", blockReactor)

	addrBook := pex.NewAddrBook(config.P2P.AddrBookFile(), config.P2P.AddrBookStrict)
	addrBook.SetLogger(p2pLogger.With("book", config.P2P.AddrBookFile()))
	if config.P2P.PexReactor {
		// TODO persistent peers ? so we can have their DNS addrs saved
		pexReactor := pex.NewPEXReactor(addrBook,
			&pex.PEXReactorConfig{
				Seeds:          cmn.SplitAndTrim(config.P2P.Seeds, ",", " "),
				SeedMode:       config.P2P.SeedMode,
				PrivatePeerIDs: cmn.SplitAndTrim(config.P2P.PrivatePeerIDs, ",", " ")})
		pexReactor.SetLogger(p2pLogger)
		sw.AddReactor("PEX", pexReactor)
	}
	sw.SetAddrBook(addrBook)

	node := &Node{
		config:   config,
		sw:       sw,
		addrBook: addrBook,
	}
	node.BaseService = *cmn.NewBaseService(logger, "Node", node)
	return node, nil
}

// OnStart starts the Node. It implements cmn.Service.
func (n *Node) OnStart() error {

	protocol, address := cmn.ProtocolAndAddress(n.config.P2P.ListenAddress)
	l := p2p.NewDefaultListener(protocol, address, n.config.P2P.SkipUPNP, n.Logger.With("module", "p2p"))
	n.sw.AddListener(l)

	nodeKey, err := p2p.LoadOrGenNodeKey(n.config.NodeKeyFile())
	if err != nil {
		return err
	}
	n.Logger.Info("P2P Node ID", "ID", nodeKey.ID(), "file", n.config.NodeKeyFile())

	nodeInfo := n.makeNodeInfo(nodeKey.ID())
	n.sw.SetNodeInfo(nodeInfo)
	n.sw.SetNodeKey(nodeKey)

	// Add ourselves to addrbook to prevent dialing ourselves
	n.addrBook.AddOurAddress(nodeInfo.NetAddress())

	// Start the switch (the P2P server).
	err = n.sw.Start()
	if err != nil {
		return err
	}

	// Always connect to persistent peers
	if n.config.P2P.PersistentPeers != "" {
		err = n.sw.DialPeersAsync(n.addrBook, cmn.SplitAndTrim(n.config.P2P.PersistentPeers, ",", " "), true)
		if err != nil {
			return err
		}
	}
	return nil
}

// OnStop stops the Node. It implements cmn.Service.
func (n *Node) OnStop() {
	n.BaseService.OnStop()

	n.Logger.Info("Stopping Node")
	n.sw.Stop()

}

// RunForever waits for an interrupt signal and stops the node.
func (n *Node) RunForever() {
	cmn.TrapSignal(func() {
		n.Stop()
	})
}

// Switch returns the Node's Switch.
func (n *Node) Switch() *p2p.Switch {
	return n.sw
}

func (n *Node) makeNodeInfo(nodeID p2p.ID) p2p.NodeInfo {

	nodeInfo := p2p.NodeInfo{
		ID:       nodeID,
		Network:  n.config.ChainID,
		Version:  version.MainVersion.StringValue,
		Channels: []byte{},
		Moniker:  n.config.Moniker,
		Other: []string{
			cmn.Fmt("amino_version=%v", amino.Version),
			cmn.Fmt("p2p_version=%v", version.P2PVersion.StringValue),
		},
	}

	if n.config.P2P.PexReactor {
		nodeInfo.Channels = append(nodeInfo.Channels, pex.PexChannel)
	}

	if !n.sw.IsListening() {
		return nodeInfo
	}

	p2pListener := n.sw.Listeners()[0]
	p2pHost := p2pListener.ExternalAddress().IP.String()
	p2pPort := p2pListener.ExternalAddress().Port
	nodeInfo.ListenAddr = cmn.Fmt("%v:%v", p2pHost, p2pPort)

	return nodeInfo
}
