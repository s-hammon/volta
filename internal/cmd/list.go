package cmd

import (
	"fmt"

	"github.com/s-hammon/volta/internal/hcapi"
	"github.com/spf13/cobra"
)

var cmdList = &cobra.Command{
	Use: "list [storeId]",
	Short: "list messages from the provided store ID",
	Args: cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		client, err = hcapi.NewClient(cmd.Context())
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		messages, err := client.ListHl7v2Messages(args[0])
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "found %d messages\n", len(messages))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cmdList)
}
