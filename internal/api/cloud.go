package api

import (
	"context"
	"encoding/base64"
	"fmt"

	hc "google.golang.org/api/healthcare/v1"
)

type hcService hc.Service

func NewHCService(ctx context.Context) (*hcService, error) {
	svc, err := hc.NewService(ctx)
	if err != nil {
		return nil, err
	}

	return (*hcService)(svc), nil
}

func (h *hcService) GetHL7V2Message(messagePath string) ([]byte, error) {
	messageService := h.Projects.Locations.Datasets.Hl7V2Stores.Messages
	message, err := messageService.Get(messagePath).Do()
	if err != nil {
		return nil, fmt.Errorf("messageService.Get: got error trying to fetch message %s (%w)", messagePath, err)
	}

	return base64.StdEncoding.DecodeString(message.Data)
}
