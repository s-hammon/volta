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
	Site        Site              `json:"site,omitempty"`
	Code        string            `json:"code"`
	Description string            `json:"description"`
	Specialty   objects.Specialty `json:"specialty,omitempty"`
	Modality    objects.Modality  `json:"modality,omitempty"`
}
