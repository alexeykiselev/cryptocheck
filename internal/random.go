package internal

import (
	"bytes"
	"encoding/binary"
)

const (
	multiplier       = 0x5deece66d
	addend           = 0xb
	mask             = (1 << 48) - 1
	maxMessageLength = 150 * 1024
)

func AccountSeed(seed, n uint64) []byte {
	nv := (n*multiplier + addend) & mask
	value := nv<<32 | nv
	r := make([]byte, 32)
	binary.BigEndian.PutUint64(r[0:8], value^(seed&0xf000f000f000f000))
	binary.BigEndian.PutUint64(r[8:16], value^(seed&0x0f000f000f000f00))
	binary.BigEndian.PutUint64(r[16:24], value^(seed&0x00f000f000f000f0))
	binary.BigEndian.PutUint64(r[24:32], value^(seed&0x000f000f000f000f))
	return r
}

func Message(template []byte, n uint64) []byte {
	l := int(n%maxMessageLength) + 1
	r := make([]byte, l)
	sb := make([]byte, 4)
	binary.LittleEndian.PutUint32(sb, uint32(n/maxMessageLength))
	e := 4
	if l < e {
		e = l
	}
	copy(r[:e], sb)

	if l > 4 {
		copy(r[4:], template)
	}
	return r
}

func Template(seed []byte) []byte {
	return bytes.Repeat(seed, maxMessageLength/8)
}
