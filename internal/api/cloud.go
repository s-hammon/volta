package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"

	"google.golang.org/api/healthcare/v1"
	"google.golang.org/api/option"
)

type pubSubMessage struct {
	Message struct {
		Data       []byte     `json:"data,omitempty"`
		Attributes attributes `json:"attributes,omitempty"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

type attributes struct {
	Type string `json:"type"`
}

func NewPubSubMessage(body io.Reader) (*pubSubMessage, error) {
	var m pubSubMessage
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&m); err != nil {
		return nil, err
	}
	if m.Message.Data == nil {
		return nil, errors.New("empty message data")
	}
	if !slices.Contains([]string{"ORM", "ORU", "ADT"}, m.Message.Attributes.Type) {
		return nil, errors.New("unknown message type")
	}

	return &m, nil
}

type Hl7Client healthcare.Service

func NewHl7Client(ctx context.Context, opts ...option.ClientOption) (*Hl7Client, error) {
	client, err := healthcare.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating HL7 client: %w", err)
	}

	return (*Hl7Client)(client), nil
}

func (h *Hl7Client) GetHL7V2Message(messagePath string) ([]byte, error) {
	messagesSvc := h.Projects.Locations.Datasets.Hl7V2Stores.Messages
	msg, err := messagesSvc.Get(messagePath).Do()
	if err != nil {
		return nil, fmt.Errorf("error getting HL7 message: %w", err)
	}

	return base64.StdEncoding.DecodeString(msg.Data)
}
