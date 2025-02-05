package api

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"slices"

	json "github.com/json-iterator/go"
	"github.com/s-hammon/volta/pkg/hl7"
	"google.golang.org/api/healthcare/v1"
	"google.golang.org/api/option"
)

type pubSubMessage struct {
	Message      message `json:"message"`
	Subscription string  `json:"subscription"`
}

type message struct {
	Data       []byte     `json:"data,omitempty"`
	Attributes attributes `json:"attributes,omitempty"`
}

type attributes struct {
	Type string `json:"type"`
}

func NewPubSubMessage(body io.Reader) (*pubSubMessage, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var m pubSubMessage
	if err := json.Unmarshal(data, &m); err != nil {
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

func (h *Hl7Client) GetHL7V2Message(messagePath string) (hl7.Message, error) {
	messagesSvc := h.Projects.Locations.Datasets.Hl7V2Stores.Messages
	msg, err := messagesSvc.Get(messagePath).View("RAW_ONLY").Do()
	if err != nil {
		return nil, fmt.Errorf("error getting HL7 message: %w", err)
	}

	raw, err := base64.StdEncoding.DecodeString(msg.Data)
	if err != nil {
		return nil, fmt.Errorf("error decoding HL7 message: %w", err)
	}

	msgMap, err := hl7.NewMessage(raw, byte(SegDelim))
	if err != nil {
		return nil, fmt.Errorf("error parsing HL7 message: %w", err)
	}

	return msgMap, nil
}
