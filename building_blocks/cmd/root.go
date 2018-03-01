package cmd

import (
	"fmt"
	"os"

	"github.com/buoyantio/conduit-test/building_blocks/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var config = &service.Config{
	ExtraArguments: map[string]string{},
}

var logLevel string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "bb",
	Short: "Various microservices that can be used to build a test lab for Conduit",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// set global log level
		level, err := log.ParseLevel(logLevel)
		if err != nil {
			log.Fatalf("invalid log-level: %s", logLevel)
		}
		log.SetLevel(level)
		if config.Id == "" {
			config.Id = fmt.Sprintf("%s-grpc:%d-h1:%d", cmd.Name(), config.GrpcServerPort, config.H1ServerPort)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&config.Id, "id", "", "identifier for this container")
	RootCmd.PersistentFlags().IntVar(&config.GrpcServerPort, "grpc-server-port", -1, "port to bind a gRPC server to")
	RootCmd.PersistentFlags().IntVar(&config.H1ServerPort, "h1-server-port", -1, "port to bind a HTTP 1.1 server to")
	RootCmd.PersistentFlags().IntVar(&config.PercentageFailedRequests, "percent-failure", 0, "percentage of requests that this service will automatically fail")
	RootCmd.PersistentFlags().IntVar(&config.SleepInMillis, "sleep-in-millis", 0, "amount of milliseconds to wait before actually start processing as request")
	RootCmd.PersistentFlags().BoolVar(&config.FireAndForget, "fire-and-forget", false, "do not wait for a response when contacting downstream services.")
	RootCmd.PersistentFlags().StringSliceVar(&config.GrpcDownstreamServers, "grpc-downstream-server", []string{}, "list of servers (hostname:port) to send messages to using gRPC, can be repeated")
	RootCmd.PersistentFlags().StringSliceVar(&config.H1DownstreamServers, "h1-downstream-server", []string{}, "list of servers (protocol://hostname:port) to send messages to using HTTP 1.1, can be repeated")
	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", log.DebugLevel.String(), "log level, must be one of: panic, fatal, error, warn, info, debug")
}
