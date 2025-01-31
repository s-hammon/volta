package objects

import (
	"testing"
)

func TestNewNPI(t *testing.T) {
	tests := []struct {
		name    string
		npi     string
		wantErr bool
		wantVal NPI
	}{
		{
			name:    "valid npi",
			npi:     "1234567890",
			wantVal: NPI{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		},
		{
			name:    "not enough digits",
			npi:     "12345",
			wantErr: true,
		},
		{
			name:    "too many digits",
			npi:     "12345678901",
			wantErr: true,
		},
		{
			name:    "contains non-digits",
			npi:     "123456789a",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewNPI(tt.npi)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s expected error: %v, got %v", tt.name, tt.wantErr, err)
				return
			}

			if err == nil && got != tt.wantVal {
				t.Errorf("%s: want %v, got %v", tt.name, tt.wantVal, got)
			}
		})
	}
}

func TestNPIString(t *testing.T) {
	npi, err := NewNPI("1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := npi.String(); got != "1234567890" {
		t.Errorf("got '%v', want '1234567890'", got)
	}
}
