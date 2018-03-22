package cmd

import (
	"github.com/buoyantio/bb/strategies"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var urlToInvoke string
var methodToUse string
var clientTimeout string

var httpEgressCmd = &cobra.Command{
	Use:     strategies.HTTPEgressStrategyName,
	Short:   "Receives a request, makes a HTTP(S) call to a specified URL and return the body of the response",
	Example: "bb http-egress --h1-server-port 8080 --method GET --url http://httpbin.org/anything",
	Run: func(cmd *cobra.Command, args []string) {
		config.ExtraArguments[strategies.HTTPEgressURLToInvokeArgName] = urlToInvoke
		config.ExtraArguments[strategies.HTTPEgressHTTPMethodToUseArgName] = methodToUse
		config.ExtraArguments[strategies.HTTPEgressHTTPTimeoutArgName] = clientTimeout
		svc, err := newService(config, strategies.HTTPEgressStrategyName)

		if err != nil {
			log.Fatalln(err)
		}
		defer svc.Close()
	},
}

func init() {
	RootCmd.AddCommand(httpEgressCmd)
	httpEgressCmd.PersistentFlags().StringVar(&urlToInvoke, strategies.HTTPEgressURLToInvokeArgName, "", "HTTP(S) URL to make a request to")
	httpEgressCmd.PersistentFlags().StringVar(&methodToUse, strategies.HTTPEgressHTTPMethodToUseArgName, "GET", "HTTP method to use in request, can be GET, POST, PUT, DELETE, or PATCH")
	httpEgressCmd.PersistentFlags().StringVar(&clientTimeout, strategies.HTTPEgressHTTPTimeoutArgName, "10s", "Timeout for the HTTP client used, must be valid as per time.ParseDuration()")
}
