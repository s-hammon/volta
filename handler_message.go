package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	"cloud.google.com/go/pubsub"
	hc "google.golang.org/api/healthcare/v1"

	"github.com/s-hammon/volta/pkg/jsonhl7"
)

type App struct {
	HealthcareService *hc.Service
}

func (a *App) handleSubMessage(ctx context.Context, m *pubsub.Message) {
	log.Printf("Got message: %s", m.Data)
	data := string(m.Data)

	b, err := getHL7V2Message(a.HealthcareService, data)
	if err != nil {
		log.Printf("[WARNING] error getting HL7v2 message: %v", err)
		m.Nack()
		return
	}

	message, err := jsonhl7.NewMessage(b)
	if err != nil {
		log.Printf("[WARNING] error creating message JSON: %v", err)
		m.Nack()
		return
	}

	switch message.Type {
	case "ORM":
		msg, err := NewORM(message.Map())
		if err != nil {
			log.Printf("[WARNING] error creating ORM: %v", err)
			m.Nack()
			return
		}
		log.Printf("ORM: %+v", msg)
	case "ADT":
		msg, err := NewADT(message.Map())
		if err != nil {
			log.Printf("[WARNING] error creating ADT: %v", err)
			m.Nack()
			return
		}
		log.Printf("ADT: %+v", msg)
	case "ORU":
		log.Printf("ORUs not yet supported")
	default:
		log.Printf("[WARNING] unrecognized message type: %s\t (rejected)", message.Type)
		m.Nack()
		return
	}

	messageID := path.Base(data)
	if len(messageID) >= 7 {
		messageID = messageID[:7]
	}
	filename := fmt.Sprintf("message-%s.json", messageID)

	if err := saveMessage(filename, message.Map()); err != nil {
		log.Printf("[WARNING] error saving message: %v", err)
		m.Nack()
		return
	}

	m.Ack()
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
