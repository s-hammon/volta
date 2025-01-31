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

func TestPhoneNumberString(t *testing.T) {
	num, err := NewPhoneNumber("1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := num.String(); got != "1234567890" {
		t.Errorf("got '%v', want '1234567890'", got)
	}
}

func TestPhoneNumberPrint(t *testing.T) {
	tests := []struct {
		name string
		num  PhoneNumber
		want string
	}{
		{
			name: "10 digit number",
			num:  PhoneNumber(1234567890),
			want: "123-456-7890",
		},
		{
			name: "11 digit number",
			num:  PhoneNumber(11234567890),
			want: "+1 123-456-7890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.num.Print(); got != tt.want {
				t.Errorf("got '%v', want '%v'", got, tt.want)
			}
		})
	}
}
