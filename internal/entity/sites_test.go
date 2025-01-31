package entity

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/s-hammon/volta/internal/database"
)

var createdAt, _ = time.Parse("20060102150405", "20210101000000")
var updatedAt, _ = time.Parse("20060102150405", "20210101000000")

func TestSiteDBtoSite(t *testing.T) {
	dbSite := database.Site{
		ID:        1,
		CreatedAt: pgtype.Timestamp{Time: createdAt, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: updatedAt, Valid: true},
		Code:      "MHS",
		Name:      "Methodist Hospital",
		Address:   "7700 Floyd Curl Dr",
	}

	want := Site{
		Base: Base{
			ID:        1,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		Code:    "MHS",
		Name:    "Methodist Hospital",
		Address: "7700 Floyd Curl Dr",
	}

	if got := DBtoSite(dbSite); got != want {
		t.Errorf("got '%v', want '%v'", got, want)
	}
}
