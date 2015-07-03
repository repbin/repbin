package utils

import (
	"bytes"
	"testing"
)

func TestMaxRead(t *testing.T) {
	mintest := bytes.NewBuffer(make([]byte, 100))
	minread, err := MaxRead(101, mintest)
	if err != nil {
		t.Errorf("Minread failed: %s", err)
	}
	if len(minread) != 100 {
		t.Errorf("Minread short read: %d", len(minread))
	}
	prectest := bytes.NewBuffer(make([]byte, 100))
	precread, err := MaxRead(100, prectest)
	if err != nil {
		t.Errorf("Precread failed: %s", err)
	}
	if len(precread) != 100 {
		t.Errorf("Precread short read: %d", len(precread))
	}
	maxtest := bytes.NewBuffer(make([]byte, 100))
	maxread, err := MaxRead(99, maxtest)
	if err == nil {
		t.Error("Maxread MUST failed")
	}
	if len(maxread) != 0 {
		t.Errorf("Maxread long read: %d", len(maxread))
	}
}

func TestBase58(t *testing.T) {
	msg := []byte("testing this to encode/decode")
	encoded := B58encode(msg)
	decoded := B58decode(encoded)
	if !bytes.Equal(msg, decoded) {
		t.Error("B58 encode/decode no match")
	}
}
