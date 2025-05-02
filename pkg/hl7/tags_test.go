package hl7

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTag(t *testing.T) {
	name, opts := parseTag("huey,dewey,louie")
	require.Equal(t, "huey", name)
	assert.True(t, opts.Contains("dewey"))
	assert.True(t, opts.Contains("louie"))
	assert.False(t, opts.Contains("scrooge"))
}
