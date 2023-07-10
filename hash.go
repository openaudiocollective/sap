package sap

import (
	"hash/crc32"
)

func ComputeMsgIdHash(payload []byte) uint16 {
	hash := crc32.ChecksumIEEE(payload)

	// truncating to 16-bit
	// when you do a bitwise AND with 0xFFFF, it essentially keeps only the last 16 bits of the hash and discards the rest.
	return uint16(hash & 0xFFFF)
}
