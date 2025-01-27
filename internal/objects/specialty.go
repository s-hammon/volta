package objects

const (
	Body    Specialty = "Body"
	Breast  Specialty = "Breast"
	General Specialty = "General"
	IR      Specialty = "IR"
	MSK     Specialty = "MSK"
	MSKIR   Specialty = "MSKIR"
	Neuro   Specialty = "Neuro"
	Spine   Specialty = "Spine"
)

type Specialty string

func (s Specialty) String() string {
	return string(s)
}

type Modality string
