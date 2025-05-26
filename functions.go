package volta

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-hammon/volta/internal/database"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/idtoken"
)

var (
	db        *pgxpool.Pool
	makotoURL string
)

const (
	defaultBodyLimit    = 1 << 20
	singleModeTimeout   = 10 * time.Second
	allModeTimeout      = 100 * time.Second
	maxWorkers          = 10
	postPredictEndpoint = "/predict"
)

type mode string

const (
	one mode = "one"
	all mode = "all"
)

type ConfigBody struct {
	Mode mode `json:"mode"`
}

type Result struct {
	RecordsSubmitted int `json:"records_submitted"`
	RecordsUpdated   int `json:"records_updated"`
}

func init() {
	makotoURL = os.Getenv("MAKOTO_URL")
	if makotoURL == "" {
		log.Fatal("MAKOTO_URL not set")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	var err error
	db, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("couldn't create connection to '%s': %v", dbURL, err)
	}
	if err = db.Ping(context.Background()); err != nil {
		log.Fatalf("couldn't reach database: %v", err)
	}
	log.Println("connected to database")
}

func AssignSpecialty(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := database.New(db)

	cfg, err := parseConfig(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client, err := idtoken.NewClient(ctx, makotoURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("makoto client: %v", err), http.StatusInternalServerError)
		return
	}

	var submitted, updated int
	switch cfg.Mode {
	case one:
		ctxOne, cancel := context.WithTimeout(ctx, singleModeTimeout)
		defer cancel()
		submitted, updated, err = assignOne(ctxOne, q, client)
	case all:
		// do job for all missing
		ctxAll, cancel := context.WithTimeout(ctx, allModeTimeout)
		defer cancel()
		submitted, updated, err = assignAll(ctxAll, q, client)
	default:
		http.Error(w, "mode must be 'one' or 'all'", http.StatusBadRequest)
		return
	}

	if err != nil && updated == 0 {
		http.Error(w, fmt.Sprintf("no records updated: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, Result{submitted, updated})
}

func handleJob(ctx context.Context, q *database.Queries, c *http.Client, procs []database.Procedure) (int, int, error) {
	if len(procs) == 0 {
		return 0, 0, nil
	}
	req, err := newPredictRequest(ctx, procs)
	if err != nil {
		return 0, 0, fmt.Errorf("error building request: %v", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("error making request to '%s': %v", req.URL.String(), err)
	}
	defer resp.Body.Close()

	preds, err := getPredictions(resp)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing predictions: %v", err)
	}
	if len(procs) != len(preds) {
		return 0, 0, fmt.Errorf("mismatch procedures (%d) and predictions (%d)", len(procs), len(preds))
	}

	submitted := len(preds)
	var (
		wg      sync.WaitGroup
		sem     = make(chan struct{}, maxWorkers)
		mu      sync.Mutex
		updated int
		lastErr error
	)

	for i, pred := range preds {
		wg.Add(1)
		sem <- struct{}{}
		id := procs[i].ID
		go func(id int32, pred string) {
			defer wg.Done()
			defer func() { <-sem }()

			err := q.UpdateProcedureSpecialty(ctx, database.UpdateProcedureSpecialtyParams{
				ID:        id,
				Specialty: pgtype.Text{String: pred, Valid: true},
			})

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				lastErr = err
			} else {
				updated++
			}
		}(id, pred)
	}
	wg.Wait()
	return submitted, updated, lastErr
}

type Predictions struct {
	Prediction []string `json:"prediction"`
}

func getPredictions(r *http.Response) ([]string, error) {
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("makoto returned %d", r.StatusCode)
	}
	d := json.NewDecoder(r.Body)
	var pr Predictions
	if err := d.Decode(&pr); err != nil {
		return nil, fmt.Errorf("decode JSON: %v", err)
	}
	return pr.Prediction, nil
}

func newPredictRequest(ctx context.Context, procs []database.Procedure) (*http.Request, error) {
	input := make([]string, len(procs))
	for i, proc := range procs {
		input[i] = proc.Description
	}
	body := struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}{Model: "specialty", Input: input}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %v", err)
	}

	url := makotoURL + postPredictEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(data)))
	return req, nil
}

func getMaxID(procedures []database.Procedure) int32 {
	maxID := procedures[0].ID
	for _, proc := range procedures {
		if proc.ID > maxID {
			maxID = proc.ID
		}
	}
	return maxID
}

func assignOne(ctx context.Context, q *database.Queries, client *http.Client) (int, int, error) {
	procs, err := q.GetProceduresForSpecialtyAssignment(ctx, 0)
	if err != nil {
		return 0, 0, err
	}
	if len(procs) == 0 {
		return 0, 0, nil
	}
	return handleJob(ctx, q, client, procs)
}

func assignAll(ctx context.Context, q *database.Queries, client *http.Client) (int, int, error) {
	sem := make(chan struct{}, maxWorkers)
	g, ctx := errgroup.WithContext(ctx)

	var (
		mu                 sync.Mutex
		submitted, updated int
	)

	cursor := int32(0)
	for {
		procs, err := q.GetProceduresForSpecialtyAssignment(ctx, cursor)
		if err != nil {
			return 0, 0, err
		}
		if len(procs) == 0 {
			break
		}
		cursor = getMaxID(procs)

		sem <- struct{}{}
		p := procs
		g.Go(func() error {
			defer func() { <-sem }()

			sub, upd, err := handleJob(ctx, q, client, p)
			mu.Lock()
			submitted += sub
			updated += upd
			mu.Unlock()
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return submitted, updated, err
	}
	return submitted, updated, nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, fmt.Sprintf("encoding response: %v", err), http.StatusInternalServerError)
	}
}

func parseConfig(r io.Reader) (cfg ConfigBody, err error) {
	d := json.NewDecoder(io.LimitReader(r, defaultBodyLimit))
	d.DisallowUnknownFields()
	if err = d.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("invalid JSON request: %v", err)
	}
	if cfg.Mode != one && cfg.Mode != all {
		return cfg, errors.New("mode must be 'one' or 'all'")
	}
	return cfg, nil
}
