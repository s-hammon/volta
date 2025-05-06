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

	msg, err := a.Client.GetHL7V2Message(string(m.Message.Data))
	if err != nil {
		logMsg.RespondJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	logMsg.Hl7Size = strconv.Itoa(len(msg))

	if a.debugMode {
		logMsg.RespondJSON(w, http.StatusOK, "received message")
		return
	}

	res, code, err := HandleByMsgType(a.DB, msg)
	if err != nil {
		logMsg.RespondJSON(w, code, err.Error())
		return
	}

	logMsg.RespondJSON(w, code, res)
}

func HandleByMsgType(db *database.Queries, data []byte) (string, int, error) {
	msg := &models.MessageModel{}
	d := hl7.NewDecoder(data)
	if err := d.Decode(msg); err != nil {
		return "error unmarshaling HL7", http.StatusInternalServerError, err
	}
	m := msg.ToEntity()
	ctx := context.Background()
	if _, err := m.ToDB(ctx, db); err != nil {
		return "error writing message to db", http.StatusInternalServerError, err
	}

	switch msg.Type.Name {
	case "ORM":
		return UpsertORM(ctx, db, d)
	case "ORU":
		return UpsertORU(ctx, db, d)
	case "ADT":
		return "ADT message type not implemented", http.StatusNotImplemented, fmt.Errorf("ADT message type not implemented")
	case "":
		return "no message type found", http.StatusBadRequest, fmt.Errorf("MSH.9.1 is blank--is the HL7 formatted correctly?")
	default:
		return "unsupported message type", http.StatusInternalServerError, fmt.Errorf("unsupported message type")
	}
}

func UpsertORM(ctx context.Context, db *database.Queries, d *hl7.Decoder) (res string, code int, err error) {
	patient := &models.PatientModel{}
	if err = d.Decode(patient); err != nil {
		return "error unmarshaling patient", http.StatusInternalServerError, err
	}
	p := patient.ToEntity()
	pID, err := p.ToDB(ctx, db)
	if err != nil {
		return "error writing patient to db", http.StatusInternalServerError, err
	}

	visit := &models.VisitModel{}
	if err = d.Decode(visit); err != nil {
		return "error unmarshaling visit", http.StatusInternalServerError, err
	}
	v := visit.ToEntity()
	sID, err := v.Site.ToDB(ctx, db)
	if err != nil {
		return "error writing site to db", http.StatusInternalServerError, err
	}
	mID, err := v.MRN.ToDB(ctx, sID, pID, db)
	if err != nil {
		return "error writing site to db", http.StatusInternalServerError, err
	}

	exam := &models.ExamModel{}
	if err = d.Decode(exam); err != nil {
		return "error unmarshaling exam", http.StatusInternalServerError, err
	}
	e := exam.ToEntity()
	if v.VisitNo == "" {
		// set this equal to the accession--it's the best we can do :/
		v.VisitNo = e.Accession
	}
	vID, err := v.ToDB(ctx, sID, mID, db)
	if err != nil {
		return "error writing visit to db", http.StatusInternalServerError, err
	}

	phID, err := e.Provider.ToDB(ctx, db)
	if err != nil {
		return "error writing physician to db", http.StatusInternalServerError, err
	}
	prID, err := e.Procedure.ToDB(ctx, sID, db)
	if err != nil {
		return "error writing procedure to db", http.StatusInternalServerError, err
	}
	if _, err = e.ToDB(ctx, vID, mID, phID, sID, prID, db); err != nil {
		return "error writing exam to db", http.StatusInternalServerError, err
	}
	return "ORM message processed", http.StatusCreated, nil
}

func UpsertORU(ctx context.Context, db *database.Queries, d *hl7.Decoder) (res string, code int, err error) {
	patient := &models.PatientModel{}
	if err = d.Decode(patient); err != nil {
		return "error unmarshaling patient", http.StatusInternalServerError, err
	}
	p := patient.ToEntity()
	pID, err := p.ToDB(ctx, db)
	if err != nil {
		return "error writing patient to db", http.StatusInternalServerError, err
	}

	visit := &models.VisitModel{}
	if err = d.Decode(visit); err != nil {
		return "error unmarshaling visit", http.StatusInternalServerError, err
	}
	v := visit.ToEntity()
	sID, err := v.Site.ToDB(ctx, db)
	if err != nil {
		return "error writing site to db", http.StatusInternalServerError, err
	}
	mID, err := v.MRN.ToDB(ctx, sID, pID, db)
	if err != nil {
		return "error writing site to db", http.StatusInternalServerError, err
	}
	vID, err := v.ToDB(ctx, sID, mID, db)
	if err != nil {
		return "error writing visit to db", http.StatusInternalServerError, err
	}

	exams := []models.ExamModel{}
	if err = d.Decode(exams); err != nil {
		return "error unmarshaling exams", http.StatusInternalServerError, err
	}
	eg := models.ToEntities(exams)
	if len(eg) < 1 {
		panic("couldn't get exam entities from models for some reason")
	}

	phID, err := eg[0].Provider.ToDB(ctx, db)
	if err != nil {
		return "error writing physician to db", http.StatusInternalServerError, err
	}
	examIDs := []int64{}
	for i, e := range eg {
		prID, err := eg[i].Procedure.ToDB(ctx, sID, db)
		if err != nil {
			return "error writing procedure to db", http.StatusInternalServerError, err
		}
		eID, err := e.ToDB(ctx, vID, mID, phID, sID, prID, db)
		if err != nil {
			return "error writing exam to db", http.StatusInternalServerError, err
		}
		examIDs = append(examIDs, eID)
	}

	report := []models.ReportModel{}
	if err = d.Decode(report); err != nil {
		return "error unmarshaling report", http.StatusInternalServerError, err
	}
	r := models.GetReport(report)
	radID, err := r.Radiologist.ToDB(ctx, db)
	if err != nil {
		return "error writing radiologist to db", http.StatusInternalServerError, err
	}
	rID, err := r.ToDB(ctx, db, radID)
	if err != nil {
		return "error writing report to db", http.StatusInternalServerError, err
	}
	switch r.Status {
	case objects.Final:
		for _, examID := range examIDs {
			if _, err := db.UpdateExamFinalReport(ctx, database.UpdateExamFinalReportParams{
				ID:            examID,
				FinalReportID: pgtype.Int8{Int64: rID, Valid: true},
			}); err != nil {
				return "error updating exam with final report", http.StatusInternalServerError, err
			}
		}
	case objects.Addendum:
		for _, examID := range examIDs {
			if _, err := db.UpdateExamAddendumReport(ctx, database.UpdateExamAddendumReportParams{
				ID:               examID,
				AddendumReportID: pgtype.Int8{Int64: rID, Valid: true},
			}); err != nil {
				return "error updating exam with final report", http.StatusInternalServerError, err
			}
		}

	}

	return "ORU message processed", http.StatusCreated, nil
}
