package cmd

import (
	"github.com/buoyantio/conduit-test/strategies"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var pointToPointChannelCmd = &cobra.Command{
	Use:   strategies.PointToPointStrategyName,
	Short: "Forwards the request to one and only one downstream service.",

	Run: func(cmd *cobra.Command, args []string) {
		svc, err := NewService(config, strategies.PointToPointStrategyName)
		if err != nil {
			log.Fatalln(err)
		}
		defer svc.Close()
	},
}

func init() {
	RootCmd.AddCommand(pointToPointChannelCmd)
}
