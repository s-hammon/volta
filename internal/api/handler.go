package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/s-hammon/volta/internal/entity"
	"github.com/s-hammon/volta/pkg/hl7"
)

type HealthcareClient interface {
	GetHL7V2Message(string) ([]byte, error)
}

type HL7Store interface {
	SaveORM(context.Context, *entity.Order) error
	SaveORU(context.Context, *entity.Observation) error
	// TODO: already, I can see how this will get out of hand
	// think of another design pattern that decouples the storing logic
	// but still allows us to conveniently add handlers
	GetProcedures(context.Context, int32) ([]byte, error)
	UpdateProcedures(context.Context, []byte) (int, int, error)
}

type API struct {
	Store     HL7Store
	Client    HealthcareClient
	debugMode bool
}

func New(store HL7Store, client HealthcareClient, debugMode bool) http.Handler {
	a := &API{
		Store:     store,
		Client:    client,
		debugMode: debugMode,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /", a.handleMessage)
	mux.HandleFunc("GET /healthz", handleReadiness)

	mux.HandleFunc("GET /procedure/specialty", a.handleGetProceduresForSpecialtyUpdate)
	mux.HandleFunc("PUT /procedure", a.handleUpdateProcedureSpecialty)

	return mux
}

type response struct {
	Message              string `json:"message"`
	RequestContentLength int    `json:"request_content_length,omitempty"`
	HL7Path              string `json:"hl7_path,omitempty"`
	HL7Size              int    `json:"hl7_size,omitempty"`
	ControlID            string `json:"hl7_control_id,omitempty"`
	VoltaError           string `json:"volta_error,omitempty"`
}

type updateResult struct {
	Message        string `json:"message"`
	RecordsUpdated int    `json:"records_updated"`
}

func (a *API) handleGetProceduresForSpecialtyUpdate(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	cursorID := r.URL.Query().Get("cursor_id")
	if cursorID == "" {
		resp.Message = "must provide value for cursor_id param (int)"
		respondJSON(w, http.StatusBadRequest, resp)
		return
	}
	id, err := convertCursor(cursorID)
	if err != nil {
		resp.Message = err.Error()
		respondJSON(w, http.StatusBadRequest, resp)
		return
	}
	ctx := context.Background()
	procedures, err := a.Store.GetProcedures(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			resp.Message = "no procedures found"
			respondJSON(w, http.StatusNotFound, resp)
		} else {
			resp.Message = fmt.Sprintf("error getting procedures: %v", err)
			respondJSON(w, http.StatusInternalServerError, resp)
		}
		return
	}
	respondJSON(w, http.StatusOK, procedures)
}

func (a *API) handleUpdateProcedureSpecialty(w http.ResponseWriter, r *http.Request) {
	resp := updateResult{}
	if r.Body == nil {
		resp.Message = "empty request body"
		respondJSON(w, http.StatusBadRequest, resp)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			resp.Message = fmt.Sprintf("error closing client connection: %v", err)
			respondJSON(w, http.StatusBadRequest, resp)
			return
		}
	}()
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = fmt.Sprintf("could not decode request body: %v", err)
		respondJSON(w, http.StatusInternalServerError, resp)
		return
	}
	ctx := context.Background()
	requested, n, err := a.Store.UpdateProcedures(ctx, reqBody)
	if err != nil {
		resp.Message = fmt.Sprintf("could not update procedures from request: %v", err)
		respondJSON(w, http.StatusInternalServerError, resp)
		return
	}
	resp.Message = "records updated successfully"
	if n < requested {
		resp.Message = fmt.Sprintf("%d out of %d updated", n, requested)
	}
	resp.RecordsUpdated = n
	respondJSON(w, http.StatusOK, resp)
}

func (a *API) handleMessage(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	if r.Body == nil {
		resp.Message = "empty request body"
		respondJSON(w, http.StatusBadRequest, resp)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			resp.Message = fmt.Sprintf("error closing client connection: %v", err)
			respondJSON(w, http.StatusBadRequest, resp)
			return
		}
	}()
	contentLen, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err == nil {
		resp.RequestContentLength = contentLen
	}

	m, err := NewPubSubMessage(r.Body)
	if err != nil {
		resp.Message = fmt.Sprintf("error getting Healthcare API response: %v", err)
		respondJSON(w, http.StatusBadRequest, resp)
		return
	}

	hl7Path := string(m.Message.Data)
	resp.HL7Path = hl7Path
	msg, err := a.Client.GetHL7V2Message(hl7Path)
	if err != nil {
		resp.Message = "server error"
		resp.VoltaError = err.Error()
		respondJSON(w, http.StatusInternalServerError, resp)
		return
	}
	resp.HL7Size = len(msg)

	if a.debugMode {
		resp.Message = "message received!"
		respondJSON(w, http.StatusOK, resp)
		return
	}

	controlID, code, err := HandleByMsgType(a.Store, msg)
	if err != nil {
		resp.Message = "server error"
		resp.VoltaError = err.Error()
	} else if code != http.StatusCreated {
		resp.Message = "couldn't save message"
	} else {
		resp.Message = "message saved"
	}
	resp.ControlID = controlID
	respondJSON(w, code, resp)
}

func HandleByMsgType(store HL7Store, data []byte) (string, int, error) {
	var controlID string
	msg := &Message{}
	d := hl7.NewDecoder(data)
	if err := d.Decode(msg); err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("error unmarshaling HL7: %v", err)
	}
	controlID = msg.ControlID
	ctx := context.Background()

	switch msg.MsgType.Name {
	case "ORM":
		orm := &ORM{}
		if err := d.Decode(orm); err != nil {
			return "", http.StatusInternalServerError, fmt.Errorf("error unmarshaling ORM: %v", err)
		}
		if err := store.SaveORM(ctx, orm.ToOrder()); err != nil {
			return "", http.StatusInternalServerError, err
		}
		return controlID, http.StatusCreated, nil
	case "ORU":
		oru := &ORU{}
		if err := d.Decode(oru); err != nil {
			return "", http.StatusInternalServerError, fmt.Errorf("error unmarshaling ORU: %v", err)
		}
		exams := []Exam{}
		if err := d.Decode(&exams); err != nil {
			return "", http.StatusInternalServerError, fmt.Errorf("error unmarshaling exams from ORU: %v", err)
		}
		report := []Report{}
		if err := d.Decode(&report); err != nil {
			return "", http.StatusInternalServerError, fmt.Errorf("error unmarshaling report from OBX: %v", err)
		}
		obs := oru.ToObservation(GetReport(report), exams...)
		if err := store.SaveORU(ctx, obs); err != nil {
			return "", http.StatusInternalServerError, err
		}
		return controlID, http.StatusCreated, nil
	case "ADT":
		return controlID, http.StatusNotImplemented, fmt.Errorf("ADT message type not implemented")
	case "":
		return controlID, http.StatusBadRequest, fmt.Errorf("MSH.9.1 is blank--is the HL7 formatted correctly?")
	default:
		return controlID, http.StatusBadRequest, fmt.Errorf("unsupported message type: %s", msg.MsgType.Name)
	}
}

func convertCursor(s string) (int32, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	if i < 0 {
		return 0, fmt.Errorf("integer must be positive, got %s", s)
	}
	if i > 100 {
		i = 100
	}
	return int32(i), nil
}
