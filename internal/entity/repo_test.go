package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateExamParam(t *testing.T) {
	obj := Exam{
		Accession:     "12345",
		CurrentStatus: NewExamStatus("SC"),
		Scheduled:     time.Date(2025, time.May, 1, 12, 0, 0, 0, time.UTC),
	}
	params := createExamParam(obj, 1, 1, 1, 1, 1)
	require.True(t, params.ScheduleDt.Valid)
	require.False(t, params.BeginExamDt.Valid)
	require.False(t, params.EndExamDt.Valid)
	require.False(t, params.ExamCancelledDt.Valid)
	require.Equal(t, time.Date(2025, time.May, 1, 12, 0, 0, 0, time.UTC), params.ScheduleDt.Time)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), params.BeginExamDt.Time)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), params.EndExamDt.Time)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), params.ExamCancelledDt.Time)
}
