package hcapi

import (
	"context"
	"fmt"

	"google.golang.org/api/healthcare/v1"
)

type Client struct {
	svc *healthcare.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := healthcare.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("healthcare.NewService: %v", err)
	}

	return &Client{svc: svc}, nil
}

func (c *Client) GetHl7v2Message(messagePath string) (Message, error) {
	msgSvc := c.svc.Projects.Locations.Datasets.Hl7V2Stores.Messages
	msg, err := msgSvc.Get(messagePath).View("RAW_ONLY").Do()
	if err != nil {
		return Message{}, fmt.Errorf("Messages.Get: %v", err)
	}

	return newMessage(msg), nil
}

type Message struct {
	Data        string
	MessageType string
	// TODO: do as time.Time
	SendTime   string
	CreateTime string
}

func newMessage(msg *healthcare.Message) Message {
	return Message{
		Data:        msg.Data,
		MessageType: msg.MessageType,
		SendTime:    msg.SendTime,
		CreateTime:  msg.CreateTime,
	}
}
