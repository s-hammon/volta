package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestORM_ToOrderStatusDT(t *testing.T) {
	orm := &ORM{
		OrderDT:     "20250501120000",
		OrderStatus: "SC",
	}
	order := orm.ToOrder()
	require.Equal(t, "SC", order.Exam.CurrentStatus.String())
	require.Equal(t, time.Date(2025, time.May, 1, 17, 0, 0, 0, time.UTC), order.Exam.Scheduled)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), order.Exam.Begin)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), order.Exam.End)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), order.Exam.Cancelled)

	orm = &ORM{
		OrderDT:     "20250501120000",
		OrderStatus: "IP",
	}
	order = orm.ToOrder()
	require.Equal(t, "IP", order.Exam.CurrentStatus.String())
	require.Equal(t, time.Date(2025, time.May, 1, 17, 0, 0, 0, time.UTC), order.Exam.Begin)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), order.Exam.Scheduled)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), order.Exam.End)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), order.Exam.Cancelled)
	orm = &ORM{
		OrderDT:     "20250501120000",
		OrderStatus: "CM",
	}
	order = orm.ToOrder()
	require.Equal(t, "CM", order.Exam.CurrentStatus.String())
	require.Equal(t, time.Date(2025, time.May, 1, 17, 0, 0, 0, time.UTC), order.Exam.End)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), order.Exam.Begin)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), order.Exam.Scheduled)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), order.Exam.Cancelled)
	orm = &ORM{
		OrderDT:     "20250501120000",
		OrderStatus: "CA",
	}
	order = orm.ToOrder()
	require.Equal(t, "CA", order.Exam.CurrentStatus.String())
	require.Equal(t, time.Date(2025, time.May, 1, 17, 0, 0, 0, time.UTC), order.Exam.Cancelled)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), order.Exam.Begin)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), order.Exam.End)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), order.Exam.Scheduled)
}
