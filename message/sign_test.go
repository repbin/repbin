package message

import (
	"crypto/sha256"
	"testing"
)

func TestGenKey(t *testing.T) {
	var msgID [sha256.Size]byte
	keypair, err := GenKey(10)
	if err != nil {
		t.Errorf("Key Generation Failed: %s", err)
	}
	sigHeader := keypair.Sign(msgID)
	_ = sigHeader
}

func TestVerifySignature(t *testing.T) {
	msgID := [32]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
		0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02}
	keypair, _ := GenKey(12)
	signHeader := keypair.Sign(msgID)
	sigDetails, err := VerifySignature(*signHeader, 10)
	if err != nil {
		t.Fatalf("Verification error: %s", err)
	}
	if sigDetails.HashCashBits < 12 {
		t.Error("Bit count failed")
	}
	if msgID != sigDetails.MsgID {
		t.Error("MsgID extraction failed")
	}
	_, err = VerifySignature(*signHeader, 32)
	if err == nil {
		t.Fatal("Must fail, too few hashcash bits")
	}
	signHeader[signHeaderMsgIDStart+3] = 0x00
	_, err = VerifySignature(*signHeader, 32)
	if err == nil {
		t.Fatal("Must fail, signature does not match")
	}
}

func TestMarshal(t *testing.T) {
	keypair, _ := GenKey(12)
	m := keypair.Marshal()
	keypair2 := new(SignKeyPair)
	kp, err := keypair2.Unmarshal(m)
	if err != nil {
		t.Fatalf("Unmarshal failed: %s", err)
	}
	if keypair.Bits != kp.Bits {
		t.Error("Unmarshal failed, bits")
	}
	if keypair.Nonce != kp.Nonce {
		t.Error("Unmarshal failed, Nonce")
	}
	if *keypair.PrivateKey != *kp.PrivateKey {
		t.Error("Unmarshal failed, PrivateKey")
	}
	if *keypair.PublicKey != *kp.PublicKey {
		t.Error("Unmarshal failed, PublicKey")
	}
}
