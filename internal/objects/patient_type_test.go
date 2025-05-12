package objects

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
			got := test.pType.Int16()
			require.Equal(t, test.want, got)
		})
	}
}
