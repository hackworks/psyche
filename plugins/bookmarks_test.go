package plugins

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBookmarkPlugin_extractTags(t *testing.T) {
	const msg = `
	during the cluster split, gocql query takes too long and opened the circuit (because of circuit timeout)
	usually it should return gocql error before we timeout.

	but gocql is not returning an error. It means, timeout happened during the "establish" phase
	This is a test for extracting tags from messages with #imp and regular but important text
	`

	// Since we are testing the text extraction logic only - no DB
	p := &bookmarkPlugin{}

	tags := p.extractTags(msg, 0.05)
	require.NotEmpty(t, tags)

	// Test message with ignore pattern
	tags = p.extractTags(msg+"some ignore text @search pattern in message", 0.05)
	require.Empty(t, tags)
}
