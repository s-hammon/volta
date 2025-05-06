package integration

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			msg := &models.MessageModel{}
			d := hl7.NewDecoder(data)
			if err := d.Decode(msg); err != nil {
				t.Fatalf("error unmarshaling HL7: %v", err)
			}
			m := msg.ToEntity()
			dbMsg, err := m.ToDB(ctx, repo.Queries)
			assert.NoError(t, err)
			assert.NotEqual(t, 0, dbMsg.ID)
			getMsg, err := repo.Queries.GetMessageByID(ctx, dbMsg.ID)
			assert.NoError(t, err)
			assert.Equal(t, dbMsg.ID, getMsg.ID)

			switch msg.Type.Name {
			case "ORM":
				testUpsertORM(t, ctx, repo, d)
			case "ORU":
				testUpsertORU(t, ctx, repo, d)
			case "ADT":
				t.Logf("skipping ADT message: %s", hl7Msg.Name())
			default:
				t.Fatalf("unsupported message type: %s", msg.Type.Name)
			}
		})
	}
}

func testUpsertORU(t *testing.T, ctx context.Context, repo api.DB, d *hl7.Decoder) {
	t.Helper()

	patient := &models.PatientModel{}
	err := d.Decode(patient)
	require.NoError(t, err)

	p := patient.ToEntity()
	pID, err := p.ToDB(ctx, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, pID)
	dbPatient, err := repo.Queries.GetPatientById(ctx, pID)
	require.NoError(t, err)
	require.Equal(t, pID, dbPatient.ID)

	visit := &models.VisitModel{}
	err = d.Decode(visit)
	require.NoError(t, err)
	v := visit.ToEntity()

	sID, err := v.Site.ToDB(ctx, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, sID)
	site, err := repo.Queries.GetSiteById(ctx, sID)
	require.NoError(t, err)
	require.Equal(t, sID, site.ID)

	mID, err := v.MRN.ToDB(ctx, sID, pID, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, mID)
	mrn, err := repo.Queries.GetMrnById(ctx, mID)
	require.NoError(t, err)
	require.Equal(t, mID, mrn.ID)

	vID, err := v.ToDB(ctx, sID, mID, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, vID)
	dbVisit, err := repo.Queries.GetVisitById(ctx, vID)
	require.NoError(t, err)
	require.Equal(t, vID, dbVisit.ID)

	exams := []models.ExamModel{}
	err = d.Decode(&exams)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(exams), 1)
	eg := models.ToEntities(exams)
	require.Equal(t, len(exams), len(eg))

	phID, err := eg[0].Provider.ToDB(ctx, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, phID)
	physician, err := repo.Queries.GetPhysicianById(ctx, phID)
	require.NoError(t, err)
	require.Equal(t, phID, physician.ID)

	examIDs := []int64{}
	for i, e := range eg {
		prID, err := eg[i].Procedure.ToDB(ctx, sID, repo.Queries)
		require.NoError(t, err)
		require.NotEqual(t, 0, prID)
		proc, err := repo.Queries.GetProcedureById(ctx, prID)
		require.NoError(t, err)
		require.Equal(t, prID, proc.ID)

		eID, err := e.ToDB(ctx, vID, mID, phID, sID, prID, repo.Queries)
		require.NoError(t, err)
		require.NotEqual(t, 0, eID)
		dbExam, err := repo.Queries.GetExamById(ctx, eID)
		require.NoError(t, err)
		require.Equal(t, eID, dbExam.ID)
		assertNotNullStatusTimestamp(t, dbExam)
		examIDs = append(examIDs, eID)
	}
	report := []models.ReportModel{}
	err = d.Decode(&report)
	require.NoError(t, err)
	require.Greater(t, len(report), 0)
	r := models.GetReport(report)
	radID, err := r.Radiologist.ToDB(ctx, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, radID)
	rID, err := r.ToDB(ctx, repo.Queries, radID)
	require.NoError(t, err)
	require.NotEqual(t, 0, rID)
	dbReport, err := repo.Queries.GetReportById(ctx, rID)
	require.NoError(t, err)
	reportEnt, err := repo.Queries.GetReportByRadID(ctx, pgtype.Int8{Int64: radID, Valid: true})
	require.NoError(t, err)
	require.Equal(t, rID, dbReport.ID)
	assert.Equal(t, r.Body, dbReport.Body)
	assert.Equal(t, r.Impression, dbReport.Impression)
	assert.Equal(t, "F", reportEnt.ReportStatus)
	assert.Equal(t, r.SubmittedDT, reportEnt.SubmittedDt.Time)
	assert.Equal(t, radID, dbReport.RadiologistID.Int64)

	switch r.Status {
	case objects.Final:
		for _, examID := range examIDs {
			exam, err := repo.Queries.UpdateExamFinalReport(ctx, database.UpdateExamFinalReportParams{
				ID:            examID,
				FinalReportID: pgtype.Int8{Int64: dbReport.ID, Valid: true},
			})
			require.NoError(t, err)
			require.NotEqual(t, 0, exam.ID)
			updatedExam, err := repo.Queries.GetExamById(ctx, exam.ID)
			require.NoError(t, err)
			assert.Equal(t, exam.ID, updatedExam.ID)
			assert.Equal(t, exam.FinalReportID.Int64, dbReport.ID)
			assertNotNullStatusTimestamp(t, updatedExam)
		}
	case objects.Addendum:
		for _, examID := range examIDs {
			exam, err := repo.Queries.UpdateExamAddendumReport(ctx, database.UpdateExamAddendumReportParams{
				ID:               examID,
				AddendumReportID: pgtype.Int8{Int64: dbReport.ID, Valid: true},
			})
			require.NoError(t, err)
			require.NotEqual(t, 0, exam.ID)
			updatedExam, err := repo.Queries.GetExamById(ctx, exam.ID)
			require.NoError(t, err)
			assert.Equal(t, exam.ID, updatedExam.ID)
			assert.Equal(t, exam.AddendumReportID, dbReport.ID)
			assertNotNullStatusTimestamp(t, updatedExam)
		}
	}
}

func testUpsertORM(t *testing.T, ctx context.Context, repo api.DB, d *hl7.Decoder) {
	t.Helper()

	patient := &models.PatientModel{}
	err := d.Decode(patient)
	require.NoError(t, err)

	p := patient.ToEntity()
	pID, err := p.ToDB(ctx, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, pID)
	dbPatient, err := repo.Queries.GetPatientById(ctx, pID)
	require.NoError(t, err)
	require.Equal(t, pID, dbPatient.ID)

	visit := &models.VisitModel{}
	err = d.Decode(visit)
	require.NoError(t, err)
	v := visit.ToEntity()

	sID, err := v.Site.ToDB(ctx, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, sID)
	site, err := repo.Queries.GetSiteById(ctx, sID)
	require.NoError(t, err)
	require.Equal(t, sID, site.ID)

	mID, err := v.MRN.ToDB(ctx, sID, pID, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, mID)
	mrn, err := repo.Queries.GetMrnById(ctx, mID)
	require.NoError(t, err)
	require.Equal(t, mID, mrn.ID)

	vID, err := v.ToDB(ctx, sID, mID, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, vID)
	dbVisit, err := repo.Queries.GetVisitById(ctx, vID)
	require.NoError(t, err)
	require.Equal(t, vID, dbVisit.ID)

	exam := &models.ExamModel{}
	err = d.Decode(exam)
	require.NoError(t, err)
	e := exam.ToEntity()

	phID, err := e.Provider.ToDB(ctx, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, phID)
	prID, err := e.Procedure.ToDB(ctx, sID, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, prID)

	eID, err := e.ToDB(ctx, vID, mID, phID, sID, prID, repo.Queries)
	require.NoError(t, err)
	require.NotEqual(t, 0, eID)
	dbExam, err := repo.Queries.GetExamById(ctx, eID)
	require.NoError(t, err)
	require.Equal(t, eID, dbExam.ID)
	assertNotNullStatusTimestamp(t, dbExam)
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

func assertNotNullStatusTimestamp(t *testing.T, exam database.Exam) {
	t.Helper()

	switch exam.CurrentStatus {
	case "SC":
		if exam.ScheduleDt.Time.IsZero() {
			t.Fatalf("exam schedule timestamp is empty for status %s", exam.CurrentStatus)
		}
	case "IP":
		if exam.BeginExamDt.Time.IsZero() {
			t.Fatalf("exam begin timestamp is empty for status %s", exam.CurrentStatus)
		}
	case "CM":
		if exam.EndExamDt.Time.IsZero() {
			t.Fatalf("exam end timestamp is empty for status %s", exam.CurrentStatus)
		}
	case "":
		t.Fatalf("exam status is empty")
	default:
		t.Fatalf("error: unexpected exam status %s", exam.CurrentStatus)
	}
}
