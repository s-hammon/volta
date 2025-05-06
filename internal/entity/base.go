package entity

import (
	// "strings"
	"time"
)

type Base struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// func extractErrCode(err error) string {
// 	if err == nil {
// 		return ""
// 	}
//
// 	strErr := err.Error()
//
// 	parts := strings.Split(strErr, "(")
// 	if len(parts) < 2 {
// 		return ""
// 	}
//
// 	code := strings.TrimRight(parts[1], ")")
// 	codeSplit := strings.Split(code, " ")
// 	if len(codeSplit) < 1 || codeSplit[0] != "SQLSTATE" {
// 		return ""
// 	}
//
// 	return codeSplit[1]
// }
