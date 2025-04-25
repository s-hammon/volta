package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertCSTtoUTC(t *testing.T) {
	// normal CST to UTC
	cstString := "20250425135027"
	utcDT := convertCSTtoUTC(cstString)
	assert.Equal(
		t,
		time.Date(2025, 04, 25, 13, 50, 27, 0, cst).UTC(),
		utcDT,
	)
}
