package objects

type PatientType int

const (
	outPatient PatientType = iota + 1
	inPatient
	edPatient
)
