package cmd

import (
	"fmt"

	"github.com/s-hammon/volta/internal/hcapi"
	"github.com/spf13/cobra"
)

var cmdReplay = &cobra.Command{
	Use: "replay [msgPath]",
	Short: "replay messages from store ID to Pub/Sub topic",
	Args: cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		client, err = hcapi.NewClient(cmd.Context())
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "commencing replay...")
		id, err := client.ReplayMessage(args[0])
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "id: %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cmdReplay)
}
