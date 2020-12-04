package internal

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomAccountSeed(t *testing.T) {
	for _, test := range []struct {
		s uint64
		n uint64
		r string
	}{
		{0, 0, "0000000b0000000b0000000b0000000b0000000b0000000b0000000b0000000b"},
		{0, 1, "deece67ddeece678deece67ddeece678deece67ddeece678deece67ddeece678"},
		{1, 0, "0000000b0000000b0000000b0000000b0000000b0000000b0000000b0000000a"},
		{1, 1, "deece67ddeece678deece67ddeece678deece67ddeece678deece67ddeece679"},
		{12345, 67890, "c4cbd6fdc4cbe655c4cbd6fdc4cbd655c4cbd6fdc4cbd665c4cbd6fdc4cbd65c"},
		{67890, 12345, "0df3bf5b0df3be500df3bf5b0df3b7500df3bf5b0df3be600df3bf5b0df2be52"},
	} {
		r := AccountSeed(test.s, test.n)
		assert.Equal(t, test.r, hex.EncodeToString(r))
	}
}

func TestMessage(t *testing.T) {
	seed, err := hex.DecodeString("1234567890abcdef")
	require.NoError(t, err)
	template := Template(seed)
	for _, test := range []struct {
		t []byte
		n uint64
		r string
	}{
		{template, 0, "00"},
		{template, 1, "0000"},
		{template, 10, "000000001234567890abcd"},
		{template, 20, "000000001234567890abcdef1234567890abcdef12"},
		{template, 21, "000000001234567890abcdef1234567890abcdef1234"},
		{template, 95, "000000001234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12345678"},
		{template, maxMessageLength, "01"},
		{template, 2 * maxMessageLength, "02"},
		{template, maxMessageLength + 1, "0100"},
		{template, 3*maxMessageLength + 2, "030000"},
		{template, maxMessageLength + 4, "0100000012"},
		{template, 2*maxMessageLength + 4, "0200000012"},
	} {
		r := Message(test.t, test.n)
		assert.Equal(t, test.r, hex.EncodeToString(r))
	}
}
