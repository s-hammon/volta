package hl7

type utfChars map[byte]struct{}

func newCharDict() utfChars {
	chars := make(map[byte]struct{})
	for i := 32; i < 128; i++ {
		chars[byte(i)] = struct{}{}
	}
	return chars
}
