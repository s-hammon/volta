package objects

import "testing"

func TestNewSpecialty(t *testing.T) {
	tests := []struct {
		input string
		want  Specialty
	}{
		{"Body", Body},
		{"Breast", Breast},
		{"General", General},
		{"IR", IR},
		{"MSK", MSK},
		{"MSKIR", MSKIR},
		{"Neuro", Neuro},
		{"Spine", Spine},
		{"Unknown", Specialty("Unknown")},
		{"", Specialty("")},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if got := NewSpecialty(test.input); got != test.want {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestSpecialtyString(t *testing.T) {
	tests := []struct {
		specialty Specialty
		want      string
	}{
		{Body, "Body"},
		{Breast, "Breast"},
		{General, "General"},
		{IR, "IR"},
		{MSK, "MSK"},
		{MSKIR, "MSKIR"},
		{Neuro, "Neuro"},
		{Spine, "Spine"},
		{Specialty("Unknown"), "Unknown"},
		{Specialty(""), ""},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			if got := test.specialty.String(); got != test.want {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}
