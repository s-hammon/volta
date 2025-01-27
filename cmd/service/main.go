package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
		Data       []byte                 `json:"data,omitempty"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	var m pubSubMessage
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("io.ReadAll: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	fmt.Printf("received message of size %d bytes\n", len(body))

	if err := json.Unmarshal(body, &m); err != nil {
		log.Printf("json.Unmarshal: %v", err)
		log.Printf("body: %s", body)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	msgType, ok := m.Message.Attributes["type"]
	if !ok {
		log.Printf("missing message_type attribute")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	msgType, ok = msgType.(string)
	if !ok {
		log.Printf("invalid message_type attribute: expecting string, got %T", msgType)
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

	switch strings.ToUpper(msgType.(string)) {
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

		log.Printf("message processed: %+v", resp)

		w.WriteHeader(http.StatusCreated)
	case "ORU":
		oru, err := models.NewORU(msgMap)
		if err != nil {
			log.Printf("error creating ORU: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		log.Printf("received message ID: %s", oru.MSH.ControlID)
	default:
		log.Printf("unknown message type: %s", msgType)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}
