package objects

import "testing"

func TestPatientTypeInt16(t *testing.T) {
	tests := []struct {
		name  string
		pType PatientType
		want  int16
	}{
		{"OutPatient", OutPatient, int16(1)},
		{"InPatient", InPatient, int16(2)},
		{"EdPatient", EdPatient, int16(3)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.pType.Int16(); got != test.want {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}
