package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-hammon/volta/internal/api"
	"github.com/s-hammon/volta/internal/database"
)

var (
	dbURL        string
	domain       string
	port         string
	hl7AuthToken string
	baseURL      string

	db *pgxpool.Pool
	a  *http.ServeMux

	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	timeout time.Duration
)

func init() {
	// TODO: put next 3 as CLI args
	dbURL = os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	host := os.Getenv("SERVICE_HOST")
	if host == "" {
		host = "localhost"
		log.Println("using default host")
	}

	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s\n", port)
	}

	hl7AuthToken = os.Getenv("HL7_AUTH_TOKEN")
	if hl7AuthToken == "" {
		log.Fatal("HL7_AUTH_TOKEN is required")
	}

	baseURL = os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://healthcare.googleapis.com/v1"
		log.Printf("defaulting to base URL %s\n", baseURL)
	}

	timeout = 5 * time.Second

	client, err := api.NewHl7Client(baseURL, hl7AuthToken, timeout)
	if err != nil {
		log.Fatalf("error creating HL7 client: %v", err)
	}

	ctx, cancel = context.WithCancel(context.Background())

	db, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("error connecting to DB: %v", err)
	}

	a = api.New(database.New(db), client)
}

func main() {
	defer cleanup()
	go handleShutdown()

	srv := &http.Server{
		Addr:    net.JoinHostPort(domain, port),
		Handler: a,
	}

	wg.Add(1)
	log.Printf("Listening on port %s", port)
	log.Fatal(srv.ListenAndServe())
	wg.Wait()
}

func cleanup() {
	log.Println("shutting down services...")

	if db != nil {
		db.Close()
	}

	cancel()
}

func handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	<-sigChan
	log.Println("received shutdown signal, shutting down...")

	cleanup()
	os.Exit(0)
}
