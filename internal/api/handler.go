package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/s-hammon/volta/internal/api/models"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/pkg/hl7"
)

const SegDelim = "\r"

type HealthcareClient interface {
	GetHL7V2Message(messagePath string) ([]byte, error)
}

type logMsg struct {
	notifSize string
	hl7Size   string
	result    string
	elapsed   time.Duration
}

func (l *logMsg) Log(result string) {
	l.result = result
	log.Info().
		Str("notifSize", l.notifSize).
		Str("hl7Size", l.hl7Size).
		Str("result", l.result).
		Dur("elapsed", l.elapsed).
		Msg("message processed")
}

func (l *logMsg) Error(err error, sendingFac, ControlID string) {
	l.result = err.Error()
	log.Error().
		Err(err).
		Str("sendingFacility", sendingFac).
		Str("controlID", ControlID).
		Msg("could not process message")
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
	// time this function
	start := time.Now()
	var logMsg logMsg

	if r.Body == nil {
		logMsg.Error(errors.New("empty request body"), "", "")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	size := r.Header.Get("Content-Length")
	if size != "" {
		logMsg.notifSize = size
	}

	m, err := NewPubSubMessage(r.Body)
	if err != nil {
		logMsg.Error(err, "", "")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	raw, err := a.Client.GetHL7V2Message(string(m.Message.Data))
	if err != nil {
		logMsg.Error(err, "", "")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logMsg.hl7Size = strconv.Itoa(len(raw))

	msgMap, err := hl7.NewMessage(raw, []byte(SegDelim))
	if err != nil {
		logMsg.Error(err, "", "")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch m.Message.Attributes.Type {
	case "ORM":
		orm, err := models.NewORM(msgMap)
		if err != nil {
			logMsg.Error(err, orm.MSH.SendingFac, orm.MSH.ControlID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// TODO: get rid of first return value
		_, err = orm.ToDB(context.Background(), a.DB)
		if err != nil {
			logMsg.Error(err, orm.MSH.SendingFac, orm.MSH.ControlID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logMsg.result = "ORM processed successfully"
	case "ORU":
		oru, err := models.NewORU(msgMap)
		if err != nil {
			logMsg.Error(err, oru.MSH.SendingFac, oru.MSH.ControlID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = oru.ToDB(context.Background(), a.DB)
		if err != nil {
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
