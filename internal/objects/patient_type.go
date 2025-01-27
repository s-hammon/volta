package objects

const (
	OutPatient PatientType = iota + 1
	InPatient
	EdPatient
)

type PatientType int

func (p PatientType) Int16() int16 {
	return int16(p)
}
