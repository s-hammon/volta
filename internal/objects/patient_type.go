package objects

type PatientType int16

const (
	OutPatient PatientType = iota + 1
	InPatient
	EdPatient
)

func NewPatientType(s string) PatientType {
	switch s {
	case "O":
		return OutPatient
	case "I":
		return InPatient
	case "E":
		return EdPatient
	default:
		return OutPatient
	}
}

func (p PatientType) Int16() int16 {
	return int16(p)
}
