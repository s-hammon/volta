package hcapi

import (
	"context"
	"fmt"
	"log"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/s-hammon/p"
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

func (c *Client) ListHl7v2Messages(storeId string) ([]Message, error) {
	parent := p.Format("projects/silver-pact-448614-t7/locations/us-central1/datasets/strg/hl7V2Stores/%s", storeId)
	resp, err := listResponse(c, parent, "")
	if err != nil {
		return nil, fmt.Errorf("Messages.List: %v", err)
	}

	ret := make([]Message, len(resp.Hl7V2Messages))
	for i, msg := range resp.Hl7V2Messages {
		ret[i] = newMessage(msg)
	}

	return ret, nil
}

func listResponse(c *Client, parent, pageToken string) (*healthcare.ListMessagesResponse, error) {
	msgSvc := c.svc.Projects.Locations.Datasets.Hl7V2Stores.Messages
	return msgSvc.List(parent).PageToken(pageToken).View("RAW_ONLY").Filter("sendDate=\"2025-07-01\" AND messageType=ORM").Do()
}

func (c *Client) GetPubSubTopics(storeId string) ([]string, error) {
	storeSvc := c.svc.Projects.Locations.Datasets.Hl7V2Stores
	parent := p.Format("projects/silver-pact-448614-t7/locations/us-central1/datasets/strg/hl7V2Stores/%s", storeId)

	store, err := storeSvc.Get(parent).Do()
	if err != nil {
		return nil, fmt.Errorf("Stores.Get: %v", err)
	}

	ret := make([]string, len(store.NotificationConfigs))
	for i, cfg := range store.NotificationConfigs {
		ret[i] = cfg.PubsubTopic
	}

	return ret, nil
}

func (c *Client) ReplayMessages(storeId, pageToken string) (int, error) {
	ctx := context.Background()
	psClient, err := pubsub.NewClient(ctx, "silver-pact-448614-t7")
	if err != nil {
		return 0, fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer psClient.Close()

	parent := p.Format("projects/silver-pact-448614-t7/locations/us-central1/datasets/strg/hl7V2Stores/%s", storeId)


	topic := psClient.Topic("methodist")
	defer topic.Stop()
	sent, pages := 0, 0
	nextPageToken := pageToken

	var wg sync.WaitGroup

	for range 10 {
		resp, err := listResponse(c, parent, nextPageToken)
		if err != nil {
			return 0, err
		}

		for _, message := range resp.Hl7V2Messages {
			wg.Go(func() {
				result := topic.Publish(ctx, &pubsub.Message{
					Data: []byte(message.Name),
					Attributes: map[string]string{
						"msgType": message.MessageType,
					},
				})

				_, err := result.Get(ctx)
				if err != nil {
					log.Printf("failed to send notification: %v\n", err)
				} else {
					sent++
				}
			})
		}

		pages++
		if pages % 5 == 0 {
			log.Printf("processed %d pages so far...", pages)
		}
		nextPageToken = resp.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	wg.Wait()
	if nextPageToken != "" {
		log.Printf("next page token: %q\n", nextPageToken)
	}
	return sent, nil
}

func (c *Client) ReplayMessage(messagePath string) (string, error) {
	msg, err := c.GetHl7v2Message(messagePath)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	psClient, err := pubsub.NewClient(ctx, "silver-pact-448614-t7")
	if err != nil {
		return "", fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer psClient.Close()


	topic := psClient.Topic("methodist")
	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte(messagePath),
		Attributes: map[string]string{
			"msgType": msg.MessageType,
		},
	})

	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("Publish.Get: %v", err)
	}

	return id, nil
}

type Message struct {
	Name string
	Data        string
	MessageType string
	// TODO: do as time.Time
	SendTime   string
	CreateTime string
}

func newMessage(msg *healthcare.Message) Message {
	return Message{
		Name: msg.Name,
		Data:        msg.Data,
		MessageType: msg.MessageType,
		SendTime:    msg.SendTime,
		CreateTime:  msg.CreateTime,
	}
}
