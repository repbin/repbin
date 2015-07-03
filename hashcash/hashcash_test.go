package hashcash

import (
	"encoding/hex"
	"testing"
)

func TestBitCount(t *testing.T) {
	c := BitCount([32]byte{0x00, 0x00, 0x08, 0x02})
	if c != 20 {
		t.Error("Count error")
	}
}

func TestHashCash(t *testing.T) {
	d := []byte("abcefghi")
	n, ok := ComputeNonce(d, 10, 0, 0)
	if !ok {
		t.Error("No match found")
	}
	if hex.EncodeToString(n) != "b301000000000000" {
		t.Errorf("Nonce error b301000000000000 != %s", hex.EncodeToString(n))
	}
}

func TestComputeParallel(t *testing.T) {
	d := []byte("abcefghi")
	bits := byte(21)
	nonce, _ := ComputeNonceSelect(d, bits, 0, 0) // 26.235s
	ok, _ := TestNonce(d, nonce, bits)
	if !ok {
		t.Error("No Nonce found")
	}
}

func TestTestNonce(t *testing.T) {
	d := []byte("abcefghi")
	n := []byte{0x09, 0x0e, 0x28, 0x00, 0x00, 0x00, 0x00, 0x00}
	ok, _ := TestNonce(d, n, 20)
	if !ok {
		t.Error("Nonce verification failed")
	}
	n = []byte{0x09, 0x0e, 0x28, 0x00, 0x00, 0x00, 0x00, 0x01}
	ok, c := TestNonce(d, n, 20)
	if ok {
		t.Errorf("Nonce verification MUST fail: %d < 20", c)
	}
}
