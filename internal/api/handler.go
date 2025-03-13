package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"

	json "github.com/json-iterator/go"
	"github.com/s-hammon/volta/internal/api/models"
	"github.com/s-hammon/volta/pkg/hl7"
)

type HealthcareClient interface {
	GetHL7V2Message(string) (hl7.Message, error)
}

type Repository interface {
	UpsertORM(context.Context, models.ORM) error
	InsertORU(context.Context, models.ORU) error
}

type API struct {
	DB        Repository
	Client    HealthcareClient
	debugMode bool
}

func New(db Repository, client HealthcareClient, debugMode bool) http.Handler {
	a := &API{
		DB:        db,
		Client:    client,
		debugMode: debugMode,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /", a.handleMessage)
	mux.HandleFunc("GET /healthz", handleReadiness)

	loggedMux := middlewareLogging(mux)

	return loggedMux
}

func (a *API) handleMessage(w http.ResponseWriter, r *http.Request) {
	logMsg := NewLogMsg()

	if r.Body == nil {
		logMsg.RespondJSON(w, http.StatusBadRequest, "empty request body")
		return
	}
	defer r.Body.Close()
	logMsg.NotifSize = r.Header.Get("Content-Length")

	m, err := NewPubSubMessage(r.Body)
	if err != nil {
		logMsg.RespondJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	msgMap, err := a.Client.GetHL7V2Message(string(m.Message.Data))
	if err != nil {
		logMsg.RespondJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	logMsg.Hl7Size = strconv.Itoa(len(msgMap))

	if a.debugMode {
		logMsg.RespondJSON(w, http.StatusOK, "received message")
		return
	}

	var msg string
	switch m.Message.Attributes.Type {
	case "ORM":
		orm := models.ORM{}
		if err = json.Unmarshal(msgMap, &orm); err != nil {
			logMsg.RespondJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		if err = a.DB.UpsertORM(context.Background(), orm); err != nil {
			logMsg.RespondJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		msg = "ORM processed successfully"

	case "ORU":
		oru := models.ORU{}
		if err = json.Unmarshal(msgMap, &oru); err != nil {
			logMsg.RespondJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		if err = a.DB.InsertORU(context.Background(), oru); err != nil {
			logMsg.RespondJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		msg = "ORU processed successfully"

	case "ADT":
		log.Warn().
			Str("messagePath", string(m.Message.Data)).
			Msg("ADT message type not implemented")
		logMsg.RespondJSON(w, http.StatusNotImplemented, "ADT message type not implemented")
		return
	default:
		logMsg.RespondJSON(w, http.StatusBadRequest, "unsupported message type")
		return
	}

	logMsg.RespondJSON(w, http.StatusCreated, msg)
}
