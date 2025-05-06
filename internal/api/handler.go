package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/api/models"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
	"github.com/s-hammon/volta/pkg/hl7"
)

type HealthcareClient interface {
	GetHL7V2Message(string) ([]byte, error)
}

type API struct {
	DB        *database.Queries
	Client    HealthcareClient
	debugMode bool
}

func New(db *database.Queries, client HealthcareClient, debugMode bool) http.Handler {
	a := &API{
		DB:        db,
		Client:    client,
		debugMode: debugMode,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /", a.handleMessage)
	mux.HandleFunc("GET /healthz", handleReadiness)

	return mux
}

func (a *API) handleMessage(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Message              string `json:"message"`
		RequestContentLength int    `json:"request_content_length,omitempty"`
		HL7Path              string `json:"hl7_path,omitempty"`
		HL7Size              int    `json:"hl7_size,omitempty"`
		ControlID            string `json:"hl7_control_id,omitempty"`
		VoltaError           string `json:"volta_error,omitempty"`
	}
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
	contentLen, _ := strconv.Atoi(r.Header.Get("Content-Length"))
	resp.RequestContentLength = contentLen

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

	controlID, code, err := HandleByMsgType(a.DB, msg)
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

func HandleByMsgType(db *database.Queries, data []byte) (string, int, error) {
	var controlID string
	msg := &models.MessageModel{}
	d := hl7.NewDecoder(data)
	if err := d.Decode(msg); err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("error unmarshaling HL7: %v", err)
	}
	m := msg.ToEntity()
	controlID = m.ControlID
	ctx := context.Background()
	if _, err := m.ToDB(ctx, db); err != nil {
		return controlID, http.StatusInternalServerError, fmt.Errorf("error writing message to db: %v", err)
	}

	switch msg.Type.Name {
	case "ORM":
		code, err := UpsertORM(ctx, db, d)
		return controlID, code, err
	case "ORU":
		code, err := UpsertORU(ctx, db, d)
		return controlID, code, err
	case "ADT":
		return controlID, http.StatusNotImplemented, fmt.Errorf("ADT message type not implemented")
	case "":
		return controlID, http.StatusBadRequest, fmt.Errorf("MSH.9.1 is blank--is the HL7 formatted correctly?")
	default:
		// return "unsupported message type", http.StatusInternalServerError, fmt.Errorf("unsupported message type")
		return controlID, http.StatusBadRequest, fmt.Errorf("unsupported message type: %s", msg.Type.Name)
	}
}

func UpsertORM(ctx context.Context, db *database.Queries, d *hl7.Decoder) (code int, err error) {
	code = http.StatusInternalServerError // guilty until proven innocent

	patient := &models.PatientModel{}
	if err = d.Decode(patient); err != nil {
		return code, fmt.Errorf("error unmarshaling patient: %v", err)
	}
	p := patient.ToEntity()
	pID, err := p.ToDB(ctx, db)
	if err != nil {
		return code, fmt.Errorf("error writing patient to db: %v", err)
	}

	visit := &models.VisitModel{}
	if err = d.Decode(visit); err != nil {
		return code, fmt.Errorf("error unmarshaling visit: %v", err)
	}
	v := visit.ToEntity()
	sID, err := v.Site.ToDB(ctx, db)
	if err != nil {
		return code, fmt.Errorf("error writing site to db: %v", err)
	}
	mID, err := v.MRN.ToDB(ctx, sID, pID, db)
	if err != nil {
		return code, fmt.Errorf("error writing MRN to db: %v", err)
	}

	exam := &models.ExamModel{}
	if err = d.Decode(exam); err != nil {
		return code, fmt.Errorf("error unmarshaling exam: %v", err)
	}
	e := exam.ToEntity()
	if v.VisitNo == "" {
		// set this equal to the accession--it's the best we can do :/
		v.VisitNo = e.Accession
	}
	vID, err := v.ToDB(ctx, sID, mID, db)
	if err != nil {
		return code, fmt.Errorf("error writing visit to db: %v", err)
	}

	phID, err := e.Provider.ToDB(ctx, db)
	if err != nil {
		return code, fmt.Errorf("error writing physician to db: %v", err)
	}
	prID, err := e.Procedure.ToDB(ctx, sID, db)
	if err != nil {
		return code, fmt.Errorf("error writing procedure to db: %v", err)
	}
	if _, err = e.ToDB(ctx, vID, mID, phID, sID, prID, db); err != nil {
		return code, fmt.Errorf("error writing exam to db: %v", err)
	}
	code = http.StatusCreated
	return code, nil
}

func UpsertORU(ctx context.Context, db *database.Queries, d *hl7.Decoder) (code int, err error) {
	code = http.StatusInternalServerError // guilty until proven innocent

	patient := &models.PatientModel{}
	if err = d.Decode(patient); err != nil {
		return code, fmt.Errorf("error unmarshaling patient: %v", err)
	}
	p := patient.ToEntity()
	pID, err := p.ToDB(ctx, db)
	if err != nil {
		return code, fmt.Errorf("error writing patient to db: %v", err)
	}

	visit := &models.VisitModel{}
	if err = d.Decode(visit); err != nil {
		return code, fmt.Errorf("error unmarshaling visit: %v", err)
	}
	v := visit.ToEntity()
	sID, err := v.Site.ToDB(ctx, db)
	if err != nil {
		return code, fmt.Errorf("error writing site to db: %v", err)
	}
	mID, err := v.MRN.ToDB(ctx, sID, pID, db)
	if err != nil {
		return code, fmt.Errorf("error writing MRN to db: %v", err)
	}
	vID, err := v.ToDB(ctx, sID, mID, db)
	if err != nil {
		return code, fmt.Errorf("error writing visit to db: %v", err)
	}

	exams := []models.ExamModel{}
	if err = d.Decode(&exams); err != nil {
		return code, fmt.Errorf("error unmarshaling exams: %v", err)
	}
	eg := models.ToEntities(exams)
	if len(eg) < 1 {
		panic("couldn't get exam entities from models for some reason")
	}

	phID, err := eg[0].Provider.ToDB(ctx, db)
	if err != nil {
		return code, fmt.Errorf("error writing physician to db: %v", err)
	}
	examIDs := []int64{}
	for i, e := range eg {
		prID, err := eg[i].Procedure.ToDB(ctx, sID, db)
		if err != nil {
			return code, fmt.Errorf("error writing procedure to db: %v", err)
		}
		eID, err := e.ToDB(ctx, vID, mID, phID, sID, prID, db)
		if err != nil {
			return code, fmt.Errorf("error writing exam to db: %v", err)
		}
		examIDs = append(examIDs, eID)
	}

	report := []models.ReportModel{}
	if err = d.Decode(&report); err != nil {
		return code, fmt.Errorf("error unmarshaling report: %v", err)
	}
	r := models.GetReport(report)
	radID, err := r.Radiologist.ToDB(ctx, db)
	if err != nil {
		return code, fmt.Errorf("error writing radiologist to db: %v", err)
	}
	rID, err := r.ToDB(ctx, db, radID)
	if err != nil {
		return code, fmt.Errorf("error writing report to db: %v", err)
	}
	switch r.Status {
	case objects.Final:
		for _, examID := range examIDs {
			if _, err := db.UpdateExamFinalReport(ctx, database.UpdateExamFinalReportParams{
				ID:            examID,
				FinalReportID: pgtype.Int8{Int64: rID, Valid: true},
			}); err != nil {
				return code, fmt.Errorf("error updated exam with final report: %v", err)
			}
		}
	case objects.Addendum:
		for _, examID := range examIDs {
			if _, err := db.UpdateExamAddendumReport(ctx, database.UpdateExamAddendumReportParams{
				ID:               examID,
				AddendumReportID: pgtype.Int8{Int64: rID, Valid: true},
			}); err != nil {
				return code, fmt.Errorf("error updating exam with addendum report: %v", err)
			}
		}

	}

	code = http.StatusCreated
	return code, nil
}
