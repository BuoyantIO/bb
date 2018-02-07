package cmd

import (
	"github.com/buoyantio/conduit-test/building_blocks/strategies"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var broadcastChannelCmd = &cobra.Command{
	Use:   strategies.BroadcastChannelStrategyName,
	Short: "Forwards the request to all downstream services.",

	Run: func(cmd *cobra.Command, args []string) {
		svc, err := NewService(config, strategies.BroadcastChannelStrategyName)
		if err != nil {
			log.Fatalln(err)
		}
		defer svc.Close()
	},
}

func init() {
	RootCmd.AddCommand(broadcastChannelCmd)
}
