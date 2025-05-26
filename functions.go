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

type model string

const (
	specialty model = "specialty"
	modality  model = "modality"
)

type ConfigBody struct {
	Model model `json:"model"`
	Mode  mode  `json:"mode"`
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
	if err != nil || db.Ping(context.Background()) != nil {
		log.Fatalf("couldn't create connection to '%s': %v", dbURL, err)
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
		submitted, updated, err = assignOne(ctxOne, q, client, cfg.Model)
	case all:
		// do job for all missing
		ctxAll, cancel := context.WithTimeout(ctx, allModeTimeout)
		defer cancel()
		submitted, updated, err = assignAll(ctxAll, q, client, cfg.Model)
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

func handleJob(ctx context.Context, q *database.Queries, c *http.Client, procs []database.Procedure, model model) (int, int, error) {
	if len(procs) == 0 {
		return 0, 0, nil
	}
	req, err := newPredictRequest(ctx, model, procs)
	if err != nil {
		return 0, 0, fmt.Errorf("error building request: %v", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("error making request to '%s': %v", req.URL.String(), err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("couldn't close request body: %v", err)
		}
	}()

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

			var err error
			switch model {
			case specialty:
				err = q.UpdateProcedureSpecialty(ctx, database.UpdateProcedureSpecialtyParams{
					ID:        id,
					Specialty: pgtype.Text{String: pred, Valid: true},
				})
			case modality:
				err = q.UpdateProcedureModality(ctx, database.UpdateProcedureModalityParams{
					ID:       id,
					Modality: pgtype.Text{String: pred, Valid: true},
				})
			}

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

func newPredictRequest(ctx context.Context, model model, procs []database.Procedure) (*http.Request, error) {
	input := make([]string, len(procs))
	for i, proc := range procs {
		input[i] = proc.Description
	}
	body := struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}{Model: string(model), Input: input}
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

func getPredictions(r *http.Response) ([]string, error) {
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("makoto returned %d", r.StatusCode)
	}
	var pr Predictions
	if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("decode JSON: %v", err)
	}
	return pr.Prediction, nil
}

func getMaxID(procedures []database.Procedure) int32 {
	var maxID int32
	for _, proc := range procedures {
		if proc.ID > maxID {
			maxID = proc.ID
		}
	}
	return maxID
}

func assignOne(ctx context.Context, q *database.Queries, client *http.Client, model model) (int, int, error) {
	var (
		procs []database.Procedure
		err   error
	)
	switch model {
	case specialty:
		procs, err = q.GetProceduresForSpecialtyUpdate(ctx, 0)
	case modality:
		procs, err = q.GetProceduresForModalityUpdate(ctx, 0)
	}
	if err != nil {
		return 0, 0, err
	}
	if len(procs) == 0 {
		return 0, 0, nil
	}
	return handleJob(ctx, q, client, procs, model)
}

func assignAll(ctx context.Context, q *database.Queries, client *http.Client, model model) (int, int, error) {
	sem := make(chan struct{}, maxWorkers)
	g, ctx := errgroup.WithContext(ctx)

	var (
		mu                 sync.Mutex
		submitted, updated int
		getFn              func(context.Context, int32) ([]database.Procedure, error)
	)
	switch model {
	case specialty:
		getFn = q.GetProceduresForSpecialtyUpdate
	case modality:
		getFn = q.GetProceduresForModalityUpdate
	}

	cursor := int32(0)
	for {
		procs, err := getFn(ctx, cursor)
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

			sub, upd, err := handleJob(ctx, q, client, p, model)
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

func parseConfig(r io.Reader) (ConfigBody, error) {
	var cfg ConfigBody
	d := json.NewDecoder(io.LimitReader(r, defaultBodyLimit))
	d.DisallowUnknownFields()
	if err := d.Decode(&cfg); err != nil {
		return ConfigBody{}, fmt.Errorf("invalid JSON: %v", err)
	}
	if cfg.Model != specialty && cfg.Model != modality {
		return ConfigBody{}, errors.New("model must be 'specialty' or 'modality'")
	}
	if cfg.Mode != one && cfg.Mode != all {
		return ConfigBody{}, errors.New("mode must be 'one' or 'all'")
	}
	return cfg, nil
}
