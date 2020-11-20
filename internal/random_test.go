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
	seed, err := hex.DecodeString("0df3bf5b0df3be500df3bf5b0df3b7500df3bf5b0df3be600df3bf5b0df2be52")
	require.NoError(t, err)
	for _, test := range []struct {
		s []byte
		n uint64
		r string
	}{
		{seed, 0, "00"},
		{seed, 1, "0000"},
		{seed, 10, "000000000df3be500df3bf"},
		{seed, 20, "000000000df3be500df3bf5b0df3b7500df3bf5b0d"},
		{seed, 21, "000000000df3be500df3bf5b0df3b7500df3bf5b0df3"},
		{seed, maxMessageLength, "01"},
		{seed, 2*maxMessageLength, "02"},
		{seed, maxMessageLength+1, "0100"},
		{seed, 3*maxMessageLength+2, "030000"},
		{seed, maxMessageLength+4, "010000000d"},
		{seed, 2*maxMessageLength+4, "020000000d"},
	} {
		r := Message(test.s, test.n)
		assert.Equal(t, test.r, hex.EncodeToString(r))
	}
}
