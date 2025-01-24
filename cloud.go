package main

import (
	"encoding/base64"
	"fmt"

	hc "google.golang.org/api/healthcare/v1"
)

func getHL7V2Message(hcService *hc.Service, messagePath string) ([]byte, error) {
	messageService := hcService.Projects.Locations.Datasets.Hl7V2Stores.Messages
	message, err := messageService.Get(messagePath).Do()
	if err != nil {
		return nil, fmt.Errorf("messageService.Get: %w", err)
	}

	return base64.StdEncoding.DecodeString(message.Data)
}
