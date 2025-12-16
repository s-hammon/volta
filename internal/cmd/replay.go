package cmd

import (
	"fmt"

	"github.com/s-hammon/volta/internal/hcapi"
	"github.com/spf13/cobra"
)

var cmdReplay = &cobra.Command{
	Use: "replay [msgPath] [pageToken]",
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

		token := ""
		if len(args) >= 2 {
			token = args[1]
		}
		sent, err := client.ReplayMessages(args[0], token)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "sent %d notifications\n", sent)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cmdReplay)
}
