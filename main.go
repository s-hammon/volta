package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/s-hammon/volta/pkg/hl7"
	hc "google.golang.org/api/healthcare/v1"
)

func main() {
	hcService, err := hc.NewService(context.Background())
	if err != nil {
		log.Fatalf("error creating hcClient: %v", err)
	}

	// JANK
	message, err := hl7.FromJSON("GvBdYxJb.json")
	if err != nil {
		// get message path from arg
		if len(os.Args) < 2 {
			log.Fatalf("usage: %s <path to message>", os.Args[0])
		}
		msgPath := os.Args[1]
		b, err := getHL7V2Message(hcService, msgPath)
		if err != nil {
			log.Fatalf("error getting HL7v2 message: %v", err)
		}

		message, err = hl7.NewMessage(b, []byte("\r"))
		if err != nil {
			log.Fatalf("error creating message JSON: %v", err)
		}
	}

	messageID := "GvBdYxJb"
	if err := saveMessage(messageID+".json", message); err != nil {
		log.Fatalf("error saving raw message: %v", err)
	}

	var msg interface{}
	msgType := message.Type()
	switch msgType {
	case "ORM":
		msg, err = NewORM(message)
		if err != nil {
			log.Fatalf("error creating ORM: %v", err)
		}
		orm := msg.(ORM)
		orm.Materialize()
	case "ADT":
		msg, err = NewADT(message)
		if err != nil {
			log.Fatalf("error creating ADT: %v", err)
		}
	case "ORU":
		log.Println("ORUs not yet supported")
	default:
		log.Fatalf("unrecognized message type: %s\t (rejected)", msgType)
	}

	filename := "message-" + messageID + ".json"
	if err := saveMessage(filename, msg); err != nil {
		log.Fatalf("error saving message: %v", err)
	}
}

func saveMessage(name string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(name, jsonData, 0644); err != nil {
		return err
	}

	return nil
}
