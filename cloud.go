package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/pubsub"

	hc "google.golang.org/api/healthcare/v1"
)

func setupSubscription(ctx context.Context, client *pubsub.Client, topicID, subID string) (*pubsub.Subscription, error) {
	topic := client.Topic(topicID)

	sub := client.Subscription(subID)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("sub.Exists: %v", err)
	}

	if !exists {
		sub, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
			Topic: topic,
		})
		if err != nil {
			return nil, fmt.Errorf("client.CreateSubscription: %v", err)
		}
	}

	return sub, nil
}

func getHL7V2Message(hcService *hc.Service, messagePath string) ([]byte, error) {
	messageService := hcService.Projects.Locations.Datasets.Hl7V2Stores.Messages
	message, err := messageService.Get(messagePath).Do()
	if err != nil {
		return nil, fmt.Errorf("messageService.Get: %w", err)
	}

	return base64.StdEncoding.DecodeString(message.Data)
}

func handleShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Println("Received shutdown signal, exiting...")
	cancel()
}
