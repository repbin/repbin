package structs

import (
	"testing"

	"github.com/repbin/repbin/hashcash"
	"github.com/repbin/repbin/message"
)

func TestSignerStructEncode(t *testing.T) {
	ss := SignerStruct{
		PublicKey:           [message.SignerPubKeySize]byte{0x00, 0x01, 0xff, 0xaf},
		Nonce:               [hashcash.NonceSize]byte{0x00, 0x01, 0xff, 0xaf, 0x00, 0x03, 0xaa, 0xc0},
		Bits:                35,
		MessagesPosted:      1,
		MessagesRetained:    3,
		MaxMessagesPosted:   200,
		MaxMessagesRetained: ^uint64(0),
		ExpireTarget:        20,
	}
	enc := ss.Encode().Fill()
	if len(enc) != SignerStructSize {
		t.Error("Fill failed")
	}
	st := SignerStructDecode(enc)
	if ss.PublicKey != st.PublicKey {
		t.Errorf("Decode Publickey failed: %x", st.PublicKey)
	}
	if ss.Nonce != st.Nonce {
		t.Errorf("Decode Nonce failed: %x", st.Nonce)
	}
	if ss.Bits != st.Bits {
		t.Errorf("Decode Bits failed: %d", st.Bits)
	}
	if ss.MessagesPosted != st.MessagesPosted {
		t.Errorf("Decode MessagesPosted failed: %d", st.MessagesPosted)
	}
	if ss.MessagesRetained != st.MessagesRetained {
		t.Errorf("Decode MessagesRetained failed: %d", st.MessagesRetained)
	}
	if ss.MaxMessagesPosted != st.MaxMessagesPosted {
		t.Errorf("Decode MaxMessagesPosted failed: %d", st.MaxMessagesPosted)
	}
	if ss.MaxMessagesRetained != st.MaxMessagesRetained {
		t.Errorf("Decode MaxMessagesRetained failed: %d", st.MaxMessagesRetained)
	}
	if ss.ExpireTarget != st.ExpireTarget {
		t.Errorf("Decode ExpireTarget failed: %d", st.ExpireTarget)
	}
}
