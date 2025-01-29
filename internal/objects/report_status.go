package objects

const (
	Pending ReportStatus = iota + 1
	Final
	Addendum
)

type ReportStatus int

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

func (r ReportStatus) Int16() int16 {
	return int16(r)
}

func (r ReportStatus) String() string {
	switch r {
	case Pending:
		return "P"
	case Final:
		return "F"
	case Addendum:
		return "A"
	default:
		return "P"
	}
}
