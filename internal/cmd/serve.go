package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// TODO: include YAML config file to parse & pass to service.
var (
	host string
	port string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Volta service",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		log.Printf("host: %s\tport: %s\n", host, port)
		// log.Info().Str("host", host).Str("port", port).Msg("service configuration")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	// TODO: init config

	serveCmd.PersistentFlags().StringVarP(&host, "host", "H", "0.0.0.0", "")
	serveCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "port for server to listen on")
}
