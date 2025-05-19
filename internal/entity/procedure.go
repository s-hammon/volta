package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
)

type Procedure struct {
	Base
	Site        Site
	Code        string
	Description string
	Specialty   objects.Specialty
	Modality    objects.Modality
}

func (h *HL7Repo) GetProcedures(ctx context.Context, cursorID int32) ([]byte, error) {
	res, err := h.Queries.GetProceduresForSpecialtyUpdate(ctx, cursorID)
	if err != nil {
		return nil, err
	}
	procedures := make([]Procedure, len(res))
	for i, r := range res {
		procedures[i] = Procedure{
			Base: Base{
				ID:        int(r.ID),
				CreatedAt: r.CreatedAt.Time,
				UpdatedAt: r.UpdatedAt.Time,
			},
			Code:        r.Code,
			Description: r.Description,
			Specialty:   objects.NewSpecialty(r.Specialty.String),
		}
	}
	return json.Marshal(procedures)
}

type updateProcedureRequest struct {
	ID        int32  `json:"id"`
	Specialty string `json:"specialty"`
}

func (r *updateProcedureRequest) toArg() database.UpdateProcedureSpecialtyParams {
	return database.UpdateProcedureSpecialtyParams{
		ID:        r.ID,
		Specialty: pgtype.Text{String: r.Specialty},
	}
}

func (h *HL7Repo) UpdateProcedures(ctx context.Context, data []byte) (requested, updated int, err error) {
	var requests []updateProcedureRequest
	if err = json.Unmarshal(data, &requests); err != nil {
		return 0, 0, err
	}
	requested = len(requests)
	if requested == 0 {
		return requested, 0, nil
	}

	var (
		mu        sync.Mutex
		wg        sync.WaitGroup
		lastError error
	)

	for _, request := range requests {
		wg.Add(1)
		go func(req updateProcedureRequest) {
			defer wg.Done()
			if err = h.Queries.UpdateProcedureSpecialty(ctx, request.toArg()); err != nil {
				mu.Lock()
				lastError = err
				mu.Unlock()
				return
			}
			mu.Lock()
			updated++
			mu.Unlock()
		}(request)
	}
	wg.Wait()

	if updated <= 0 && lastError != nil {
		return requested, updated, fmt.Errorf("failed to update any records. last error: %v", lastError)
	}
	return requested, updated, nil
}
