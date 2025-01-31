package api

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/s-hammon/volta/internal/api/models"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/pkg/hl7"
)

const SegDelim = "\r"

type HealthcareClient interface {
	GetHL7V2Message(messagePath string) ([]byte, error)
}

type API struct {
	DB     *database.Queries
	Client HealthcareClient
}

func New(db *database.Queries, client HealthcareClient) *http.ServeMux {
	a := &API{
		DB:     db,
		Client: client,
	}

	mux := http.NewServeMux()

	// TODO: possiby add a health check?

	mux.HandleFunc("POST /", a.handleMessage)

	return mux
}

func (a *API) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		log.Printf("error reading request: %v", errors.New("empty request body"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	size := r.Header.Get("Content-Length")
	if size != "" {
		log.Printf("received message of size %s bytes\n", size)
	} else {
		log.Println("received message")
	}

	m, err := NewPubSubMessage(r.Body)
	if err != nil {
		log.Printf("error parsing request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	raw, err := a.Client.GetHL7V2Message(string(m.Message.Data))
	if err != nil {
		log.Printf("error getting HL7 message: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("got HL7 message: %s", raw)

	msgMap, err := hl7.NewMessage(raw, []byte(SegDelim))
	if err != nil {
		log.Printf("error creating message from raw: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch m.Message.Attributes.Type {
	case "ORM":
		orm, err := models.NewORM(msgMap)
		if err != nil {
			log.Printf("error creating ORM: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = orm.ToDB(context.Background(), a.DB)
		if err != nil {
			log.Printf("error processing ORM: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// TOOD: advanced logging
		log.Printf("ORM processed successfully")
	case "ORU":
		oru, err := models.NewORU(msgMap)
		if err != nil {
			log.Printf("error creating ORU: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = oru.ToDB(context.Background(), a.DB)
		if err != nil {
			log.Printf("error processing ORU: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("ORU processed successfully")
	case "ADT":
		log.Printf("ADT message type not implemented yet")
		w.WriteHeader(http.StatusNotImplemented)
		return
	default:
		log.Printf("unknown message type: %s", m.Message.Attributes.Type)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
