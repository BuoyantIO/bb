package cmd

import (
	"github.com/buoyantio/conduit-test/building_blocks/strategies"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var responseText string

// terminusCmd represents the terminus command
var terminusCmd = &cobra.Command{
	Use:   "terminus",
	Short: "Receives the request and returns a response",
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
