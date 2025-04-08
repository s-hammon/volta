package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pressly/goose/v3"
	goosedb "github.com/pressly/goose/v3/database"
	"github.com/s-hammon/volta/internal/api"
	"github.com/s-hammon/volta/internal/api/models"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
	"github.com/s-hammon/volta/internal/testing/testdb"
	"github.com/s-hammon/volta/pkg/hl7"
)

func TestHL7Upserts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo, hl7Messages := setupDB(t, ctx)

	for _, hl7Msg := range hl7Messages {
		if hl7Msg.IsDir() || !strings.HasSuffix(hl7Msg.Name(), ".hl7") {
			continue
		}
		t.Run(hl7Msg.Name(), func(t *testing.T) {
			fPath := filepath.Join("test_hl7", hl7Msg.Name())
			data, err := hl7.HL7.ReadFile(fPath)
			if err != nil {
				t.Fatalf("couldn't read test file at %s: %v", fPath, err)
			}
			if len(data) == 0 {
				t.Fatalf("file %s is empty", fPath)
			}
			msg, err := hl7.NewMessage(data)
			if err != nil {
				t.Fatalf("failed to parse HL7 message from file %s: %v", fPath, err)
			}

			switch extractMsgType(t, msg) {
			case "ORM":
				testUpsertORM(t, ctx, repo, msg)
			case "ORU":
				testUpsertORU(t, ctx, repo, msg)
			case "ADT":
				t.Logf("skipping ADT message: %s", hl7Msg.Name())
			default:
				t.Fatalf("unsupported message type: %s", extractMsgType(t, msg))
			}
		})
	}
}

func testUpsertORU(t *testing.T, ctx context.Context, repo api.DB, msg hl7.Message) {
	t.Helper()

	var oru models.ORU
	if err := json.Unmarshal(msg, &oru); err != nil {
		t.Fatalf("failed to unmarshal HL7 message: %v", err)
	}
	p := oru.PID.ToEntity()
	m := oru.MSH.ToEntity()
	v := oru.PV1.ToEntity(m.SendingFac, oru.PID.MRN)

	orderGroups, err := oru.GroupOrders()
	if err != nil {
		t.Fatalf("error grouping orders and exams: %v", err)
	}
	oe := models.NewOrderEntities(v.Site.Code, oru.PID.MRN, orderGroups...)

	siteID, err := v.Site.ToDB(ctx, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create site: %v", err)
	}
	site, err := repo.Queries.GetSiteById(ctx, siteID)
	if err != nil {
		t.Fatalf("failed to get site by code: %v", err)
	}
	assertEqual(t, siteID, site.ID)

	patientID, err := p.ToDB(ctx, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create patient: %v", err)
	}
	patient, err := repo.Queries.GetPatientById(ctx, patientID)
	if err != nil {
		t.Fatalf("failed to get patient by ID: %v", err)
	}
	assertEqual(t, patientID, patient.ID)

	mrnID, err := v.MRN.ToDB(ctx, siteID, patientID, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create mrn: %v", err)
	}
	mrn, err := repo.Queries.GetMrnById(ctx, mrnID)
	if err != nil {
		t.Fatalf("failed to get mrn by ID: %v", err)
	}
	assertEqual(t, mrnID, mrn.ID)

	visitID, err := v.ToDB(ctx, siteID, mrnID, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create visit: %v", err)
	}
	visit, err := repo.Queries.GetVisitById(ctx, visitID)
	if err != nil {
		t.Fatalf("failed to get visit by ID: %v", err)
	}
	assertEqual(t, visitID, visit.ID)

	provider := oe[0].GetOrder().Provider
	physicianID, err := provider.ToDB(ctx, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create physician: %v", err)
	}
	physician, err := repo.Queries.GetPhysicianById(ctx, physicianID)
	if err != nil {
		t.Fatalf("failed to get physician by ID: %v", err)
	}
	assertEqual(t, physicianID, physician.ID)

	examIDs := make([]int64, len(oe))
	for i, orderEntity := range oe {
		thisOrder := orderEntity.GetOrder()
		orderID, orderStatus, err := thisOrder.ToDB(ctx, siteID, visitID, mrnID, physicianID, repo.Queries)
		if err != nil {
			t.Fatalf("failed to create order: %v", err)
		}
		order, err := repo.Queries.GetOrderById(ctx, orderID)
		if err != nil {
			t.Fatalf("failed to get order by ID: %v", err)
		}
		assertEqual(t, orderID, order.ID)

		thisExam := orderEntity.GetExam()
		procedureID, err := thisExam.Procedure.ToDB(ctx, siteID, repo.Queries)
		if err != nil {
			t.Fatalf("failed to create procedure: %v", err)
		}
		procedure, err := repo.Queries.GetProcedureById(ctx, procedureID)
		if err != nil {
			t.Fatalf("failed to get procedure by ID: %v", err)
		}
		assertEqual(t, procedureID, procedure.ID)

		examID, err := thisExam.ToDB(ctx, orderID, visitID, mrnID, siteID, procedureID, orderStatus, repo.Queries)
		if err != nil {
			t.Fatalf("failed to create exam: %v", err)
		}
		exam, err := repo.Queries.GetExamById(ctx, examID)
		if err != nil {
			t.Fatalf("failed to get exam by ID: %v", err)
		}
		assertEqual(t, examID, exam.ID)

		examIDs[i] = examID
	}

	reportModel := oru.GetReport()
	report, err := repo.Queries.CreateReport(ctx, database.CreateReportParams{
		RadiologistID: pgtype.Int8{Int64: int64(physicianID), Valid: true},
		Body:          reportModel.Body,
		Impression:    reportModel.Impression,
		ReportStatus:  reportModel.Status.String(),
		SubmittedDt:   pgtype.Timestamp{Time: reportModel.SubmittedDT, Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create report: %v", err)
	}
	getReport, err := repo.Queries.GetReportById(ctx, report.ID)
	if err != nil {
		t.Fatalf("failed to get report by ID: %v", err)
	}
	assertEqual(t, report.ID, getReport.ID)
	assertEqual(t, report.Body, getReport.Body)
	assertEqual(t, report.Impression, getReport.Impression)
	assertEqual(t, report.ReportStatus, getReport.ReportStatus)
	assertEqual(t, report.SubmittedDt, getReport.SubmittedDt)
	assertEqual(t, report.RadiologistID, getReport.RadiologistID)
	assertEqual(t, report.CreatedAt, getReport.CreatedAt)
	assertEqual(t, report.UpdatedAt, getReport.UpdatedAt)

	switch reportModel.Status {
	case objects.Final:
		for _, examID := range examIDs {
			exam, err := repo.Queries.UpdateExamFinalReport(ctx, database.UpdateExamFinalReportParams{
				ID:            int64(examID),
				FinalReportID: pgtype.Int8{Int64: int64(report.ID), Valid: true},
			})
			if err != nil {
				t.Fatalf("failed to update exam with final report: %v", err)
			}
			examID, err := repo.Queries.GetExamById(ctx, exam.ID)
			if err != nil {
				t.Fatalf("failed to get exam by ID: %v", err)
			}
			assertEqual(t, exam.ID, examID.ID)
			assertEqual(t, exam.FinalReportID, examID.FinalReportID)
			assertEqual(t, exam.UpdatedAt, examID.UpdatedAt)
		}
	case objects.Addendum:
		for _, examID := range examIDs {
			exam, err := repo.Queries.UpdateExamAddendumReport(ctx, database.UpdateExamAddendumReportParams{
				ID:               int64(examID),
				AddendumReportID: pgtype.Int8{Int64: int64(report.ID), Valid: true},
			})
			if err != nil {
				t.Fatalf("failed to update exam with addendum report: %v", err)
			}
			examID, err := repo.Queries.GetExamById(ctx, exam.ID)
			if err != nil {
				t.Fatalf("failed to get exam by ID: %v", err)
			}
			assertEqual(t, exam.ID, examID.ID)
			assertEqual(t, exam.AddendumReportID, examID.AddendumReportID)
			assertEqual(t, exam.UpdatedAt, examID.UpdatedAt)
		}
	}
}

func testUpsertORM(t *testing.T, ctx context.Context, repo api.DB, msg hl7.Message) {
	t.Helper()

	var orm models.ORM
	if err := json.Unmarshal(msg, &orm); err != nil {
		t.Fatalf("failed to unmarshal HL7 message: %v", err)
	}
	m := orm.MSH.ToEntity()
	v := orm.PV1.ToEntity(m.SendingFac, orm.PID.MRN)
	p := orm.PID.ToEntity()
	o := orm.ORC.ToEntity()
	e := orm.OBR.ToEntity(v.Site.Code, o.CurrentStatus, orm.PID.MRN)

	writeMsg, err := m.ToDB(ctx, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}
	readMsg, err := repo.Queries.GetMessageByID(ctx, writeMsg.ID)
	if err != nil {
		t.Fatalf("failed to get message by ID: %v", err)
	}
	assertEqual(t, writeMsg.ID, readMsg.ID)

	siteID, err := v.Site.ToDB(ctx, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create site: %v", err)
	}
	site, err := repo.Queries.GetSiteById(ctx, siteID)
	if err != nil {
		t.Fatalf("failed to get site by code: %v", err)
	}
	assertEqual(t, siteID, site.ID)

	patientID, err := p.ToDB(ctx, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create patient: %v", err)
	}
	patient, err := repo.Queries.GetPatientById(ctx, patientID)
	if err != nil {
		t.Fatalf("failed to get patient by ID: %v", err)
	}
	assertEqual(t, patientID, patient.ID)

	mrnID, err := v.MRN.ToDB(ctx, siteID, patientID, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create mrn: %v", err)
	}
	mrn, err := repo.Queries.GetMrnById(ctx, mrnID)
	if err != nil {
		t.Fatalf("failed to get mrn by ID: %v", err)
	}
	assertEqual(t, mrnID, mrn.ID)

	if v.VisitNo == "" {
		v.VisitNo = o.Number
	}
	visitID, err := v.ToDB(ctx, siteID, mrnID, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create visit: %v", err)
	}
	visit, err := repo.Queries.GetVisitById(ctx, visitID)
	if err != nil {
		t.Fatalf("failed to get visit by ID: %v", err)
	}
	assertEqual(t, visitID, visit.ID)

	physicianID, err := o.Provider.ToDB(ctx, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create physician: %v", err)
	}
	physician, err := repo.Queries.GetPhysicianById(ctx, physicianID)
	if err != nil {
		t.Fatalf("failed to get physician by ID: %v", err)
	}
	assertEqual(t, physicianID, physician.ID)

	orderID, orderStatus, err := o.ToDB(ctx, siteID, visitID, mrnID, physicianID, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}
	order, err := repo.Queries.GetOrderById(ctx, orderID)
	if err != nil {
		t.Fatalf("failed to get order by ID: %v", err)
	}
	assertEqual(t, orderID, order.ID)
	assertEqual(t, orderStatus, order.CurrentStatus)

	procedureID, err := e.Procedure.ToDB(ctx, siteID, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create procedure: %v", err)
	}
	procedure, err := repo.Queries.GetProcedureById(ctx, procedureID)
	if err != nil {
		t.Fatalf("failed to get procedure by ID: %v", err)
	}
	assertEqual(t, procedureID, procedure.ID)

	examID, err := e.ToDB(ctx, orderID, visitID, mrnID, siteID, procedureID, orderStatus, repo.Queries)
	if err != nil {
		t.Fatalf("failed to create exam: %v", err)
	}
	exam, err := repo.Queries.GetExamById(ctx, examID)
	if err != nil {
		t.Fatalf("failed to get exam by ID: %v", err)
	}
	assertEqual(t, examID, exam.ID)
}

func extractMsgType(t *testing.T, msg hl7.Message) string {
	t.Helper()

	var msgMap map[string]any
	if err := json.Unmarshal(msg, &msgMap); err != nil {
		t.Fatalf("failed to unmarshal HL7 message: %v", err)
	}
	msh, ok := msgMap["MSH"].(map[string]any)
	if !ok {
		t.Fatalf("failed extracting MSH")
	}
	msh9, ok := msh["MSH.9"].(map[string]any)
	if !ok {
		t.Fatalf("failed extracting MSH.9")
	}
	msgType, ok := msh9["MSH.9.1"].(string)
	if !ok {
		t.Fatalf("failed to extract message type from HL7 message")
	}
	return msgType
}

func setupDB(t *testing.T, ctx context.Context) (api.DB, []fs.DirEntry) {
	t.Helper()

	pool, cleanup, err := testdb.NewPGPool()
	if err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}
	t.Cleanup(cleanup)
	if err = pool.Ping(ctx); err != nil {
		t.Fatalf("failed to ping postgres: %v", err)
	}
	db, err := sql.Open("pgx", pool.Config().ConnString())
	if err != nil {
		t.Fatalf("failed to open pgx connection: %v", err)
	}
	if err = db.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping pgx connection: %v", err)
	}

	schemaFiles := collectMigrations(t, migrationsDir)
	p, err := goose.NewProvider(goosedb.DialectPostgres, db, os.DirFS(migrationsDir))
	if err != nil {
		t.Fatalf("failed to create goose provider: %v", err)
	}
	if len(schemaFiles) != len(p.ListSources()) {
		t.Fatalf("expected %d migrations, got %d", len(schemaFiles), len(p.ListSources()))
	}
	res, err := p.Up(ctx)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	if len(res) != len(schemaFiles) {
		t.Fatalf("expected %d migrations to run, got %d", len(schemaFiles), len(res))
	}
	if err = db.Close(); err != nil {
		t.Fatalf("failed to close db connection: %v", err)
	}
	hl7Messages, err := hl7.HL7.ReadDir("test_hl7")
	if err != nil {
		t.Fatalf("failed to read embedded test directory: %v", err)
	}
	if len(hl7Messages) == 0 {
		t.Fatalf("no HL7 messages found in test directory")
	}

	return api.NewDB(pool), hl7Messages
}

func assertEqual[T any](t *testing.T, expected, actual T) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
