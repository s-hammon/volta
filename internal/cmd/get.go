package cmd

import (
	"encoding/base64"
	"fmt"
	"io"

	"github.com/s-hammon/volta/internal/hcapi"
	"github.com/spf13/cobra"
)

var client *hcapi.Client

var cmdGet = &cobra.Command{
	// TODO: split path into dataset and store name(?)
	// "get [dataset] [store]"
	Use:   "get [path]",
	Short: "use the client to get a record from the Cloud Healthcare API",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		client, err = hcapi.NewClient(cmd.Context())
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// log.Println("welcome to the client")
		msg, err := client.GetHl7v2Message(args[0])
		if err != nil {
			return err
		}

		prettyPrint(cmd.OutOrStdout(), msg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cmdGet)
}

func prettyPrint(w io.Writer, msg hcapi.Message) {
	fmt.Fprintf(w, "Message Type %q found.\n", msg.MessageType)
	decoded, err := base64.StdEncoding.DecodeString(msg.Data)
	if err != nil {
		fmt.Fprintf(w, "couldn't decode data: %v\n", err)
	} else {
		fmt.Fprintf(w, "%s\n", decoded)
	}
}
