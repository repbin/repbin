package message

import (
	"testing"
)

func TestGenSenderKeys(t *testing.T) {
	_, errx := GenLongTermKey(false, false)
	if errx != nil {
		t.Fatalf("Key Generation failed: %s", errx)
	}
	_, err := GenSenderKeys(nil)
	if err != nil {
		t.Fatalf("Key Generation failed: %s", err)
	}
	keypackSender, _ := GenSenderKeys(nil)
	keypacKPeer, _ := GenReceiveKeys(nil, nil)
	nonce, _ := GenNonce()
	shared1 := CalcSharedSecret(keypackSender, keypacKPeer, nonce, true)
	shared2 := CalcSharedSecret(keypacKPeer, keypackSender, nonce, false)
	if shared1 != shared2 {
		t.Error("Key agreement failed Shared")
	}
	hmac1, symmetric1 := CalcKeys(shared1)
	hmac2, symmetric2 := CalcKeys(shared2)
	if *hmac1 != *hmac2 {
		t.Error("Key agreement failed HMAC")
	}
	if *symmetric1 != *symmetric2 {
		t.Error("Key agreement failed SYMM")
	}
}

func TestMatchPrivate(t *testing.T) {
	priv, err := GenLongTermKey(true, true)
	if err != nil {
		t.Fatalf("Key generation failed: %s", err)
	}
	priv2, _ := GenLongTermKey(true, false)
	_ = priv2
	pub, pub2 := new(Curve25519Key), new(Curve25519Key)
	scalarBaseMult(pub, priv)
	scalarBaseMult(pub2, priv2)
	keypackSender, err := GenReceiveKeys(pub, nil)
	if err == nil {
		t.Fatal("Key packing MUST fail on missing temporary pubkey if constant pubkey is present")
	}
	keypackSender, err = GenReceiveKeys(pub, pub2)
	if err != nil {
		t.Fatalf("Key packing failed: %s", err)
	}
	ok := keypackSender.MatchPrivate(priv, priv2)
	if !ok {
		t.Error("Match failed")
	}
	keypackSender, err = GenReceiveKeys(nil, nil)
	if err != nil {
		t.Fatalf("Key packing failed: %s", err)
	}
	ok = keypackSender.MatchPrivate(keypackSender.ConstantPrivKey, nil)
	if !ok {
		t.Error("Match failed")
	}
	_ = keypackSender
}

func TestGenLongterm(t *testing.T) {
	priv, _ := GenLongTermKey(true, false)
	pub := new(Curve25519Key)
	scalarBaseMult(pub, priv)
	if KeyIsSync(pub) || !KeyIsHidden(pub) {
		t.Error("Sync/Hidden bad 1")
	}
	priv, _ = GenLongTermKey(false, true)
	scalarBaseMult(pub, priv)
	if !KeyIsSync(pub) || KeyIsHidden(pub) {
		t.Error("Sync/Hidden bad 2")
	}
}
