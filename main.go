package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/s-hammon/volta/pkg/jsonhl7"
)

func main() {
	b, err := os.ReadFile("message.json")
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	response := Response{}
	if err := json.Unmarshal(b, &response); err != nil {
		log.Fatalf("error unmarshalling json: %v", err)
	}

	message, err := jsonhl7.NewMessage(response.Data)
	if err != nil {
		log.Fatalf("error creating message: %v", err)
	}

	log.Printf("%+v", message)
}
