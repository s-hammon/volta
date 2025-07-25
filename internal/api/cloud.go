package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"slices"

	json "github.com/json-iterator/go"
	"google.golang.org/api/healthcare/v1"
	"google.golang.org/api/option"
)

type pubSubMessage struct {
	Message      message `json:"message"`
	Subscription string  `json:"subscription"`
}

type message struct {
	Data       []byte     `json:"data,omitempty"`
	Attributes attributes `json:"attributes"`
}

type attributes struct {
	Type string `json:"msgType"`
}

func NewPubSubMessage(body io.Reader) (*pubSubMessage, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}

	var m pubSubMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PubSub message: %v", err)
	}

	if len(m.Message.Data) == 0 {
		return nil, fmt.Errorf("empty message data")
	}
	if !slices.Contains([]string{"ORM", "ORU", "ADT"}, m.Message.Attributes.Type) {
		return nil, fmt.Errorf("invalid message type; %s", m.Message.Attributes.Type)
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
	msg, err := messagesSvc.Get(messagePath).View("RAW_ONLY").Do()
	if err != nil {
		return nil, fmt.Errorf("error getting HL7 message: %w", err)
	}

	return base64.StdEncoding.DecodeString(msg.Data)
}
