package objects

import "testing"

func TestNewReportStatus(t *testing.T) {
	tests := []struct {
		input string
		want  ReportStatus
	}{
		{"P", Pending},
		{"F", Final},
		{"A", Addendum},
		{"X", Pending},
		{"", Pending},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if got := NewReportStatus(test.input); got != test.want {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}
