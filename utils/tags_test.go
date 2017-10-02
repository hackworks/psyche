package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractTags(t *testing.T) {
	const msg = `
	during the cluster split, gocql query takes too long and opened the circuit (because of circuit timeout)
	usually it should return gocql error before we timeout.

	but gocql is not returning an error. It means, timeout happened during the "establish" phase
	This is a test for extracting tags from messages with #imp and regular but important text
	`

	tags, keywords := ExtractTags(msg, 0.05)
	require.NotEmpty(t, append(tags, keywords...))

	// Test message with ignore pattern
	tags, keywords = ExtractTags(msg+"some ignore text @search pattern in message", 0.05)
	require.Empty(t, append(tags, keywords...))
}

func TestExtractQueryTags(t *testing.T) {
	const msg = `
	Hello    this is+a message to search
	`

	op, tags := ExtractQueryTags(msg)
	require.Equal(t, byte('+'), op)
	require.NotEmpty(t, tags)
}
