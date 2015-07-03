package message

import (
	"encoding/hex"
	"testing"
)

func TestGenPadKey(t *testing.T) {
	pk, err := GenPadKey()
	if err != nil {
		t.Errorf("GenPadKey error: %s", err)
	}
	_ = pk
}

func TestGenPad(t *testing.T) {
	testDat := "61dd15ecbbd8cebb9fbd81a6b5a5649479cc77f500eb0c912ae6b72102a29f24d4661aea1bdc0c463abb2b5a16d07d46690122071e5cec477a37c8b18cad7d2fe82ce53f67321b8af6f85351eeb6af90a9150b9f754121dc2f2b669e1a3773cbb5fe3344"
	pad := GenPad(&[PadKeySize]byte{0x01, 0x01, 0x02, 0x00, 0x01, 0x02, 0x00, 0x01, 0x02, 0x00, 0x01, 0x02, 0x00, 0x01, 0x02, 0x00, 0x01, 0x02, 0x00, 0x01, 0x02, 0x00, 0x01, 0x02, 0x00, 0x01, 0x02, 0x00, 0x01}, 100)
	if len(pad) != 100 {
		t.Error("Padding length bad")
	}
	if testDat != hex.EncodeToString(pad) {
		t.Error("Bad padding content produced")
	}
}
