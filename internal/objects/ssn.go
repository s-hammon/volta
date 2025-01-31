package objects

import "strings"

type SSN string

func NewSSN(s string) SSN {
	drop := func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}
	ssn := strings.Map(drop, s)

	if len(ssn) != 9 {
		return ""
	}
	if ssn[0:3] == "000" || ssn[0:3] == "666" || ssn[0:3] >= "900" {
		return ""
	}
	if ssn[3:5] == "00" {
		return ""
	}
	if ssn[5:9] == "0000" {
		return ""
	}
	if ssn == "123456789" || ssn == "987654321" {
		return ""
	}
	if ssn == strings.Repeat(string(ssn[0]), 9) {
		return ""
	}

	return SSN(ssn)
}

func (s SSN) String() string {
	return string(s)
}
