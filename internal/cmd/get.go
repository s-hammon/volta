package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/s-hammon/hl7"
	v23 "github.com/s-hammon/hl7/standards/v23"
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
	data, err := base64.StdEncoding.DecodeString(msg.Data)
	if err != nil {
		fmt.Fprintf(w, "couldn't decode data: %v\n", err)
		return
	}

	switch msg.MessageType {
	default:
		printMsg(w, data)
	case "ORM":
		orm := v23.ORM_O01{}
		if err := hl7.Unmarshal(data, &orm); err != nil {
			fmt.Fprintln(w, err.Error())
			printMsg(w, data)
		} else {
			fmt.Fprintf(w, "%v\n", orm)
		}
	case "ORU":
		oru := v23.ORU_R01{}
		if err := hl7.Unmarshal(data, &oru); err != nil {
			fmt.Fprintln(w, err.Error())
		} else {
			fmt.Fprintf(w, "%v\n", oru)
			printMsg(w, data)
		}
	}
}

func printMsg(w io.Writer, data []byte) {
	for d := range bytes.SplitSeq(data, []byte{'\r'}) {
		fmt.Fprintf(w, "%s\n", d)
	}
}
