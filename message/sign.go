package message

import (
	"errors"

	"github.com/agl/ed25519"
	"github.com/repbin/repbin/hashcash"
)

var (
	// ErrNoKeyFound is returned if no key could be found
	ErrNoKeyFound = errors.New("message: No signature key found in generation")
	// ErrBadVersion is returned if a message does not conform to the right version
	ErrBadVersion = errors.New("message: Wrong version")
	// ErrHashCash is returned if not enough bits are produced by the ErrHashCash challenge
	ErrHashCash = errors.New("message: Not enough HashCash bits")
	// ErrBadSignature is returned if the post signature does not verify
	ErrBadSignature = errors.New("message: Signature verification failed")
)

const (
	// SignerPubKeySize is the size of a public key used for signing
	SignerPubKeySize         = ed25519.PublicKeySize
	signHeaderVersionStart   = 0
	signHeaderPubkeyStart    = 1
	signHeaderVersionEnd     = signHeaderPubkeyStart
	signHeaderNonceStart     = signHeaderPubkeyStart + SignerPubKeySize
	signHeaderPubkeyEnd      = signHeaderNonceStart
	signHeaderSignatureStart = signHeaderNonceStart + hashcash.NonceSize
	signHeaderNonceEnd       = signHeaderSignatureStart
	signHeaderMsgIDStart     = signHeaderSignatureStart + ed25519.SignatureSize
	signHeaderSignatureEnd   = signHeaderMsgIDStart
	signHeaderMsgIDEnd       = signHeaderMsgIDStart + MessageIDSize
)

// SignKeyPair represents a signature key pari
type SignKeyPair struct {
	PublicKey  *[SignerPubKeySize]byte
	PrivateKey *[ed25519.PrivateKeySize]byte
	Nonce      [hashcash.NonceSize]byte // HashCash nonce
	Bits       byte                     // Bits in hashcash
}

// GenKey calculates a keypair fit for signing, including hashcahs
func GenKey(bits byte) (keypair *SignKeyPair, err error) {
	pubkey, privkey, err := ed25519.GenerateKey(randSource)
	if err != nil {
		return nil, err
	}
	nonce, ok := hashcash.ComputeNonceSelect(pubkey[:], bits, 0, 0)
	if !ok {
		return nil, ErrNoKeyFound
	}
	sk := SignKeyPair{
		PublicKey:  pubkey,
		PrivateKey: privkey,
		Bits:       bits,
	}
	copy(sk.Nonce[:], nonce)
	return &sk, nil
}

// Marshal a keypair into a byte slice
func (keypair *SignKeyPair) Marshal() []byte {
	r := make([]byte, SignerPubKeySize+ed25519.PrivateKeySize+hashcash.NonceSize+1)
	copy(r[0:SignerPubKeySize], keypair.PublicKey[:])
	copy(r[SignerPubKeySize:SignerPubKeySize+ed25519.PrivateKeySize], keypair.PrivateKey[:])
	copy(r[SignerPubKeySize+ed25519.PrivateKeySize:SignerPubKeySize+ed25519.PrivateKeySize+hashcash.NonceSize], keypair.Nonce[:])
	r[SignerPubKeySize+ed25519.PrivateKeySize+hashcash.NonceSize] = keypair.Bits
	return r
}

// Unmarshal d into keypair
func (keypair SignKeyPair) Unmarshal(d []byte) (*SignKeyPair, error) {
	kp := new(SignKeyPair)
	kp.PublicKey = new([SignerPubKeySize]byte)
	kp.PrivateKey = new([ed25519.PrivateKeySize]byte)
	copy(kp.PublicKey[:], d[0:SignerPubKeySize])
	copy(kp.PrivateKey[:], d[SignerPubKeySize:SignerPubKeySize+ed25519.PrivateKeySize])
	copy(kp.Nonce[:], d[SignerPubKeySize+ed25519.PrivateKeySize:SignerPubKeySize+ed25519.PrivateKeySize+hashcash.NonceSize])
	kp.Bits = d[SignerPubKeySize+ed25519.PrivateKeySize+hashcash.NonceSize]
	msg := []byte("validationtest")
	if ed25519.Verify(kp.PublicKey, msg, ed25519.Sign(kp.PrivateKey, msg)) {
		return kp, nil
	}
	return nil, ErrNoKeyFound
}

// Sign a messageid
func (keypair *SignKeyPair) Sign(msgID [MessageIDSize]byte) *[SignHeaderSize]byte {
	signHeader := new([SignHeaderSize]byte)
	signHeader[0] = byte(Version)
	copy(signHeader[signHeaderPubkeyStart:signHeaderPubkeyEnd], keypair.PublicKey[:])
	copy(signHeader[signHeaderNonceStart:signHeaderNonceEnd], keypair.Nonce[:])
	signature := ed25519.Sign(keypair.PrivateKey, msgID[:])
	copy(signHeader[signHeaderSignatureStart:signHeaderSignatureEnd], signature[:])
	copy(signHeader[signHeaderMsgIDStart:signHeaderMsgIDEnd], msgID[:])
	return signHeader
}

// SignatureDetails contains the fields of the signature header (minus signature itself)
type SignatureDetails struct {
	MsgID         [MessageIDSize]byte      // MsgID of the message (sha256(KeyHeader|Body))
	PublicKey     [SignerPubKeySize]byte   // Public key of signer
	HashCashNonce [hashcash.NonceSize]byte // The HashCash nonce
	HashCashBits  byte                     // HashCash bits
}

// VerifySignature verifies a signature header. It checks if the version, signatures and hashcash are correct
func VerifySignature(header [SignHeaderSize]byte, minbits byte) (details *SignatureDetails, err error) {
	var ok bool
	if header[0] != byte(Version) {
		return nil, ErrBadVersion
	}
	sd := new(SignatureDetails)
	copy(sd.MsgID[:], header[signHeaderMsgIDStart:signHeaderMsgIDEnd])
	copy(sd.PublicKey[:], header[signHeaderPubkeyStart:signHeaderPubkeyEnd])
	copy(sd.HashCashNonce[:], header[signHeaderNonceStart:signHeaderNonceEnd])
	signature := new([ed25519.SignatureSize]byte)
	copy(signature[:], header[signHeaderSignatureStart:signHeaderSignatureEnd])

	ok, sd.HashCashBits = hashcash.TestNonce(sd.PublicKey[:], sd.HashCashNonce[:], minbits)
	if !ok {
		return sd, ErrHashCash
	}
	ok = ed25519.Verify(&sd.PublicKey, sd.MsgID[:], signature)
	if !ok {
		return nil, ErrBadSignature
	}
	return sd, nil
}
