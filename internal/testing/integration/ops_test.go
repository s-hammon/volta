package integration

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pressly/goose/v3"
	goosedb "github.com/pressly/goose/v3/database"
	"github.com/s-hammon/volta/internal/api"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/entity"
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
			msg := &api.Message{}
			d := hl7.NewDecoder(data)
			err = d.Decode(msg)
			require.NoError(t, err)

			switch msg.MsgType.Name {
			case "ORM":
				testUpsertORM(t, ctx, repo, d)
			case "ORU":
				testUpsertORU(t, ctx, repo, d)
			case "ADT":
				t.Logf("skipping ADT message: %s", hl7Msg.Name())
			default:
				t.Fatalf("unsupported message type: %s", msg.MsgType.Name)
			}
		})
	}

	procs, err := repo.Queries.GetProceduresForSpecialtyUpdate(ctx, 0)
	require.NoError(t, err)
	fmt.Printf("%+v", procs)
	assert.Equal(t, 6, len(procs))
	for _, proc := range procs {
		require.Equal(t, "", proc.Specialty.String)
	}
	procs[0].Specialty = pgtype.Text{String: "Breast", Valid: true}
	procs[1].Specialty = pgtype.Text{String: "MSK", Valid: true}
	procs[2].Specialty = pgtype.Text{String: "MSK", Valid: true}
	procs[3].Specialty = pgtype.Text{String: "MSK", Valid: true}
	procs[4].Specialty = pgtype.Text{String: "Vascular", Valid: true}
	procs[5].Specialty = pgtype.Text{String: "Breast", Valid: true}

	for i, proc := range procs {
		req := database.UpdateProcedureSpecialtyParams{
			ID:        proc.ID,
			Specialty: proc.Specialty,
		}
		err = repo.Queries.UpdateProcedureSpecialty(ctx, req)
		require.NoError(t, err)

		res, err := repo.Queries.GetProcedureById(ctx, procs[i].ID)
		require.NoError(t, err)
		require.Equal(t, proc.ID, res.ID)
		require.Equal(t, proc.Specialty.String, res.Specialty.String)
		require.Equal(t, "postgres", res.UpdatedBy)
	}
}

func TestORMVisit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo, _ := setupDB(t, ctx)

	data, err := hl7.HL7.ReadFile("test_hl7/5.hl7")
	require.NoError(t, err, "couldn't read test file at 5.hl7: %v", err)
	require.Greater(t, len(data), 0, "file is empty")

	orm := &api.ORM{}
	d := hl7.NewDecoder(data)
	err = d.Decode(orm)
	require.NoError(t, err)
	err = repo.SaveORM(ctx, orm.ToOrder())
	require.NoError(t, err)

	v, err := repo.Queries.GetVisitById(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, int16(1), v.PatientType)
}

func TestORMProcedure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo, _ := setupDB(t, ctx)

	data, err := hl7.HL7.ReadFile("test_hl7/5.hl7")
	require.NoError(t, err, "couldn't read test file at 5.hl7: %v", err)
	require.Greater(t, len(data), 0, "file is empty")

	orm := &api.ORM{}
	d := hl7.NewDecoder(data)
	err = d.Decode(orm)
	require.NoError(t, err)
	err = repo.SaveORM(ctx, orm.ToOrder())
	require.NoError(t, err)

	msg, err := repo.Queries.GetMessageByID(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, time.Date(2025, time.April, 4, 15, 9, 51, 0, time.UTC), msg.ReceivedAt.Time)

	site, err := repo.Queries.GetSiteById(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "BMCNE", site.Code)

	proc, err := repo.Queries.GetProcedureById(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "MAMSTOM2", proc.Code)

	t.Logf("attempting exam fetch w/ %s...\n", msg.ReceivingApplication)
	res, err := repo.Queries.GetExamBySendingAppAccession(ctx, database.GetExamBySendingAppAccessionParams{
		SendingApp: msg.ReceivingApplication,
		Accession:  "29737914",
	})
	require.NoError(t, err)
	exam := entity.DBtoExam(res)
	require.Equal(t, "N^N", exam.Priority)
	require.Equal(t, "MAMSTOM2", exam.Procedure.Code)
	require.Equal(t, "Mammogram Digital Screening Bilateral w/CAD & DBT", exam.Procedure.Description)
}

func TestExamTimestampSequence(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo, _ := setupDB(t, ctx)

	data, err := hl7.HL7.ReadFile("test_hl7/5.hl7")
	require.NoError(t, err, "couldn't read test file at 5.hl7: %v", err)
	require.Greater(t, len(data), 0, "file is empty")

	orm := &api.ORM{}
	d := hl7.NewDecoder(data)
	err = d.Decode(orm)
	require.NoError(t, err)
	err = repo.SaveORM(ctx, orm.ToOrder())
	require.NoError(t, err)

	fmt.Printf("finding exam with %s and '29737914'\n", orm.ReceivingApp)
	res, err := repo.Queries.GetExamBySendingAppAccession(ctx, database.GetExamBySendingAppAccessionParams{
		SendingApp: orm.ReceivingApp,
		Accession:  "29737914",
	})
	require.NoError(t, err)
	require.True(t, res.ScheduleDt.Valid)
	require.False(t, res.BeginExamDt.Valid)
	require.False(t, res.EndExamDt.Valid)
	require.False(t, res.ExamCancelledDt.Valid)
	exam := entity.DBtoExam(res)
	require.Equal(t, 1, exam.ID)

	assert.Equal(t, "SC", exam.CurrentStatus.String())
	assert.Equal(t, time.Date(2025, time.April, 4, 15, 9, 51, 0, time.UTC), exam.Scheduled)
	assert.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), exam.Begin)
	assert.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), exam.End)
	assert.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), exam.Cancelled)

	orm.OrderDT = "20250404103000"
	orm.OrderStatus = "IP"
	err = repo.SaveORM(ctx, orm.ToOrder())
	require.NoError(t, err)

	res, err = repo.Queries.GetExamBySendingAppAccession(ctx, database.GetExamBySendingAppAccessionParams{
		SendingApp: orm.ReceivingApp,
		Accession:  "29737914",
	})
	require.NoError(t, err)
	require.True(t, res.ScheduleDt.Valid)
	require.True(t, res.BeginExamDt.Valid)
	require.False(t, res.EndExamDt.Valid)
	require.False(t, res.ExamCancelledDt.Valid)
	exam = entity.DBtoExam(res)
	require.Equal(t, 1, exam.ID)

	assert.Equal(t, "IP", exam.CurrentStatus.String())
	assert.Equal(t, time.Date(2025, time.April, 4, 15, 9, 51, 0, time.UTC), exam.Scheduled)
	assert.Equal(t, time.Date(2025, time.April, 4, 15, 30, 0, 0, time.UTC), exam.Begin)
	assert.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), exam.End)
	assert.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), exam.Cancelled)

	orm.OrderDT = "20250404110000"
	orm.OrderStatus = "CM"
	err = repo.SaveORM(ctx, orm.ToOrder())
	require.NoError(t, err)

	res, err = repo.Queries.GetExamBySendingAppAccession(ctx, database.GetExamBySendingAppAccessionParams{
		SendingApp: orm.ReceivingApp,
		Accession:  "29737914",
	})
	require.NoError(t, err)
	require.True(t, res.ScheduleDt.Valid)
	require.True(t, res.BeginExamDt.Valid)
	require.True(t, res.EndExamDt.Valid)
	require.False(t, res.ExamCancelledDt.Valid)
	exam = entity.DBtoExam(res)
	require.Equal(t, 1, exam.ID)

	assert.Equal(t, "CM", exam.CurrentStatus.String())
	assert.Equal(t, time.Date(2025, time.April, 4, 15, 9, 51, 0, time.UTC), exam.Scheduled)
	assert.Equal(t, time.Date(2025, time.April, 4, 15, 30, 0, 0, time.UTC), exam.Begin)
	assert.Equal(t, time.Date(2025, time.April, 4, 16, 0, 0, 0, time.UTC), exam.End)
	assert.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), exam.Cancelled)

	orm.OrderDT = "20250404100951"
	orm.OrderStatus = "SC"
	err = repo.SaveORM(ctx, orm.ToOrder())
	require.NoError(t, err)

	res, err = repo.Queries.GetExamBySendingAppAccession(ctx, database.GetExamBySendingAppAccessionParams{
		SendingApp: orm.ReceivingApp,
		Accession:  "29737914",
	})
	require.NoError(t, err)
	require.True(t, res.ScheduleDt.Valid)
	require.True(t, res.BeginExamDt.Valid)
	require.True(t, res.EndExamDt.Valid)
	require.False(t, res.ExamCancelledDt.Valid)
	exam = entity.DBtoExam(res)
	require.Equal(t, 1, exam.ID)

	assert.Equal(t, "CM", exam.CurrentStatus.String())
	assert.Equal(t, time.Date(2025, time.April, 4, 15, 9, 51, 0, time.UTC), exam.Scheduled)
	assert.Equal(t, time.Date(2025, time.April, 4, 15, 30, 0, 0, time.UTC), exam.Begin)
	assert.Equal(t, time.Date(2025, time.April, 4, 16, 0, 0, 0, time.UTC), exam.End)
	assert.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), exam.Cancelled)

	data, err = hl7.HL7.ReadFile("test_hl7/9.hl7")
	require.NoError(t, err, "couldn't read test file at 9.hl7: %v", err)
	require.Greater(t, len(data), 0, "file is empty")

	d = hl7.NewDecoder(data)
	testUpsertORU(t, ctx, repo, d)

	res, err = repo.Queries.GetExamBySendingAppAccession(ctx, database.GetExamBySendingAppAccessionParams{
		SendingApp: orm.ReceivingApp,
		Accession:  "29737914",
	})
	require.NoError(t, err)
	require.True(t, res.ScheduleDt.Valid)
	require.True(t, res.BeginExamDt.Valid)
	require.True(t, res.EndExamDt.Valid)
	require.False(t, res.ExamCancelledDt.Valid)
	require.Equal(t, int64(1), res.FinalReportID.Int64)
	exam = entity.DBtoExam(res)
	require.Equal(t, 1, exam.ID)

	assert.Equal(t, "CM", exam.CurrentStatus.String())
	assert.Equal(t, time.Date(2025, time.April, 4, 15, 9, 51, 0, time.UTC), exam.Scheduled)
	assert.Equal(t, time.Date(2025, time.April, 4, 15, 30, 0, 0, time.UTC), exam.Begin)
	assert.Equal(t, time.Date(2025, time.April, 4, 16, 0, 0, 0, time.UTC), exam.End)
	assert.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), exam.Cancelled)
}

func testUpsertORM(t *testing.T, ctx context.Context, repo *entity.HL7Repo, d *hl7.Decoder) {
	t.Helper()

	orm := &api.ORM{}
	err := d.Decode(orm)
	require.NoError(t, err)
	err = repo.SaveORM(ctx, orm.ToOrder())
	require.NoError(t, err)
}

func testUpsertORU(t *testing.T, ctx context.Context, repo *entity.HL7Repo, d *hl7.Decoder) {
	t.Helper()

	oru := &api.ORU{}
	err := d.Decode(oru)
	require.NoError(t, err)
	report := []api.Report{}
	err = d.Decode(&report)
	require.NoError(t, err)
	exams := []api.Exam{}
	err = d.Decode(&exams)
	require.NoError(t, err)
	err = repo.SaveORU(ctx, oru.ToObservation(api.GetReport(report), exams...))
	require.NoError(t, err)
}

func setupDB(t *testing.T, ctx context.Context) (*entity.HL7Repo, []fs.DirEntry) {
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

	return entity.NewRepo(pool), hl7Messages
}
