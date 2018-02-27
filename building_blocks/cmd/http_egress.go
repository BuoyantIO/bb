package cmd

import (
	"github.com/buoyantio/conduit-test/building_blocks/strategies"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var urlToInvoke string

var httpEgressCmd = &cobra.Command{
	Use:   strategies.HttpEgressStrategyName,
	Short: "Receives a request, makes a HTTP(S) call to a specified URL and return the body of the response",
	Run: func(cmd *cobra.Command, args []string) {
		config.ExtraArguments[strategies.HttpEgressUrlToInvokeArgName] = urlToInvoke
		svc, err := NewService(config, strategies.HttpEgressStrategyName)

		if err != nil {
			log.Fatalln(err)
		}
		defer svc.Close()
	},
}

func init() {
	RootCmd.AddCommand(httpEgressCmd)
	httpEgressCmd.PersistentFlags().StringVar(&urlToInvoke, strategies.HttpEgressUrlToInvokeArgName, "", "HTTP(S) URL to make a request to")
}
