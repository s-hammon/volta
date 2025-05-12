package hl7

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceEscapes(t *testing.T) {
	s := "Methodist Specialty \\T\\ Transplant"
	got := replaceEscapes(s)
	require.Equal(t, "Methodist Specialty & Transplant", got)

	s = "CHEST\\E\\ABD\\E\\PEL W\\CONTRAST"
	got = replaceEscapes(s)
	require.Equal(t, "CHEST\\ABD\\PEL W\\CONTRAST", got)

	s = "first line\\.br\\second line"
	got = replaceEscapes(s)
	require.Equal(t, "first line\rsecond line", got)

	s = "unrecognized escape \\O\\"
	got = replaceEscapes(s)
	require.Equal(t, "unrecognized escape \\O\\", got)
}
