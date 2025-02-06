package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	json "github.com/json-iterator/go"
	"github.com/s-hammon/volta/internal/api/models"
	"github.com/s-hammon/volta/pkg/hl7"
)

const SegDelim = '\r'

type HealthcareClient interface {
	GetHL7V2Message(messagePath string) (hl7.Message, error)
}

type Repository interface {
	UpsertORM(ctx context.Context, orm models.ORM) error
	InsertORU(ctx context.Context, oru models.ORU) error
}

type API struct {
	DB     Repository
	Client HealthcareClient
}

func New(db Repository, client HealthcareClient) http.Handler {
	a := &API{
		DB:     db,
		Client: client,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /", a.handleMessage)
	mux.HandleFunc("GET /healthz", handleReadiness)

	loggedMux := middlwareLogging(mux)

	return loggedMux
}

func (a *API) handleMessage(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var logMsg logMsg

	if r.Body == nil {
		logMsg.Result = "empty request body"
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	logMsg.NotifSize = r.Header.Get("Content-Length")

	m, err := NewPubSubMessage(r.Body)
	if err != nil {
		logMsg.Result = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	msgMap, err := a.Client.GetHL7V2Message(string(m.Message.Data))
	if err != nil {
		logMsg.Result = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logMsg.Hl7Size = strconv.Itoa(len(msgMap))

	switch m.Message.Attributes.Type {
	case "ORM":
		orm := models.ORM{}
		if err = hl7.Unmarshal(msgMap, &orm); err != nil {
			logMsg.Result = err.Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = a.DB.UpsertORM(context.Background(), orm); err != nil {
			logMsg.Result = err.Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logMsg.Result = "ORM processed successfully"

	case "ORU":
		oru := models.ORU{}
		if err = hl7.Unmarshal(msgMap, &oru); err != nil {
			logMsg.Result = err.Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = a.DB.InsertORU(context.Background(), oru); err != nil {
			logMsg.Result = err.Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logMsg.Result = "ORU processed successfully"
	case "ADT":
		log.Warn().
			Str("messagePath", string(m.Message.Data)).
			Msg("ADT message type not implemented")
		w.WriteHeader(http.StatusNotImplemented)
		return
	default:
		logMsg.Result = "unsupported message type"
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logMsg.Elapsed = time.Since(start).Seconds() * 1000 // milliseconds

	logBytes, err := json.Marshal(logMsg)
	if err != nil {
		log.Error().Err(err).Msg("could not marshal log message")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(logBytes)
}
