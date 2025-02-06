package objects

const (
	OutPatient PatientType = iota + 1
	InPatient
	EdPatient
)

// TODO: do we really need a custom type?

type PatientType int16

func (p PatientType) Int16() int16 {
	return int16(p)
}
