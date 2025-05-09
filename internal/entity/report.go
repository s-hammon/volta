package entity

import (
	"time"

	"github.com/s-hammon/volta/internal/database"
	"github.com/s-hammon/volta/internal/objects"
)

type Report struct {
	Base
	Radiologist Physician
	Body        string
	Impression  string
	Status      objects.ReportStatus
	SubmittedDT time.Time
}

func DBtoReport(report database.Report) Report {
	return Report{
		Base: Base{
			ID:        int(report.ID),
			CreatedAt: report.CreatedAt.Time,
			UpdatedAt: report.UpdatedAt.Time,
		},
		Radiologist: Physician{
			Base: Base{ID: int(report.RadiologistID.Int64)},
		},
		Body:        report.Body,
		Impression:  report.Impression,
		Status:      objects.NewReportStatus(report.ReportStatus),
		SubmittedDT: report.SubmittedDt.Time,
	}
}
