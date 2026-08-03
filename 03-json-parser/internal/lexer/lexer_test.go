package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lexAll(input string) []Token {
	l := NewLexer(input)
	var toks []Token
	for {
		tok := l.NextToken()
		toks = append(toks, *tok)
		if tok.Type == EOF {
			break
		}
	}
	return toks
}

func TestLexer(t *testing.T) {
	// ---------------------------------------------------------------
	// 1.2 — Whitespace & structural tokens
	// ---------------------------------------------------------------

	t.Run("Structural tokens", func(t *testing.T) {
		toks := lexAll("{ } [ ] , : ")
		require.Len(t, toks, 7) // 6 structural + EOF
		assert.Equal(t, LBRACE, toks[0].Type)
		assert.Equal(t, RBRACE, toks[1].Type)
		assert.Equal(t, LBRACKET, toks[2].Type)
		assert.Equal(t, RBRACKET, toks[3].Type)
		assert.Equal(t, COMMA, toks[4].Type)
		assert.Equal(t, COLON, toks[5].Type)
		assert.Equal(t, EOF, toks[6].Type)
	})

	t.Run("Empty input yields only EOF", func(t *testing.T) {
		toks := lexAll("")
		require.Len(t, toks, 1)
		assert.Equal(t, EOF, toks[0].Type)
	})

	t.Run("Mixed whitespace around structural chars", func(t *testing.T) {
		toks := lexAll("\t{\n}\r\n")
		require.Len(t, toks, 3)
		assert.Equal(t, LBRACE, toks[0].Type)
		assert.Equal(t, RBRACE, toks[1].Type)
		assert.Equal(t, EOF, toks[2].Type)
	})

	// ---------------------------------------------------------------
	// 1.3 — Literals: true, false, null
	// ---------------------------------------------------------------

	t.Run("Valid literals", func(t *testing.T) {
		cases := map[string]TokenType{
			"true":  TRUE,
			"false": FALSE,
			"null":  NULL,
		}
		for input, want := range cases {
			toks := lexAll(input)
			require.NotEmpty(t, toks)
			assert.Equal(t, want, toks[0].Type, "input %q", input)
		}
	})

	t.Run("Near-miss literals are illegal, not partial matches", func(t *testing.T) {
		cases := []string{"tru", "True", "nulll", "fals", "nul"}
		for _, input := range cases {
			toks := lexAll(input)
			require.NotEmpty(t, toks)
			assert.Equal(t, ILLEGAL, toks[0].Type, "input %q", input)
		}
	})

	t.Run("Literal followed by structural char has correct boundary", func(t *testing.T) {
		toks := lexAll("true,")
		require.Len(t, toks, 3)
		assert.Equal(t, TRUE, toks[0].Type)
		assert.Equal(t, COMMA, toks[1].Type)
		assert.Equal(t, EOF, toks[2].Type)
	})

	// ---------------------------------------------------------------
	// 1.4 — Numbers
	// ---------------------------------------------------------------

	t.Run("Valid numbers", func(t *testing.T) {
		cases := []string{
			"0", "42", "-1", "-0", "123456789",
			"3.14", "0.5",
			"1e10", "1E10", "1e+5", "1e-5", "1.5e10",
		}
		for _, input := range cases {
			toks := lexAll(input)
			require.NotEmpty(t, toks)
			assert.Equal(t, NUMBER, toks[0].Type, "input %q", input)
			assert.Equal(t, input, toks[0].Literal, "input %q", input)
		}
	})

	t.Run("Invalid numbers are not accepted", func(t *testing.T) {
		cases := []string{"01", "1.", ".1", "1e", "+1", "-", "1.2.3", "1ee5"}
		for _, input := range cases {
			toks := lexAll(input)
			require.NotEmpty(t, toks)
			assert.NotEqual(t, NUMBER, toks[0].Type, "input %q", input)
		}
	})

	// ---------------------------------------------------------------
	// 1.5 — Strings
	// ---------------------------------------------------------------

	t.Run("Valid strings decode escapes correctly", func(t *testing.T) {
		cases := map[string]string{
			`"hello"`:          "hello",
			`"hello world"`:    "hello world",
			`"line\nbreak"`:    "line\nbreak",
			`"quote\"inside"`:  `quote"inside`,
			`"back\\slash"`:    `back\slash`,
			`"\u0041"`:         "A",
			`"\uD83D\uDE00"`:   "😀",
		}
		for input, want := range cases {
			toks := lexAll(input)
			require.NotEmpty(t, toks)
			require.Equal(t, STRING, toks[0].Type, "input %q", input)
			assert.Equal(t, want, toks[0].Literal, "input %q", input)
		}
	})

	t.Run("Invalid strings are rejected", func(t *testing.T) {
		cases := []string{
			`"unterminated`,
			"\"raw\x01control\"",
			`"\q"`,
		}
		for _, input := range cases {
			toks := lexAll(input)
			require.NotEmpty(t, toks)
			assert.NotEqual(t, STRING, toks[0].Type, "input %q", input)
		}
	})
}