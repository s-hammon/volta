package jsonhl7

import (
	"reflect"
	"testing"
)

func TestGetDelimiters(t *testing.T) {
	tests := []struct {
		name string
		msh  []byte
		want Delimiters
	}{
		{
			name: "ordinary delimiters",
			msh:  []byte("|^~\\&"),
			want: Delimiters{0: '|', 1: '~', 2: '^', 3: '&', 4: '\\'},
		},
		{
			name: "carrot delimited field",
			msh:  []byte("^|~\\&"),
			want: Delimiters{0: '^', 1: '~', 2: '|', 3: '&', 4: '\\'},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDelimiters(tt.msh)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.name, err)
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("%s: want\n%v\ngot\n%v", tt.name, tt.want, got)
			}
		})
	}
}

func TestGetDelimitersError(t *testing.T) {
	tests := []struct {
		name string
		msh  []byte
	}{
		{
			name: "invalid field delimiter",
			msh:  []byte("x^~\\&"),
		},
		{
			name: "duplicate delimiter",
			msh:  []byte("||~\\&"),
		},
		{
			name: "too few delimiters",
			msh:  []byte("|~\\&"),
		},
		{
			name: "too many delimiters",
			msh:  []byte("|^~\\&x"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getDelimiters(tt.msh)
			if err == nil {
				t.Fatalf("%s: expected error, got nil", tt.name)
			}
		})
	}
}
