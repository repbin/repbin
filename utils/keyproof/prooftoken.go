// Package keyproof implements a simple authentication system based on ed25519 signatures
// Alice sends a token to Bob that authenticates Alice to Bob and contains time information
// Bob then verifies and counter-signs the token and sends it back to Alice
// Alice verifies the original token, the counter-signature and the timing information and grants access or denies it
package keyproof

import (
	"encoding/binary"

	"github.com/agl/ed25519"
)

// Version of this release
const Version = "0.0.1 very alpha"

// ProofTokenSize is the length of a proof token
const ProofTokenSize = ed25519.PublicKeySize + 8 + ed25519.PublicKeySize + ed25519.SignatureSize

// ProofTokenMax is the maximum size of a base58 encoded ProofToken
const ProofTokenMax = 186

// ProofTokenSignedSize is the length of a counter-signed prooftoken
const ProofTokenSignedSize = ProofTokenSize + ed25519.SignatureSize

// ProofTokenSignedMax is the maximum size of a base58 encoded signed prooftoken
const ProofTokenSignedMax = 274

// SignProofToken generates a signproof
// recPubkey is Bob's public key, sendPubkey and sendPrivkey are Alice's. time is the timestamp to embed
// returns the prooftoken
func SignProofToken(time int64, recPubkey, sendPubkey *[ed25519.PublicKeySize]byte, sendPrivKey *[ed25519.PrivateKeySize]byte) *[ProofTokenSize]byte {
	var proof [ProofTokenSize]byte
	timeB := make([]byte, 8)
	binary.BigEndian.PutUint64(timeB, uint64(time))
	cur := 0
	copy(proof[cur:cur+8], timeB[:])
	cur += 8
	copy(proof[cur:cur+ed25519.PublicKeySize], recPubkey[:])
	cur += ed25519.PublicKeySize
	copy(proof[cur:cur+ed25519.PublicKeySize], sendPubkey[:])
	cur += ed25519.PublicKeySize
	sig := ed25519.Sign(sendPrivKey, proof[:cur])
	copy(proof[cur:cur+ed25519.SignatureSize], sig[:])
	return &proof
}

// VerifyProofToken verifies a proof and returns the sender's public key
// recPubkeyTest is Bob's public key. Returns true/false, the embeded timestamp and Alice's pubkey
func VerifyProofToken(proof *[ProofTokenSize]byte, recPubkeyTest *[ed25519.PublicKeySize]byte) (bool, int64, *[ed25519.PublicKeySize]byte) {
	var senderPub, recPub [ed25519.PublicKeySize]byte
	var sig [ed25519.SignatureSize]byte
	var time int64
	cur := 0
	time = int64(binary.BigEndian.Uint64(proof[cur : cur+8]))
	cur += 8
	copy(recPub[:], proof[cur:cur+ed25519.PublicKeySize])
	cur += ed25519.PublicKeySize
	copy(senderPub[:], proof[cur:cur+ed25519.PublicKeySize])
	cur += ed25519.PublicKeySize
	copy(sig[:], proof[cur:cur+ed25519.SignatureSize])
	ok := ed25519.Verify(&senderPub, proof[:cur], &sig)
	if ok {
		if *recPubkeyTest == recPub {
			return true, time, &senderPub
		}
	}
	return false, 0, nil
}

// CounterSignToken creates a counter-signature for token
// recPubKey is the public key of Bob. Returns true/false and the countersigned prooftoken
func CounterSignToken(proof *[ProofTokenSize]byte, recPubKey *[ed25519.PublicKeySize]byte, recPrivKey *[ed25519.PrivateKeySize]byte) (bool, *[ProofTokenSignedSize]byte) {
	var counterSig [ProofTokenSignedSize]byte
	ok, _, _ := VerifyProofToken(proof, recPubKey)
	if !ok {
		return false, nil
	}
	copy(counterSig[:ProofTokenSize], proof[:ProofTokenSize])
	sig := ed25519.Sign(recPrivKey, proof[:ProofTokenSize])
	copy(counterSig[ProofTokenSize:], sig[:ed25519.SignatureSize])
	return true, &counterSig
}

// VerifyCounterSig verifies that a counter signature is valid
// sendPubkey is the public key of Alice. Returns true/false and embedded timestamp
func VerifyCounterSig(counterSig *[ProofTokenSignedSize]byte, sendPubkey *[ed25519.PublicKeySize]byte) (bool, int64) {
	var proof [ProofTokenSize]byte
	var receiverPub [ed25519.PublicKeySize]byte
	var sig [ed25519.SignatureSize]byte
	copy(proof[:], counterSig[:ProofTokenSize])
	copy(receiverPub[:], proof[8:8+ed25519.PublicKeySize])
	copy(sig[:], counterSig[ProofTokenSize:ProofTokenSize+ed25519.SignatureSize])
	ok, time, senderPubTest := VerifyProofToken(&proof, &receiverPub)
	if !ok {
		return false, 0
	}
	ok = ed25519.Verify(&receiverPub, proof[:], &sig)
	if !ok {
		return false, 0
	}
	if *sendPubkey != *senderPubTest {
		return false, 0
	}
	return true, time
}
