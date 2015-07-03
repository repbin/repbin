// Package hashcash implements the hashcash algorithm
package hashcash

import (
	"crypto/sha256"
	"encoding/binary"
)

// Version of this release
const Version = "0.0.1 very alpha"

// NonceSize is the size of a hashcash nonce
const NonceSize = 8

// BitCount counts leading zero bits in d and returns them
func BitCount(d [32]byte) byte {
	var ret byte
	for _, x := range d {
		if x == 0 {
			ret += 8
		} else {
			if x&128 != 0x00 {
				break
			}
			if x&64 != 0x00 {
				ret++
				break
			}
			if x&32 != 0x00 {
				ret += 2
				break
			}
			if x&16 != 0x00 {
				ret += 3
				break
			}
			if x&8 != 0x00 {
				ret += 4
				break
			}
			if x&4 != 0x00 {
				ret += 5
				break
			}
			if x&2 != 0x00 {
				ret += 6
				break
			}
			if x&1 != 0x00 {
				ret += 7
			}
			break
		}
	}
	return ret
}

// TestNonce verifies that a given nonce generates bits zero bits on d when hashed
func TestNonce(d, nonce []byte, bits byte) (bool, byte) {
	bits--
	x := make([]byte, len(d)+len(nonce))
	copy(x, d)
	copy(x[len(d):], nonce)
	c := BitCount(sha256.Sum256(x))
	if c > bits {
		return true, c
	}
	return false, c
}

// NonceToUInt64 converts a nonce back to uint64
func NonceToUInt64(nonce []byte) uint64 {
	return binary.LittleEndian.Uint64(nonce)
}

// ComputeNonce is the hashcash algorithm
// c is the start value for calculation, stop is the end value. Both can be 0 to ignore segemented calculations
func ComputeNonce(d []byte, bits byte, c, stop uint64) (nonce []byte, ok bool) {
	bits--
	x := make([]byte, len(d)+8)
	copy(x, d)
	nonce = x[len(d):]
	for c < 18446744073709551615 && (stop == 0 || c < stop) {
		nonce[0] = byte(c)
		nonce[1] = byte(c >> 8)
		nonce[2] = byte(c >> 16)
		nonce[3] = byte(c >> 24)
		nonce[4] = byte(c >> 32)
		nonce[5] = byte(c >> 40)
		nonce[6] = byte(c >> 48)
		nonce[7] = byte(c >> 56)
		if BitCount(sha256.Sum256(x)) > bits {
			return nonce, true
		}
		c++
	}
	return nil, false
}
