package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"cloud.google.com/go/pubsub"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/mattn/go-sqlite3"

	"github.com/s-hammon/volta/internal/api"
	"github.com/s-hammon/volta/internal/database"
)

var (
	dbURL     string
	projectID string
	ormTopic  string
	ormSub    string

	db *pgxpool.Pool
	ps *pubsub.Client
	a  *api.API

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
)

func init() {
	dbURL = os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	projectID = os.Getenv("PROJECT_ID")
	ormTopic = os.Getenv("ORM_TOPIC")
	ormSub = os.Getenv("ORM_SUB")

	ctx, cancel = context.WithCancel(context.Background())

	var err error
	db, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("error connecting to DB: %v", err)
	}

	ps, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("error creating pubsub client: %v", err)
	}

	svc, err := api.NewHCService(ctx)
	if err != nil {
		log.Fatalf("error creating HC service: %v", err)
	}

	a = &api.API{
		DB:  database.New(db),
		Svc: svc,
	}
}

func main() {
	defer cleanup()

	go handleShutdown()

	wg.Add(1)
	go startSubscriber(ctx, ps, ormSub, ormTopic, a.HandleORM, &wg)

	wg.Wait()
}

func startSubscriber(ctx context.Context, ps *pubsub.Client, subName, topic string, handler api.Handler, wg *sync.WaitGroup) {
	defer wg.Done()
	sub, err := getOrCreateSub(ctx, ps, topic, subName)
	if err != nil {
		log.Fatalf("error getting or creating subscription: %v", err)
	}

	log.Printf("starting subscriber: %s", sub.String())
	err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		log.Printf("received message for ORM: %s", string(m.Data))
		resp, err := handler(ctx, m)
		if err != nil {
			log.Printf("error handling message: %v", err)
			m.Nack()
		}

		b, err := json.Marshal(resp)
		if err != nil {
			log.Printf("error marshalling response: %v", err)
			m.Nack()
		}

		log.Printf("message processed: %s", b)
		m.Ack()
	})

	if err != nil {
		log.Fatalf("error receiving message: %v", err)
	}
}

func cleanup() {
	log.Println("shutting down services...")

	if db != nil {
		db.Close()
	}

	if ps != nil {
		ps.Close()
	}

	cancel()
}

func handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("received shutdown signal, exiting...")

	cleanup()
	os.Exit(0)
}

func getOrCreateSub(ctx context.Context, ps *pubsub.Client, topic, sub string) (*pubsub.Subscription, error) {
	t := ps.Topic(topic)
	if _, err := t.Exists(ctx); err != nil {
		return nil, err
	}

	s := ps.Subscription(sub)
	ok, err := s.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if !ok {
		s, err = ps.CreateSubscription(ctx, sub, pubsub.SubscriptionConfig{
			Topic: t,
			RetryPolicy: &pubsub.RetryPolicy{
				MinimumBackoff: 10,
				MaximumBackoff: 60,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}
