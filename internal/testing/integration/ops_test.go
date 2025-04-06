package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"github.com/s-hammon/volta/internal/api"
	"github.com/s-hammon/volta/internal/testing/testdb"
	"github.com/s-hammon/volta/pkg/hl7"
)

// TODO: fine-tune this so we know what HL7 messages we are processing
// i.e. we put in 5 ORMs for 4 exams, there should be 4 exams in the DB
func TestHL7Upserts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

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
	p, err := goose.NewProvider(database.DialectPostgres, db, os.DirFS(migrationsDir))
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
	repo := api.NewDB(pool)
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
			msgType := extractMsgType(t, msg)
			_, err = api.HandleByMsgType(repo, msgType, msg)
			if err != nil {
				if msgType == "ADT" {
					t.Logf("skipping ADT message: %v", err)
					return
				}
				t.Fatalf("file %s (type %s): %v", hl7Msg.Name(), msgType, err)
			}
		})
	}
	// for now, just see if there is data at all
	exams, err := repo.GetAllExams(context.Background())
	if err != nil {
		t.Fatalf("failed to get all exams: %v", err)
	}
	if len(exams) == 0 {
		t.Fatalf("no exams found in database")
	}
	reports, err := repo.GetAllReports(context.Background())
	if err != nil {
		t.Fatalf("failed to get all reports: %v", err)
	}
	if len(reports) == 0 {
		t.Fatalf("no reports found in database")
	}
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
