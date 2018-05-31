package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// NOTE: Most of the structs & relevant comments + the
// default configuration options were used to manually
// generate the config.toml. Please reflect any changes
// made here in the defaultConfigTemplate constant in
// config/toml.go
// NOTE: tmlibs/cli must know to look in the config dir!
var (
	defaultConfigDir = "config"
	defaultDataDir   = "data"

	defaultConfigFileName = "config.toml"

	defaultChainID = "testnet"

	defaultNodeKeyName  = "node_key.json"
	defaultAddrBookName = "addrbook.json"

	defaultConfigFilePath = filepath.Join(defaultConfigDir, defaultConfigFileName)
	defaultNodeKeyPath    = filepath.Join(defaultConfigDir, defaultNodeKeyName)
	defaultAddrBookPath   = filepath.Join(defaultConfigDir, defaultAddrBookName)
)

// Config defines the top level configuration for a Tendermint node
type Config struct {
	// Top level options use an anonymous struct
	BaseConfig `mapstructure:",squash"`

	P2P *P2PConfig `mapstructure:"p2p"`
}

// DefaultConfig returns a default configuration for a Tendermint node
func DefaultConfig() *Config {
	return &Config{
		BaseConfig: DefaultBaseConfig(),
		P2P:        DefaultP2PConfig(),
	}
}

// SetRoot sets the RootDir for all Config structs
func (cfg *Config) SetRoot(root string) *Config {
	cfg.BaseConfig.RootDir = root
	cfg.P2P.RootDir = root
	return cfg
}

//-----------------------------------------------------------------------------
// BaseConfig

// BaseConfig defines the base configuration for a Tendermint node
type BaseConfig struct {

	// ChainID is unexposed and immutable but here for convenience
	ChainID string

	// The root directory for all data.
	// This should be set in viper so it can unmarshal into this struct
	RootDir string `mapstructure:"home"`

	// A JSON file containing the private key to use for p2p authenticated encryption
	NodeKey string `mapstructure:"node_key_file"`

	// A custom human readable name for this node
	Moniker string `mapstructure:"moniker"`

	// Output level for logging
	LogLevel string `mapstructure:"log_level"`

	// Database backend: leveldb | memdb
	DBBackend string `mapstructure:"db_backend"`

	// Database directory
	DBPath string `mapstructure:"db_dir"`
}

// DefaultBaseConfig returns a default base configuration for a Tendermint node
func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		ChainID:   defaultChainID,
		NodeKey:   defaultNodeKeyPath,
		Moniker:   defaultMoniker,
		LogLevel:  DefaultPackageLogLevels(),
		DBBackend: "leveldb",
		DBPath:    "data",
	}
}

// P2PConfig defines the configuration options for the Tendermint peer-to-peer networking layer
type P2PConfig struct {
	RootDir string `mapstructure:"home"`

	// Address to listen for incoming connections
	ListenAddress string `mapstructure:"laddr"`

	// Comma separated list of seed nodes to connect to
	// We only use these if we can’t connect to peers in the addrbook
	Seeds string `mapstructure:"seeds"`

	// Comma separated list of nodes to keep persistent connections to
	// Do not add private peers to this list if you don't want them advertised
	PersistentPeers string `mapstructure:"persistent_peers"`

	// Skip UPNP port forwarding
	SkipUPNP bool `mapstructure:"skip_upnp"`

	// Path to address book
	AddrBook string `mapstructure:"addr_book_file"`

	// Set true for strict address routability rules
	AddrBookStrict bool `mapstructure:"addr_book_strict"`

	// Maximum number of peers to connect to
	MaxNumPeers int `mapstructure:"max_num_peers"`

	// Time to wait before flushing messages out on the connection, in ms
	FlushThrottleTimeout int `mapstructure:"flush_throttle_timeout"`

	// Maximum size of a message packet payload, in bytes
	MaxPacketMsgPayloadSize int `mapstructure:"max_packet_msg_payload_size"`

	// Rate at which packets can be sent, in bytes/second
	SendRate int64 `mapstructure:"send_rate"`

	// Rate at which packets can be received, in bytes/second
	RecvRate int64 `mapstructure:"recv_rate"`

	// Set true to enable the peer-exchange reactor
	PexReactor bool `mapstructure:"pex"`

	// Seed mode, in which node constantly crawls the network and looks for
	// peers. If another node asks it for addresses, it responds and disconnects.
	//
	// Does not work if the peer-exchange reactor is disabled.
	SeedMode bool `mapstructure:"seed_mode"`

	// Authenticated encryption
	AuthEnc bool `mapstructure:"auth_enc"`

	// Comma separated list of peer IDs to keep private (will not be gossiped to other peers)
	PrivatePeerIDs string `mapstructure:"private_peer_ids"`
}

// DefaultP2PConfig returns a default configuration for the peer-to-peer layer
func DefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		ListenAddress:           "tcp://0.0.0.0:12345",
		AddrBook:                defaultAddrBookPath,
		AddrBookStrict:          true,
		MaxNumPeers:             50,
		FlushThrottleTimeout:    100,
		MaxPacketMsgPayloadSize: 1024,   // 1 kB
		SendRate:                512000, // 500 kB/s
		RecvRate:                512000, // 500 kB/s
		PexReactor:              true,
		SeedMode:                false,
		AuthEnc:                 true,
	}
}

// NodeKeyFile returns the full path to the node_key.json file
func (cfg BaseConfig) NodeKeyFile() string {
	return rootify(cfg.NodeKey, cfg.RootDir)
}

// DBDir returns the full path to the database directory
func (cfg BaseConfig) DBDir() string {
	return rootify(cfg.DBPath, cfg.RootDir)
}

// DefaultLogLevel returns a default log level of "error"
func DefaultLogLevel() string {
	return "error"
}

// DefaultPackageLogLevels returns a default log level setting so all packages
// log at "error", while the `state` and `main` packages log at "info"
func DefaultPackageLogLevels() string {
	return fmt.Sprintf("main:info,state:info,*:%s", DefaultLogLevel())
}

// AddrBookFile returns the full path to the address book
func (cfg *P2PConfig) AddrBookFile() string {
	return rootify(cfg.AddrBook, cfg.RootDir)
}

// helper function to make config creation independent of root dir
func rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

//-----------------------------------------------------------------------------
// Moniker

var defaultMoniker = getDefaultMoniker()

// getDefaultMoniker returns a default moniker, which is the host name. If runtime
// fails to get the host name, "anonymous" will be returned.
func getDefaultMoniker() string {
	moniker, err := os.Hostname()
	if err != nil {
		moniker = "anonymous"
	}
	return moniker
}

// DefaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home == "" {
		return "./.fusion"
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Fusion")
	case "windows":
		return filepath.Join(home, "AppData", "Roaming", "fusion")
	default:
		return filepath.Join(home, ".fusion")
	}
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
