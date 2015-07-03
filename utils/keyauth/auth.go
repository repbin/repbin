// Package keyauth provides very simple knowledge proof for curve25519 private key belonging to shared public key
package keyauth

import (
	"crypto/sha256"
	"encoding/binary"
	"golang.org/x/crypto/curve25519"
	"time"
)

// Version of this release
const Version = "0.0.1 very alpha"

const (
	// PublicKeySize .
	PublicKeySize = 32
	// PrivateKeySize .
	PrivateKeySize = 32
	// ChallengeSize is the size of a challenge
	ChallengeSize = PublicKeySize + 8
	// AnswerSize is the size of the answer to a challenge
	AnswerSize = PublicKeySize + ChallengeSize
)

// GenTempKey creates a temporary keypair for NOW
func GenTempKey(secret *[PrivateKeySize]byte) (privateKey *[PrivateKeySize]byte, publicKey *[PublicKeySize]byte, challenge *[ChallengeSize]byte) {
	return GenTempKeyTime(uint64(time.Now().Unix()), secret)
}

// GenTempKeyTime generates a temporary keypair
func GenTempKeyTime(time uint64, secret *[PrivateKeySize]byte) (privateKey *[PrivateKeySize]byte, publicKey *[PublicKeySize]byte, challenge *[ChallengeSize]byte) {
	// Convert time to byte array
	var timeb [8]byte
	binary.BigEndian.PutUint64(timeb[:], time)
	return genTempKey(timeb, secret)
}

func genTempKey(time [8]byte, secret *[PrivateKeySize]byte) (privateKey *[PrivateKeySize]byte, publicKey *[PublicKeySize]byte, challenge *[ChallengeSize]byte) {
	var hashIn [PrivateKeySize*2 + 2*8]byte
	var tempPub [PublicKeySize]byte
	var tchallenge [ChallengeSize]byte
	// Generate hashIn: time | secret | time | secret
	copy(hashIn[:8], time[:])
	copy(hashIn[8:PrivateKeySize+8], secret[:])
	copy(hashIn[PrivateKeySize+8:PrivateKeySize+8+8], time[:])
	copy(hashIn[PrivateKeySize+8+8:], secret[:])
	// generate private key hashing hashIn with sha256
	tempPriv := sha256.Sum256(hashIn[:])
	// calculate public key by doing a scalar multiplication with basepoint from curve25519
	curve25519.ScalarBaseMult(&tempPub, &tempPriv)
	// encode challenge: time | public key
	copy(tchallenge[0:8], hashIn[0:8])
	copy(tchallenge[8:], tempPub[:])
	return &tempPriv, &tempPub, &tchallenge
}

// Answer an authentication challenge. Secret is the private key belonging to the public key to be tested
func Answer(challenge *[ChallengeSize]byte, secret *[PrivateKeySize]byte) *[AnswerSize]byte {
	// return challenge|hash(challenge,DH(Secret,challenge.Pub))
	var sharedSecret [PublicKeySize]byte
	var tempPub [PublicKeySize]byte
	var hashIn, answer [AnswerSize]byte
	copy(hashIn[0:ChallengeSize], challenge[:])
	copy(tempPub[:], challenge[8:])
	curve25519.ScalarMult(&sharedSecret, secret, &tempPub)
	copy(hashIn[ChallengeSize:], sharedSecret[:])
	mHash := sha256.Sum256(hashIn[:])
	copy(answer[:ChallengeSize], challenge[:])
	copy(answer[ChallengeSize:], mHash[:])
	return &answer
}

// Verify an answer. secret is the own secret key, testKey is the public key to test ownership on
func Verify(answer *[AnswerSize]byte, secret *[PrivateKeySize]byte, testKey *[PublicKeySize]byte) bool {
	var sharedSecret [PublicKeySize]byte
	var timeb [8]byte
	var inChallenge [ChallengeSize]byte
	var hashIn [AnswerSize]byte
	var inHash [PrivateKeySize]byte
	t1, t2 := false, false

	copy(inChallenge[:], answer[:ChallengeSize]) // Copy full challenge
	copy(timeb[:], inChallenge[:8])              // First 8 byte of challenge are time
	copy(inHash[:], answer[ChallengeSize:])      // Copy hash
	tempPriv, _, outChallenge := genTempKey(timeb, secret)
	if *outChallenge == inChallenge {
		t1 = true
	}
	curve25519.ScalarMult(&sharedSecret, tempPriv, testKey)
	copy(hashIn[0:ChallengeSize], outChallenge[:])
	copy(hashIn[ChallengeSize:], sharedSecret[:])
	outHash := sha256.Sum256(hashIn[:])
	if inHash == outHash {
		t2 = true
	}
	if t1 && t2 {
		return true
	}
	return false
}

// VerifyTime verifies that the answer is still valid
func VerifyTime(answer *[AnswerSize]byte, now, timeRange uint64) bool {
	challengeTime := binary.BigEndian.Uint64(answer[:8])
	// Test too old, too young
	if now+timeRange > challengeTime && now-timeRange < challengeTime {
		return true
	}
	return false
}

// VerifyTimeNow calls VerifyTime for now
func VerifyTimeNow(answer *[AnswerSize]byte, timeRange uint64) bool {
	now := uint64(time.Now().Unix())
	return VerifyTime(answer, now, timeRange)
}
