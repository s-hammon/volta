package hl7

import "strings"

var escMap = map[string]string{
	"\\.br\\": "\r",
	"\\F\\":   "|",
	"\\R\\":   "~",
	"\\S\\":   "^",
	"\\T\\":   "&",
	"\\E\\":   "\\",
	"\\X0A\\": "\n",
	"\\X0D\\": "\r",
}

func replaceEscapes(s string) string {
	var ret string
	for len(s) > 0 {
		startIdx := strings.Index(s, "\\")
		if startIdx == -1 || startIdx+1 >= len(s) {
			ret += s
			break
		}
		endIdx := strings.Index(s[startIdx+1:], "\\")
		if endIdx == -1 {
			ret += s
			break
		}
		endIdx += startIdx + 1

		ret += s[:startIdx]
		escapeSeq := s[startIdx : endIdx+1]
		ret += escaped(escapeSeq)
		s = s[endIdx+1:]
	}
	return ret
}

func escaped(s string) string {
	esc, ok := escMap[s]
	if !ok {
		return s
	}
	return esc
}
