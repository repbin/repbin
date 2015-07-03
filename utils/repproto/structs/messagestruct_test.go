package structs

import (
	"github.com/repbin/repbin/message"
	"testing"
)

func TestMessageStructToBytes(t *testing.T) {
	td := MessageStruct{
		PostTime:               5,
		ExpireTime:             ^uint64(0),
		ExpireRequest:          100,
		MessageID:              [message.MessageIDSize]byte{0x00, 0x01, 0x00, 0x02},
		ReceiverConstantPubKey: message.Curve25519Key{0x01, 0x02, 0x03, 0x04},
		SignerPub:              [message.SignerPubKeySize]byte{0xff, 0x01, 0x00, 0xaa},
		Distance:               3,
		OneTime:                true,
		Sync:                   false,
		Hidden:                 true,
	}
	enc := td.Encode().Fill()
	if len(enc) != MessageStructSize {
		t.Error("Fill failed")
	}
	dec := MessageStructDecode(enc)
	if td.PostTime != dec.PostTime {
		t.Errorf("Posttime decode failed: %d", dec.PostTime)
	}
	if td.ExpireTime != dec.ExpireTime {
		t.Errorf("ExpireTime decode failed: %d", dec.ExpireTime)
	}
	if td.ExpireRequest != dec.ExpireRequest {
		t.Errorf("ExpireRequest decode failed: %d", dec.ExpireRequest)
	}
	if td.Distance != dec.Distance {
		t.Errorf("Distance decode failed: %d", dec.Distance)
	}
	if td.MessageID != dec.MessageID {
		t.Errorf("MessageID decode failed: %x", dec.MessageID)
	}
	if td.ReceiverConstantPubKey != dec.ReceiverConstantPubKey {
		t.Errorf("ReceiverConstantPubKey decode failed: %x", dec.ReceiverConstantPubKey)
	}
	if td.SignerPub != dec.SignerPub {
		t.Errorf("SignerPub decode failed: %x", dec.SignerPub)
	}
	if td.OneTime != dec.OneTime {
		t.Errorf("OneTime decode failed: %t", dec.OneTime)
	}
	if td.Sync != dec.Sync {
		t.Errorf("Sync decode failed: %t", dec.Sync)
	}
	if td.Hidden != dec.Hidden {
		t.Errorf("Hidden decode failed: %t", dec.Hidden)
	}
}
