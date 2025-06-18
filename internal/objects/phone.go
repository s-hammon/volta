package objects

import (
	"fmt"
	"strconv"

	"github.com/s-hammon/p"
)

type PhoneNumber uint64

func NewPhoneNumber(s string) (PhoneNumber, error) {
	// make sure it is 15 characters or less
	if len(s) > 15 {
		return 0, fmt.Errorf("phone number is too long: %d", len(s))
	}

	num, ok := strconv.ParseUint(s, 10, 64)
	if ok != nil {
		return 0, fmt.Errorf("could not parse phone number: %s", s)
	}

	return PhoneNumber(num), nil
}

func (ph PhoneNumber) String() string {
	return p.Format("%d", ph)
}

// format the phone number (USA & Canada only -- other countries will work, but it will look weird)
// 1234567890 -> 123-456-7890
// 11234567890 -> +1 123-456-7890
func (ph PhoneNumber) Print() string {
	s := p.Format("%d", ph)
	if len(s) == 10 {
		return p.Format("%s-%s-%s", s[0:3], s[3:6], s[6:10])
	}
	if len(s) == 11 {
		return p.Format("+1 %s-%s-%s", s[1:4], s[4:7], s[7:11])
	}

	return s
}
