package structs

import (
	"testing"

	"github.com/repbin/repbin/message"
)

func TestExpireStructEncode(t *testing.T) {
	td := ExpireStruct{
		ExpireTime:   500,
		MessageID:    [message.MessageIDSize]byte{0xf1, 0x55, 0x71, 0x31},
		SignerPubKey: [message.SignerPubKeySize]byte{0xff, 0xff, 0xf, 0x00},
	}
	enc := td.Encode().Fill()
	if len(enc) != ExpireStructSize {
		t.Error("Fill failed")
	}
	dec := ExpireStructDecode(enc)
	if td.ExpireTime != dec.ExpireTime {
		t.Errorf("Decode ExpireTime failed: %d", dec.ExpireTime)
	}
	if td.MessageID != dec.MessageID {
		t.Errorf("Decode MessageID failed: %x", dec.MessageID)
	}
	if td.SignerPubKey != dec.SignerPubKey {
		t.Errorf("Decode SignerPubKey failed: %x", dec.SignerPubKey)
	}
}
