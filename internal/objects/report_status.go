package objects

type ReportStatus string

const (
	Pending  ReportStatus = "P"
	Final    ReportStatus = "F"
	Addendum ReportStatus = "A"
)

func NewReportStatus(s string) ReportStatus {
	switch s {
	case "P":
		return Pending
	case "F":
		return Final
	case "A":
		return Addendum
	default:
		return Pending
	}
}

func (r ReportStatus) String() string {
	return string(r)
}
