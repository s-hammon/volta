package hl7

import (
	"embed"
)

//go:embed test_hl7/*.hl7
var HL7 embed.FS
