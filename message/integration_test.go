package message

import (
	"bytes"
	log "github.com/repbin/repbin/deferconsole"
	"testing"
)

func TestEncryptDecryptAll(t *testing.T) {
	log.SetMinLevel(log.LevelError)
	msg := []byte("This is a small test message for verification, it just has to be not too short to be not boring")
	sender := new(Sender) // Default sender
	msgEnc, meta, err := sender.Encrypt(1, msg)
	if err != nil {
		t.Fatalf("Encryption failed: %s", err)
	}
	receiver := &Receiver{ // Temporary message key from meta
		ReceiveConstantPrivateKey: meta.MessageKey,
	}
	message, metaRec, err := receiver.Decrypt(msgEnc)
	if err != nil {
		t.Fatalf("Decryption failed: %s", err)
	}
	if metaRec.MessageID != meta.MessageID {
		t.Error("MessageIDs do not match")
	}
	if metaRec.MessageType != 1 {
		t.Error("Message type does not match")
	}
	if !bytes.Equal(msg, message) {
		t.Error("Message corrupted")
	}
}

func TestEncryptDecryptSenderKey(t *testing.T) {
	//  - with senderkey
	log.SetMinLevel(log.LevelError)
	msg := []byte("This is a small test message for verification, it just has to be not too short to be not boring")
	senderkey, _ := GenLongTermKey(false, false)
	senderkey2, _ := GenLongTermKey(false, false)
	sender := &Sender{
		SenderPrivateKey: senderkey,
	}
	msgEnc, meta, err := sender.Encrypt(0, msg)
	if err != nil {
		t.Fatalf("Encryption failed: %s", err)
	}
	receiver := &Receiver{ // Temporary message key from meta
		ReceiveConstantPrivateKey: meta.MessageKey,
		SenderPublicKey:           CalcPub(senderkey),
	}
	message, metaRec, err := receiver.Decrypt(msgEnc)
	if err != nil {
		t.Fatalf("Decryption failed: %s", err)
	}
	if metaRec.MessageID != meta.MessageID {
		t.Error("MessageIDs do not match")
	}
	if metaRec.MessageType != 0 {
		t.Error("Message type does not match")
	}
	if !bytes.Equal(msg, message) {
		t.Error("Message corrupted")
	}
	receiver = &Receiver{ // Temporary message key from meta
		ReceiveConstantPrivateKey: meta.MessageKey,
		SenderPublicKey:           CalcPub(senderkey2),
	}
	_, _, err = receiver.Decrypt(msgEnc)
	if err == nil {
		t.Errorf("Should fail for wrong sender")
	}
}

func TestEncryptDecryptReceiver(t *testing.T) {
	log.SetMinLevel(log.LevelError)
	msg := []byte("This is a small test message for verification, it just has to be not too short to be not boring")
	senderPrivKeyConstant, _ := GenLongTermKey(false, false)
	senderPrivKeyTemporary, _ := GenLongTermKey(false, false)
	sender := &Sender{
		ReceiveConstantPublicKey:  CalcPub(senderPrivKeyConstant),
		ReceiveTemporaryPublicKey: CalcPub(senderPrivKeyTemporary),
	}
	msgEnc, meta, err := sender.Encrypt(0, msg)
	if err != nil {
		t.Fatalf("Encryption failed: %s", err)
	}
	receiver := &Receiver{ // Temporary message key from meta
		ReceiveConstantPrivateKey:  senderPrivKeyConstant,
		ReceiveTemporaryPrivateKey: senderPrivKeyTemporary,
	}
	message, metaRec, err := receiver.Decrypt(msgEnc)
	if err != nil {
		t.Fatalf("Decryption failed: %s", err)
	}
	if metaRec.MessageID != meta.MessageID {
		t.Error("MessageIDs do not match")
	}
	if metaRec.MessageType != 0 {
		t.Error("Message type does not match")
	}
	if !bytes.Equal(msg, message) {
		t.Error("Message corrupted")
	}
}

func TestEncryptDecryptCallback(t *testing.T) {
	log.SetMinLevel(log.LevelError)
	keys := make(map[Curve25519Key]Curve25519Key)
	msg := []byte("This is a small test message for verification, it just has to be not too short to be not boring")
	senderPrivKeyConstant, _ := GenLongTermKey(false, false)
	senderPubKeyConstant := CalcPub(senderPrivKeyConstant)
	keys[*senderPubKeyConstant] = *senderPrivKeyConstant
	senderPrivKeyTemporary, _ := GenLongTermKey(false, false)
	senderPubKeyTemporary := CalcPub(senderPrivKeyTemporary)
	keys[*senderPubKeyTemporary] = *senderPrivKeyTemporary

	// Callback function for key lookup
	callback := func(key *Curve25519Key) *Curve25519Key {
		pub := keys[*key]
		return &pub
	}

	sender := &Sender{
		ReceiveConstantPublicKey:  senderPubKeyConstant,
		ReceiveTemporaryPublicKey: senderPubKeyTemporary,
	}
	msgEnc, meta, err := sender.Encrypt(0, msg)
	if err != nil {
		t.Fatalf("Encryption failed: %s", err)
	}
	receiver := &Receiver{ // Temporary message key from meta
		KeyCallBack: callback,
	}
	message, metaRec, err := receiver.Decrypt(msgEnc)
	if err != nil {
		t.Fatalf("Decryption failed: %s", err)
	}
	if metaRec.MessageID != meta.MessageID {
		t.Error("MessageIDs do not match")
	}
	if metaRec.MessageType != 0 {
		t.Error("Message type does not match")
	}
	if !bytes.Equal(msg, message) {
		t.Error("Message corrupted")
	}
}

func TestEncryptDecryptRewrap(t *testing.T) {
	log.SetMinLevel(log.LevelError)
	msg := []byte("This is a small test message for verification, it just has to be not too short to be not boring")
	sender := new(Sender) // Default sender
	msgPrePad, meta, err := sender.EncryptRepost(0, msg)
	if err != nil {
		t.Fatalf("Encryption failed: %s", err)
	}
	msgPreEnc := RePad(msgPrePad, meta.PadKey, DefaultTotalLength)
	msgEnc := EncodeBase64(msgPreEnc)
	receiver := &Receiver{ // Temporary message key from meta
		ReceiveConstantPrivateKey: meta.MessageKey,
	}
	_ = msgEnc
	message, metaRec, err := receiver.Decrypt(msgEnc)
	if err != nil {
		t.Fatalf("Decryption failed rewrap: %s", err)
	}
	if metaRec.MessageID != meta.MessageID {
		t.Error("MessageIDs do not match")
	}
	if metaRec.MessageType != 0 {
		t.Error("Message type does not match")
	}
	if !bytes.Equal(msg, message) {
		t.Error("Message corrupted")
	}
}
