package entity

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
)

type dbErr struct {
	entityName string
	errMsg     error
}

func (e dbErr) Error() string {
	return fmt.Sprintf("error writing %s to database: %v", e.entityName, e.errMsg)
}

type HL7Repo struct {
	DB      *pgxpool.Pool
	Queries *database.Queries
}

func NewRepo(db *pgxpool.Pool) *HL7Repo {
	return &HL7Repo{DB: db, Queries: database.New(db)}
}

type Order struct {
	Message   Message
	Patient   Patient
	Visit     Visit
	Provider  Physician
	Procedure Procedure
	Exam      Exam
}

type Observation struct {
	Message   Message
	Patient   Patient
	Visit     Visit
	Provider  Physician
	Procedure Procedure
	Exams     []Exam
	Report    Report
}

func (h *HL7Repo) SaveORM(ctx context.Context, orm *Order) error {
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("error with rollback: %v\n", err)
		}
	}()

	qtx := h.Queries.WithTx(tx)
	var sID, prID int32
	var msgID, pID, vID, mID, phID int64
	// TODO: bundle below 4 into goroutines
	msgID, err = qtx.CreateMessage(ctx, createMessageParam(orm.Message))
	if err != nil {
		return dbErr{"message", err}
	}
	pID, err = qtx.CreatePatient(ctx, createPatientParam(orm.Patient, msgID))
	if err != nil {
		return dbErr{"patient", err}
	}
	sID, err = qtx.CreateSite(ctx, createSiteParam(orm.Visit.Site, msgID))
	if err != nil {
		return dbErr{"site", err}
	}
	phID, err = qtx.CreatePhysician(ctx, createPhysicianParam(orm.Provider, msgID))
	if err != nil {
		return dbErr{"ordering physician", err}
	}

	// TODO: see if we can use chans to fire these off when above are finished
	prID, err = qtx.CreateProcedure(ctx, createProcedureParam(orm.Procedure, sID, msgID))
	if err != nil {
		return dbErr{"procedure", err}
	}
	mID, err = qtx.CreateMrn(ctx, createMrnParam(orm.Visit.MRN, sID, pID, msgID))
	if err != nil {
		return dbErr{"MRN", err}
	}
	if orm.Visit.VisitNo == "" {
		// set this equal to the accession--it's the best we can do :/
		orm.Visit.VisitNo = orm.Exam.Accession
	}
	vID, err = qtx.CreateVisit(ctx, createVisitParam(orm.Visit, sID, mID, msgID))
	if err != nil {
		return dbErr{"visit", err}
	}
	if _, err = qtx.CreateExam(ctx, createExamParam(
		orm.Exam,
		sID, prID,
		vID, mID, phID, msgID,
	)); err != nil {
		return dbErr{"exam", err}
	}

	return tx.Commit(ctx)
}

func (h *HL7Repo) SaveORU(ctx context.Context, oru *Observation) error {
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("error with rollback: %v\n", err)
		}
	}()

	qtx := h.Queries.WithTx(tx)
	var sID, prID int32
	var msgID, pID, vID, mID, phID, radID, rID int64
	// TODO: bundle below 4 into goroutines
	msgID, err = qtx.CreateMessage(ctx, createMessageParam(oru.Message))
	if err != nil {
		return dbErr{"message", err}
	}
	pID, err = qtx.CreatePatient(ctx, createPatientParam(oru.Patient, msgID))
	if err != nil {
		return dbErr{"patient", err}
	}
	sID, err = qtx.CreateSite(ctx, createSiteParam(oru.Visit.Site, msgID))
	if err != nil {
		return dbErr{"site", err}
	}
	phID, err = qtx.CreatePhysician(ctx, createPhysicianParam(oru.Provider, msgID))
	if err != nil {
		return dbErr{"ordering physician", err}
	}

	// TODO: see if we can use chans to fire these off when above are finished -- basically create a DAG
	mID, err = qtx.CreateMrn(ctx, createMrnParam(oru.Visit.MRN, sID, pID, msgID))
	if err != nil {
		return dbErr{"MRN", err}
	}
	if oru.Visit.VisitNo == "" {
		// set this equal to the accession--it's the best we can do :/
		oru.Visit.VisitNo = oru.Exams[0].Accession
	}
	vID, err = qtx.CreateVisit(ctx, createVisitParam(oru.Visit, sID, mID, msgID))
	if err != nil {
		return dbErr{"visit", err}
	}
	radID, err = qtx.CreatePhysician(ctx, createPhysicianParam(oru.Report.Radiologist, msgID))
	if err != nil {
		return dbErr{"radiologist", err}
	}
	rID, err = qtx.CreateReport(ctx, createReportParam(oru.Report, radID, msgID))
	if err != nil {
		return dbErr{"report", err}
	}
	for _, exam := range oru.Exams {
		prID, err = qtx.CreateProcedure(ctx, createProcedureParam(exam.Procedure, sID, msgID))
		if err != nil {
			return dbErr{"procedure", err}
		}
		var eID int64
		eID, err = qtx.GetExamIDBySiteIDAccession(ctx, getExamIDParam(exam, sID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				eID, err = qtx.CreateExam(ctx, createExamParam(
					exam,
					sID, prID,
					vID, mID, phID, msgID,
				))
				if err != nil {
					return dbErr{"exam", err}
				}
			} else {
				return fmt.Errorf("error retrieving exam ID for accession %s: %v", exam.Accession, err)
			}
		}
		switch oru.Report.Status {
		case objects.Final:
			if _, err := qtx.UpdateExamFinalReport(ctx, updateExamFinalParam(eID, rID)); err != nil {
				return fmt.Errorf("error updating exam with final report: %v", err)
			}
		case objects.Addendum:
			if _, err := qtx.UpdateExamAddendumReport(ctx, updateExamAddendumParam(eID, rID)); err != nil {
				return fmt.Errorf("error updating exam with addendum report: %v", err)
			}
		}
	}

	return tx.Commit(ctx)
}

func createMessageParam(obj Message) database.CreateMessageParams {
	params := database.CreateMessageParams{}
	params.FieldSeparator = obj.FieldSeparator
	params.EncodingCharacters = obj.EncodingChars
	params.SendingApplication = obj.SendingApp
	params.SendingFacility = obj.SendingFac
	params.ReceivingApplication = obj.ReceivingApp
	params.ReceivingFacility = obj.ReceivingFac
	params.ReceivedAt = pgtype.Timestamp{Time: obj.DateTime, Valid: true}
	params.MessageType = obj.Type
	params.TriggerEvent = obj.TriggerEvent
	params.ControlID = obj.ControlID
	params.ProcessingID = obj.ProcessingID
	params.VersionID = obj.Version

	return params
}

func createPatientParam(obj Patient, msgID int64) database.CreatePatientParams {
	params := database.CreatePatientParams{}
	params.MessageID = pgtype.Int8{Int64: msgID, Valid: true}
	params.FirstName = obj.Name.First
	params.LastName = obj.Name.Last
	params.MiddleName = pgtype.Text{String: obj.Name.Middle, Valid: true}
	params.Suffix = pgtype.Text{String: obj.Name.Suffix, Valid: true}
	params.Prefix = pgtype.Text{String: obj.Name.Prefix, Valid: true}
	params.Degree = pgtype.Text{String: obj.Name.Degree, Valid: true}
	params.Dob = pgtype.Date{Time: obj.DOB, Valid: true}
	params.Sex = obj.Sex
	if obj.SSN != "" {
		params.Ssn.String = obj.SSN.String()
	}
	return params
}

func createSiteParam(obj Site, msgID int64) database.CreateSiteParams {
	params := database.CreateSiteParams{}
	params.MessageID = pgtype.Int8{Int64: msgID, Valid: true}
	params.Code = obj.Code
	params.Address = obj.Address
	if obj.Name != "" {
		params.Name.String = obj.Name
	}
	return params
}

func createPhysicianParam(obj Physician, msgID int64) database.CreatePhysicianParams {
	params := database.CreatePhysicianParams{}
	params.MessageID = pgtype.Int8{Int64: msgID, Valid: true}
	params.FirstName = obj.Name.First
	params.LastName = obj.Name.Last
	params.MiddleName = pgtype.Text{String: obj.Name.Middle, Valid: true}
	params.Suffix = pgtype.Text{String: obj.Name.Suffix, Valid: true}
	params.Prefix = pgtype.Text{String: obj.Name.Prefix, Valid: true}
	params.Degree = pgtype.Text{String: obj.Name.Degree, Valid: true}
	params.AppCode = obj.AppCode
	if obj.NPI != "" {
		params.Npi.String = obj.NPI
	}
	return params
}

func createProcedureParam(obj Procedure, siteID int32, msgID int64) database.CreateProcedureParams {
	params := database.CreateProcedureParams{}
	params.MessageID = pgtype.Int8{Int64: msgID, Valid: true}
	params.SiteID = pgtype.Int4{Int32: siteID, Valid: true}
	params.Code = obj.Code
	params.Description = obj.Description
	return params
}

func createMrnParam(obj MRN, siteID int32, patientID, msgID int64) database.CreateMrnParams {
	params := database.CreateMrnParams{}
	params.MessageID = pgtype.Int8{Int64: msgID, Valid: true}
	params.SiteID = siteID
	params.PatientID = pgtype.Int8{Int64: patientID, Valid: true}
	params.Mrn = obj.Value
	return params
}

func createVisitParam(obj Visit, siteID int32, mrnID, msgID int64) database.CreateVisitParams {
	params := database.CreateVisitParams{}
	params.MessageID = pgtype.Int8{Int64: msgID, Valid: true}
	params.SiteID = pgtype.Int4{Int32: siteID, Valid: true}
	params.MrnID = pgtype.Int8{Int64: mrnID, Valid: true}
	params.Number = obj.VisitNo
	params.PatientType = obj.Type.Int16()
	return params
}

func createExamParam(obj Exam, siteID, procID int32, visitID, mrnID, physID, msgID int64) database.CreateExamParams {
	params := database.CreateExamParams{}
	params.MessageID = pgtype.Int8{Int64: msgID, Valid: true}
	params.SiteID = pgtype.Int4{Int32: siteID, Valid: true}
	params.ProcedureID = pgtype.Int4{Int32: procID, Valid: true}
	params.VisitID = pgtype.Int8{Int64: visitID, Valid: true}
	params.MrnID = pgtype.Int8{Int64: mrnID, Valid: true}
	params.OrderingPhysicianID = pgtype.Int8{Int64: physID, Valid: true}
	params.Accession = obj.Accession
	params.CurrentStatus = obj.CurrentStatus.String()
	obj.timestamp(&params)
	return params
}

func createReportParam(obj Report, radID, msgID int64) database.CreateReportParams {
	params := database.CreateReportParams{}
	params.MessageID = pgtype.Int8{Int64: msgID, Valid: true}
	params.RadiologistID = pgtype.Int8{Int64: radID, Valid: true}
	params.Body = obj.Body
	params.Impression = obj.Impression
	params.ReportStatus = obj.Status.String()
	params.SubmittedDt = pgtype.Timestamp{Time: obj.SubmittedDT, Valid: true}
	return params
}

func getExamIDParam(obj Exam, siteID int32) database.GetExamIDBySiteIDAccessionParams {
	params := database.GetExamIDBySiteIDAccessionParams{}
	params.SiteID = pgtype.Int4{Int32: siteID, Valid: true}
	params.Accession = obj.Accession
	return params
}

func updateExamFinalParam(examID, reportID int64) database.UpdateExamFinalReportParams {
	params := database.UpdateExamFinalReportParams{}
	params.ID = examID
	params.FinalReportID = pgtype.Int8{Int64: reportID, Valid: true}
	return params
}

func updateExamAddendumParam(examID, reportID int64) database.UpdateExamAddendumReportParams {
	params := database.UpdateExamAddendumReportParams{}
	params.ID = examID
	params.AddendumReportID = pgtype.Int8{Int64: reportID, Valid: true}
	return params
}
