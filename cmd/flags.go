package cmd

import (
	"github.com/spf13/cobra"
)

//AttachCLIFlags attaches command line flags to command
func AttachCLIFlags(rootCmd *cobra.Command) error {
	rootCmd.PersistentFlags().StringP("config", "c", "", "the config file to use")

	rootCmd.PersistentFlags().StringP("port", "p", "", "Port for api server to run")
	rootCmd.PersistentFlags().StringP("grpcport", "g", "", "Port for grpc server to run")
	rootCmd.PersistentFlags().BoolP("verbose", "", false, "should every proxy request be logged to stdout")

	return nil
}
