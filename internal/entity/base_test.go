package entity

// import (
// 	"errors"
// 	"testing"
// )

// func TestExtractErrCode(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		err  error
// 		want string
// 	}{
// 		{
// 			name: "nil error",
// 			err:  nil,
// 			want: "",
// 		},
// 		{
// 			name: "alternate error",
// 			err:  errors.New("not a SQL error"),
// 			want: "",
// 		},
// 		{
// 			name: "no parentheses",
// 			err:  errors.New("ERROR:  syntax error at or near \"a\" SQLSTATE 42601"),
// 			want: "",
// 		},
// 		{
// 			name: "no left parenthesis",
// 			err:  errors.New("ERROR:  syntax error at or near \"a\" )SQLSTATE 42601)"),
// 			want: "",
// 		},
// 		{
// 			name: "not SQLSTATE",
// 			err:  errors.New("ERROR:  syntax error at or near \"a\" (SQLERR 42601)"),
// 			want: "",
// 		},
// 		{
// 			name: "valid error",
// 			err:  errors.New("ERROR:  syntax error at or near \"a\" (SQLSTATE 42601)"),
// 			want: "42601",
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := extractErrCode(tt.err); got != tt.want {
// 				t.Errorf("got '%v', want '%v'", got, tt.want)
// 			}
// 		})
// 	}
// }
