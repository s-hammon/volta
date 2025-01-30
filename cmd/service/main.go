package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-hammon/volta/internal/api"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/models"
	"github.com/s-hammon/volta/pkg/hl7"
)

var (
	dbURL string
	port  string

	db *pgxpool.Pool
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

	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	ctx, cancel = context.WithCancel(context.Background())

	var err error
	db, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("error connecting to DB: %v", err)
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

	http.HandleFunc("/", handleMessage)

	wg.Add(1)
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
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

type pubSubMessage struct {
	Message struct {
		Data       []byte     `json:"data,omitempty"`
		Attributes attributes `json:"attributes,omitempty"`
	} `json:"message"`
	Subscription string `json:"subscription"`
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

type attributes struct {
	Type string `json:"type"`
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Empty Request Body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// TODO: make sure we're sending out Content-Length
	size := r.Header.Get("Content-Length")
	if size == "" {
		log.Println("Content-Length not provided")
	} else {
		log.Printf("received message of size %s bytes\n", size)
	}

	m, err := NewPubSubMessage(r.Body)
	if err != nil {
		log.Printf("error parsing message: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	raw, err := a.Svc.GetHL7V2Message(string(m.Message.Data))
	if err != nil {
		log.Printf("error getting HL7 message: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	msgMap, err := hl7.NewMessage(raw, []byte(api.SegDelim))
	if err != nil {
		log.Printf("error creating message from raw: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	switch m.Message.Attributes.Type {
	case "ORM":
		orm, err := models.NewORM(msgMap)
		if err != nil {
			log.Printf("error creating ORM: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		resp, err := orm.ToDB(ctx, a.DB)
		if err != nil {
			log.Printf("error processing ORM: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		log.Printf("%s:\n%s", resp.Message, resp.Entities)

		w.WriteHeader(http.StatusCreated)
	case "ORU":
		oru, err := models.NewORU(msgMap)
		if err != nil {
			log.Printf("error creating ORU: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		resp, err := oru.ToDB(ctx, a.DB)
		if err != nil {
			log.Printf("error processing ORU: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		log.Printf("%s:\n%s", resp.Message, resp.Entities)

		w.WriteHeader(http.StatusCreated)

	default:
		log.Printf("unknown message type: %s", m.Message.Attributes.Type)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}
