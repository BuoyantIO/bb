package cmd

import (
	"github.com/buoyantio/conduit-test/strategies"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var responseText string

var terminusCmd = &cobra.Command{
	Use:   strategies.TerminusStrategyName,
	Short: "Receives the request and returns a pre-defined response",
	Run: func(cmd *cobra.Command, args []string) {
		config.ExtraArguments[strategies.TerminusResponseTextArgName] = responseText
		svc, err := NewService(config, strategies.TerminusStrategyName)

		if err != nil {
			log.Fatalln(err)
		}
		defer svc.Close()
	},
}

func init() {
	RootCmd.AddCommand(terminusCmd)
	terminusCmd.PersistentFlags().StringVar(&responseText, strategies.TerminusResponseTextArgName, "", "Message that this terminus will return")
}
