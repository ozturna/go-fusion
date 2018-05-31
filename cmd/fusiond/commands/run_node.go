package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	nm "github.com/go-fusion/node"
)

func addNodeFlags(cmd *cobra.Command) {

	cmd.Flags().String("p2p.laddr", config.P2P.ListenAddress, "Node listen address. (0.0.0.0:0 means any interface, any port)")
	cmd.Flags().String("p2p.seeds", config.P2P.Seeds, "Comma-delimited ID@host:port seed nodes")
}

func newRunNodeCmd(nodeProvider nm.Provider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Run the fusiond node",
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := nodeProvider(config, logger)
			if err != nil {
				return fmt.Errorf("Failed to create node: %v", err)
			}
			if err := n.Start(); err != nil {
				return fmt.Errorf("Failed to start node: %v", err)
			}
			logger.Info("Started node", "nodeInfo", n.Switch().NodeInfo())
			n.RunForever()
			return nil
		},
	}
	addNodeFlags(cmd)
	return cmd
}

func init() {
	rootCmd.AddCommand(newRunNodeCmd(nm.DefaultNewNode))
}
