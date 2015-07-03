package message

import (
	"bytes"
	"testing"
)

func TestBodyDef_EncryptBody(t *testing.T) {
	msg1 := []byte("This")                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           // below one block
	msg2 := []byte("This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. This is a rather long message. ") // many block
	keypackSender, _ := GenSenderKeys(nil)
	keypacKPeer, _ := GenReceiveKeys(nil, nil)
	nonce, _ := GenNonce()
	shared1 := CalcSharedSecret(keypackSender, keypacKPeer, nonce, true)
	be := EncryptBodyDef{
		IV:           [IVSize]byte{0x92, 0x90, 0x52, 0x7e, 0xc8, 0x03, 0x29, 0xfb, 0xf7, 0xbc, 0xa4, 0xdf, 0x88, 0x2f, 0xad, 0xb0}, //*GenIV(nil),
		SharedSecret: shared1,
		MessageType:  MsgTypeBlob,
		TotalLength:  2048,
		PadToLength:  300,
	}
	body, err := be.EncryptBody(msg1)
	if err != nil {
		t.Fatalf("Encryption failed: %s", err)
	}
	shared2 := CalcSharedSecret(keypacKPeer, keypackSender, nonce, false)
	bd := DecryptBodyDef{
		IV:           be.IV,
		SharedSecret: shared2,
	}
	tbody := body.Bytes()

	data, msgtype, err := bd.DecryptBody(tbody)
	if err != nil {
		t.Fatalf("Decryption failed: %s", err)
	}
	if !bytes.Equal(data, msg1) {
		t.Error("Decrypted data corrupt")
	}
	//-----
	body, err = be.EncryptBody(msg2)
	if err != nil {
		t.Fatalf("Encryption failed: %s", err)
	}
	tbody = body.Bytes()

	data, msgtype, err = bd.DecryptBody(tbody)
	if err != nil {
		t.Fatalf("Decryption failed: %s", err)
	}
	if !bytes.Equal(data, msg2) {
		t.Error("Decrypted data corrupt")
	}
	if msgtype != MsgTypeBlob {
		t.Error("Decrypted MEssageType wrong")
	}
}

func TestRePad(t *testing.T) {
	totalLength := 240
	msg := []byte("This is a rather long message. This is a rather long message.") // many block
	keypackSender, _ := GenSenderKeys(nil)
	keypacKPeer, _ := GenReceiveKeys(nil, nil)
	nonce, _ := GenNonce()
	shared1 := CalcSharedSecret(keypackSender, keypacKPeer, nonce, true)
	be := EncryptBodyDef{
		IV:           [IVSize]byte{0x92, 0x90, 0x52, 0x7e, 0xc8, 0x03, 0x29, 0xfb, 0xf7, 0xbc, 0xa4, 0xdf, 0x88, 0x2f, 0xad, 0xb0}, //*GenIV(nil),
		SharedSecret: shared1,
		MessageType:  MsgTypeBlob,
		TotalLength:  totalLength,
		PadToLength:  180,
	}

	body, err := be.EncryptBody(msg)
	if err != nil {
		t.Fatalf("Encryption failed: %s", err)
	}
	noPadBody := body.BytesNoPadding()
	padBody := body.Bytes()
	padBody2 := RePad(noPadBody, &body.PadKey, totalLength)
	if !bytes.Equal(padBody, padBody2) {
		t.Error("RePad failed")
	}
}

func TestPackHeader(t *testing.T) {
	keypackSender, _ := GenSenderKeys(nil)
	keypacKPeer, _ := GenReceiveKeys(nil, nil)
	nonce, _ := GenNonce()
	header := PackKeyHeader(keypackSender, keypacKPeer, nonce)
	keypackSender1, keypacKPeer1, nonce1 := ParseKeyHeader(header)
	if *nonce != *nonce1 {
		t.Error("Nonce unpack failed")
	}
	if *keypackSender1.TemporaryPubKey != *keypackSender.TemporaryPubKey {
		t.Error("keypackSender.TemporaryPubKey unpack failed")
	}
	if *keypackSender1.ConstantPubKey != *keypackSender.ConstantPubKey {
		t.Error("keypackSender.ConstantPubKey unpack failed")
	}
	if *keypacKPeer1.TemporaryPubKey != *keypacKPeer.TemporaryPubKey {
		t.Error("keypacKPeer.TemporaryPubKey unpack failed")
	}
	if *keypacKPeer1.ConstantPubKey != *keypacKPeer.ConstantPubKey {
		t.Error("keypacKPeer.ConstantPubKey unpack failed")
	}
}

func TestCalcMessageID(t *testing.T) {
	totalLength := 240
	msg := []byte("This is a rather long message. This is a rather long message.") // many block
	keypackSender, _ := GenSenderKeys(nil)
	keypacKPeer, _ := GenReceiveKeys(nil, nil)
	nonce, _ := GenNonce()
	keypair, _ := GenKey(12)
	shared1 := CalcSharedSecret(keypackSender, keypacKPeer, nonce, true)
	be := EncryptBodyDef{
		IV:           [IVSize]byte{0x92, 0x90, 0x52, 0x7e, 0xc8, 0x03, 0x29, 0xfb, 0xf7, 0xbc, 0xa4, 0xdf, 0x88, 0x2f, 0xad, 0xb0}, //*GenIV(nil),
		SharedSecret: shared1,
		MessageType:  MsgTypeBlob,
		TotalLength:  totalLength,
		PadToLength:  180,
	}

	body, err := be.EncryptBody(msg)
	if err != nil {
		t.Fatalf("Encryption failed: %s", err)
	}
	message := new(Message)
	message.Body = body.Bytes()
	message.Header = PackKeyHeader(keypackSender, keypacKPeer, nonce)
	msgID := message.CalcMessageID()
	message.SignatureHeader = keypair.Sign(*msgID)
	fullMessage := message.Bytes()
	msgID1 := CalcMessageID(fullMessage)
	if *msgID1 != *msgID {
		t.Errorf("MessageID calculation fails: \n\t%x != %x", *msgID1, *msgID)
	}
}
