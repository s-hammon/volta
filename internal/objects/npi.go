package objects

import (
	"fmt"
	"strings"
)

type NPI [10]byte

func NewNPI(s string) (NPI, error) {
	if len(s) != 10 {
		return NPI{}, fmt.Errorf("NPI must be 10 digits, got %d", len(s))
	}

	npi := NPI{}
	for i, c := range s {
		if c < '0' || c > '9' {
			return NPI{}, fmt.Errorf("NPI must be all digits ('%s')", s)
		}

		// add c to npi
		npi[i] = byte(c - '0')
	}

	return npi, nil
}

func (n *NPI) Int() int {
	return int(n[0])*1000000000 +
		int(n[1])*100000000 +
		int(n[2])*10000000 +
		int(n[3])*1000000 +
		int(n[4])*100000 +
		int(n[5])*10000 +
		int(n[6])*1000 +
		int(n[7])*100 +
		int(n[8])*10 +
		int(n[9])
}

func (n *NPI) String() string {
	sb := strings.Builder{}
	for _, d := range n {
		sb.WriteByte(d + '0')
	}
	return sb.String()
}
