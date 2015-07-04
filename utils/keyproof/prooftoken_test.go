package keyproof

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/agl/ed25519"
)

func TestSignProof(t *testing.T) {
	pubkeyRec, privKeyRec, _ := ed25519.GenerateKey(rand.Reader)
	pubkeySend, privkeySend, _ := ed25519.GenerateKey(rand.Reader)
	now := time.Now().UnixNano()
	proof := SignProofToken(now, pubkeyRec, pubkeySend, privkeySend)
	ok, nowDec, senderPub := VerifyProofToken(proof, pubkeyRec)
	if !ok {
		t.Error("Siganture did not verify")
	}
	if *senderPub != *pubkeySend {
		t.Error("Sender Pubkey decode failed")
	}
	if now != nowDec {
		t.Error("Time decode failed")
	}
	ok, signedProofToken := CounterSignToken(proof, pubkeyRec, privKeyRec)
	if !ok {
		t.Error("CounterSignToken did not verify")
	}
	ok, nowDec2 := VerifyCounterSig(signedProofToken, pubkeySend)
	if !ok {
		t.Error("VerifyCounterSig did not verify")
	}
	if now != nowDec2 {
		t.Error("Time decode failed")
	}
}
