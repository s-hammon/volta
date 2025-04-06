package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

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
	defer func() {
		if err := r.Body.Close(); err != nil {
			logMsg.RespondJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
	}()
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

	msg, code, err := HandleByMsgType(a.DB, m.Message.Attributes.Type, msgMap)
	if err != nil {
		logMsg.RespondJSON(w, code, err.Error())
		return
	}

	logMsg.RespondJSON(w, code, msg)
}

func HandleByMsgType(db Repository, msgType string, msgMap hl7.Message) (msg string, code int, err error) {
	switch msgType {
	case "ORM":
		orm := models.ORM{}
		if err = json.Unmarshal(msgMap, &orm); err != nil {
			return "error unmarshaling HL7", http.StatusInternalServerError, err
		}
		if err = db.UpsertORM(context.Background(), orm); err != nil {
			return "error writing to database", http.StatusInternalServerError, err
		}
		msg = "ORM message processed"
	case "ORU":
		oru := models.ORU{}
		if err = json.Unmarshal(msgMap, &oru); err != nil {
			return "error unmarshaling HL7", http.StatusInternalServerError, err
		}
		if err = db.InsertORU(context.Background(), oru); err != nil {
			return "error writing to database", http.StatusInternalServerError, err
		}
		msg = "ORU message processed"
	case "ADT":
		err = fmt.Errorf("ADT message type not implemented")
		return "ADT message type not implemented", http.StatusNotImplemented, err
	default:
		err = fmt.Errorf("unsupported message type")
		return "unsupported message type", http.StatusInternalServerError, err
	}

	return msg, http.StatusCreated, nil
}
