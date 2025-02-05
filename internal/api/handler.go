package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

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

func New(db Repository, client HealthcareClient) *http.ServeMux {
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
	// TODO: reimplement logging (use middleware, perhaps)
	// time this function
	start := time.Now()
	var logMsg logMsg

	if r.Body == nil {
		logMsg.Error(errors.New("empty request body"), "", "")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	logMsg.notifSize = r.Header.Get("Content-Length")

	m, err := NewPubSubMessage(r.Body)
	if err != nil {
		logMsg.Error(err, "", "")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	msgMap, err := a.Client.GetHL7V2Message(string(m.Message.Data))
	if err != nil {
		logMsg.Error(err, "", "")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logMsg.hl7Size = strconv.Itoa(len(msgMap))

	switch m.Message.Attributes.Type {
	case "ORM":
		// TODO: to interface
		orm := models.ORM{}
		if err = hl7.Unmarshal(msgMap, &orm); err != nil {
			logMsg.Error(err, orm.MSH.SendingFac, orm.MSH.ControlID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = a.DB.UpsertORM(context.Background(), orm); err != nil {
			logMsg.Error(err, orm.MSH.SendingFac, orm.MSH.ControlID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logMsg.result = "ORM processed successfully"

	case "ORU":
		// TODO: to interface
		oru := models.ORU{}
		if err = hl7.Unmarshal(msgMap, &oru); err != nil {
			logMsg.Error(err, oru.MSH.SendingFac, oru.MSH.ControlID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = a.DB.InsertORU(context.Background(), oru); err != nil {
			logMsg.Error(err, oru.MSH.SendingFac, oru.MSH.ControlID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logMsg.result = "ORU processed successfully"
	case "ADT":
		log.Warn().
			Str("messagePath", string(m.Message.Data)).
			Msg("ADT message type not implemented")
		w.WriteHeader(http.StatusNotImplemented)
		return
	default:
		logMsg.Error(errors.New("unknown message type"), "", "")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logMsg.elapsed = time.Since(start)
	logMsg.Log(logMsg.result)

	w.WriteHeader(http.StatusCreated)
}
