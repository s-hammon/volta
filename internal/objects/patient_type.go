package objects

type PatientType int

const (
	OutPatient PatientType = iota + 1
	InPatient
	EdPatient
)
