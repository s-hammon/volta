package objects

import "testing"

func TestNewPhoneNumber(t *testing.T) {
	tests := []struct {
		name    string
		num     string
		wantErr bool
		wantVal PhoneNumber
	}{
		{
			name:    "valid phone number",
			num:     "1234567890",
			wantVal: PhoneNumber(1234567890),
		},
		{
			name:    "valid phone number with country code",
			num:     "11234567890",
			wantVal: PhoneNumber(11234567890),
		},
		{
			name:    "too many digits",
			num:     "12345678901234567890",
			wantErr: true,
		},
		{
			name:    "contains non-digits",
			num:     "123456789a",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPhoneNumber(tt.num)
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
