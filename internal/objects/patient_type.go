package objects

type PatientType int16

const (
	OutPatient PatientType = iota + 1
	InPatient
	EdPatient
)

func (p PatientType) Int16() int16 {
	return int16(p)
}
