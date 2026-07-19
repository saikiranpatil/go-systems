package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	// ---------------------------------------------------------------
	// Basic valid parsing
	// ---------------------------------------------------------------

	t.Run("Valid single header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, 23, n) // consumes only the "Host: ...\r\n" line, not the trailing \r\n
		assert.False(t, done)
	})

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Content-Type:   application/json   \r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, "application/json", headers["content-type"])
		assert.Equal(t, len(data), n)
		assert.False(t, done)
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		headers := NewHeaders()
		headers["user-agent"] = "Go-http-client/1.1"

		// First parse call for the first new header line
		data := []byte("Accept: */*\r\nConnection: keep-alive\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "*/*", headers["accept"])
		assert.False(t, done)

		// Second parse call for the remaining buffer stream
		remainingData := data[n:]
		n2, done2, err2 := headers.Parse(remainingData)
		require.NoError(t, err2)
		assert.Equal(t, "keep-alive", headers["connection"])
		assert.False(t, done2)

		// Ensure pre-existing header is preserved intact
		assert.Equal(t, "Go-http-client/1.1", headers["user-agent"])
		assert.Equal(t, len(remainingData), n2)
	})

	t.Run("Valid done - terminal CRLF", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("\r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, 2, n) // consumes the 2 bytes of CRLF
		assert.True(t, done)
	})

	t.Run("Valid empty header value", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("X-Custom:\r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, "", headers["x-custom"])
		assert.Equal(t, len(data), n)
		assert.False(t, done)
	})

	t.Run("Valid empty header value with trailing whitespace only", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("X-Custom:   \r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, "", headers["x-custom"])
		assert.Equal(t, len(data), n)
		assert.False(t, done)
	})

	// ---------------------------------------------------------------
	// Case sensitivity - keys normalized to lowercase
	// ---------------------------------------------------------------

	t.Run("Valid single header with mixed case key", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("HoSt: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, 23, n)
		assert.False(t, done)
	})

	t.Run("Valid single header with mixed case key and extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("CoNtEnT-TyPe:   application/json   \r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, "application/json", headers["content-type"])
		assert.Equal(t, len(data), n)
		assert.False(t, done)
	})

	t.Run("Valid 2 headers with mixed case existing headers", func(t *testing.T) {
		headers := NewHeaders()
		headers["user-agent"] = "Go-http-client/1.1"

		data := []byte("AcCePt: */*\r\nCoNnEcTiOn: keep-alive\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "*/*", headers["accept"])
		assert.False(t, done)

		remainingData := data[n:]
		n2, done2, err2 := headers.Parse(remainingData)
		require.NoError(t, err2)
		assert.Equal(t, "keep-alive", headers["connection"])
		assert.False(t, done2)

		assert.Equal(t, "Go-http-client/1.1", headers["user-agent"])
		assert.Equal(t, len(remainingData), n2)
	})

	t.Run("Different-case keys resolve to the same header", func(t *testing.T) {
		headers := NewHeaders()
		_, _, err := headers.Parse([]byte("Host: localhost:42069\r\n"))
		require.NoError(t, err)

		// A differently-cased repeat of the same key should combine, not create a new entry
		_, _, err = headers.Parse([]byte("HOST: example.com\r\n"))
		require.NoError(t, err)

		assert.Len(t, headers, 1)
		assert.Equal(t, "localhost:42069,example.com", headers["host"])
	})

	// ---------------------------------------------------------------
	// Duplicate header keys - combined per RFC 7230 §3.2.2
	// ---------------------------------------------------------------

	t.Run("Duplicate header keys are combined with a comma", func(t *testing.T) {
		headers := NewHeaders()

		_, _, err := headers.Parse([]byte("Accept: text/html\r\n"))
		require.NoError(t, err)
		assert.Equal(t, "text/html", headers["accept"])

		_, _, err = headers.Parse([]byte("Accept: application/json\r\n"))
		require.NoError(t, err)
		assert.Equal(t, "text/html,application/json", headers["accept"])
	})

	// ---------------------------------------------------------------
	// Incomplete / partial data (streaming buffer semantics)
	// ---------------------------------------------------------------

	t.Run("Incomplete header line, no CRLF yet", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
		assert.Len(t, headers, 0)
	})

	t.Run("Partial CRLF at end of buffer", func(t *testing.T) {
		headers := NewHeaders()
		// buffer ends mid-CRLF; must not misinterpret this as a complete line
		data := []byte("Host: localhost:42069\r")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Empty buffer", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("One complete header followed by an incomplete one", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\nAccept: text")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.False(t, done)
		// only the complete line should be consumed
		assert.Equal(t, len("Host: localhost:42069\r\n"), n)

		// the leftover "Accept: text" should not have been parsed yet
		_, ok := headers["accept"]
		assert.False(t, ok)
	})

	// ---------------------------------------------------------------
	// Invalid input
	// ---------------------------------------------------------------

	t.Run("Invalid character in header key", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("H©st: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character in field name")
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Invalid spacing header - leading space before key", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("       Host: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Invalid spacing header - space between key and colon", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host : localhost:42069\r\n")
		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Invalid spacing header - mixed case, leading space before key", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("       HoSt: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Invalid spacing header - mixed case, space between key and colon", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("HoSt : localhost:42069\r\n")
		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Invalid header - missing colon", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host localhost:42069\r\n")
		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Invalid header - empty key", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte(": localhost:42069\r\n")
		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})
}