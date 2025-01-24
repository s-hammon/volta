package main

import (
	"context"
	"log"

	"github.com/s-hammon/volta/pkg/jsonhl7"
	hc "google.golang.org/api/healthcare/v1"
)

const (
	projectID = "utopian-button-389823"
	location  = "us-central1"
	topicID   = "hl7topic"
	subID     = "volta"
)

// func main() {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	go handleShutdown(cancel)

// 	psClient, err := pubsub.NewClient(ctx, projectID)
// 	if err != nil {
// 		log.Fatalf("error creating psClient: %v", err)
// 	}
// 	defer psClient.Close()

// 	hcService, err := hc.NewService(ctx)
// 	if err != nil {
// 		log.Fatalf("error creating hcClient: %v", err)
// 	}
// 	app := &App{HealthcareService: hcService}

// 	sub, err := setupSubscription(ctx, psClient, topicID, subID)
// 	if err != nil {
// 		log.Fatalf("error setting up subscription: %v", err)
// 	}

// 	log.Println("Listening for messages...")
// 	if err = sub.Receive(ctx, app.handleSubMessage); err != nil && !errors.Is(err, context.Canceled) {
// 		log.Fatalf("error receiving message: %v", err)
// 	}
// }

func main() {
	hcService, err := hc.NewService(context.Background())
	if err != nil {
		log.Fatalf("error creating hcClient: %v", err)
	}

	b, err := getHL7V2Message(hcService, "projects/utopian-button-389823/locations/us-central1/datasets/ib_methodist/hl7V2Stores/hl7/messages/5__kiFznlHKrEFQiGijC14ap2zwu2a-DU4ECmbY5HtQ=")
	if err != nil {
		log.Fatalf("error getting HL7v2 message: %v", err)
	}

	message, err := jsonhl7.NewMessage(b)
	if err != nil {
		log.Fatalf("error creating message JSON: %v", err)
	}

	messageID := "GvBdYxJb"
	if err := saveMessage(messageID+".json", message.Map()); err != nil {
		log.Fatalf("error saving raw message: %v", err)
	}

	var msg interface{}
	switch message.Type {
	case "ORM":
		msg, err = NewORM(message.Map())
		if err != nil {
			log.Fatalf("error creating ORM: %v", err)
		}
	case "ADT":
		msg, err = NewADT(message.Map())
		if err != nil {
			log.Fatalf("error creating ADT: %v", err)
		}
	case "ORU":
		log.Println("ORUs not yet supported")
	default:
		log.Fatalf("unrecognized message type: %s\t (rejected)", message.Type)
	}

	filename := "message-" + messageID + ".json"
	if err := saveMessage(filename, msg); err != nil {
		log.Fatalf("error saving message: %v", err)
	}
}
