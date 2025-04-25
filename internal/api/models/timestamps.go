package models

import "time"

const cstName = "America/Chicago"

var cst, _ = time.LoadLocation(cstName)

func convertCSTtoUTC(stringDT string) time.Time {
	dt, err := time.ParseInLocation("20060102150405", stringDT, cst)
	if err != nil {
		dt, err = time.Parse("20060102150405", stringDT)
		if err != nil {
			dt = time.Now()
		}
	}

	return dt.UTC()
}
