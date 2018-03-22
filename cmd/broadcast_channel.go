package cmd

import (
	"github.com/buoyantio/bb/strategies"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var broadcastChannelCmd = &cobra.Command{
	Use:     strategies.BroadcastChannelStrategyName,
	Short:   "Forwards the request to all downstream services.",
	Example: "bb broadcast-channel --h1-downstream-server http://localhost:9090 --grpc-downstream-server localhost:9091 --h1-server-port 9092",

	Run: func(cmd *cobra.Command, args []string) {
		svc, err := newService(config, strategies.BroadcastChannelStrategyName)
		if err != nil {
			log.Fatalln(err)
		}
		defer svc.Close()
	},
}

func init() {
	RootCmd.AddCommand(broadcastChannelCmd)
}
