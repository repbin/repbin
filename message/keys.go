package message

/*
Keys:
	PreKey_1 := DH(KeySendConstant,KeyReceiveTemporary)
	PreKey_2 := DH(KeySendTemporary,KeyReceiveTemporary)
	PreKey_3 := DH(KeySendTemporary,KeyReceiveConstant)
	PreKey := hash512(PreKey_1 | PreKey_2 | PreKey_3 | Nonce)
	MessageKey := hmac("MessageKey",PreKey)
	HmacKey := hmac("HMACKey",PreKey)
	paddingKey := random

KeySend/Receive:
	Generate on the fly if not given.
	KeySendTemporary is ALWAYS random. If KeySendConstant is given, write KeySendTemporary to stderr

GenHeaderKeys(SendConstant,ReceiveConstant,ReceiveTemporary):
	If ReceiveConstant given:
		ReceiveTemporary may not be nil
	Else:
		ReceiveConstant = random
		ReceiveTemporary = hash(ReceiveConstant)
	If SendConstant == nil
		SendConstant = Random
	SendTemporary = random
*/

import (
	"crypto/aes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"golang.org/x/crypto/curve25519"
	"io"
	log "github.com/repbin/repbin/deferconsole"
)

var (
	// ErrIncompleteKeyPack is returned if a KeyPack cannot be filled
	ErrIncompleteKeyPack = errors.New("message: incomplete keypack")
	// ErrMissingKey is returned if too few keys are used fro keypack construction
	ErrMissingKey = errors.New("message: missing temporary pubkey")
)

const (
	// Curve25519KeySize defines how big a curve25519 key is
	Curve25519KeySize = 32
	// SharedKeySize defines how long the shared secret is
	SharedKeySize = 64
	// HMACKeySize is the size of the HMAC key
	HMACKeySize = 32
	// SymmetricKeySize is the size of the symmetric encryption key (for AES-CTR)
	SymmetricKeySize = 32
	// NonceSize bytes of nonce
	NonceSize = 32
	// PadKeySize is the size of the padding key
	PadKeySize = 32
	// HMACSize is the size of the hmac output, we're using SHA256
	HMACSize = 32
	// IVSize defines the length of the IV, calculated by hashing pubkeys and nonce
	IVSize = aes.BlockSize
)

var (
	// sha256("Repbin HMAC Key")
	hmacGen = [32]byte{0x4a, 0x6b, 0xd5, 0x58, 0x9f, 0x3c, 0x62, 0x8a, 0xe4, 0x04, 0x05, 0xef, 0xfd, 0x64, 0x12, 0xd8, 0x7f, 0x6f, 0xf2, 0x34, 0x0a, 0x92, 0xbe, 0x74, 0xda, 0xee, 0x2c, 0x07, 0x46, 0xbd, 0xd8, 0x20}
	// sha256("Repbin Symmetric Key")
	symmGen = [32]byte{0x1a, 0xb1, 0xc7, 0x9b, 0x15, 0xfb, 0x49, 0xdb, 0x4f, 0xa0, 0x1f, 0x1d, 0x2d, 0x2d, 0x09, 0xa7, 0xa9, 0xb7, 0x83, 0xbf, 0xb2, 0x78, 0xaa, 0x49, 0x86, 0xc0, 0x14, 0x0d, 0x48, 0x71, 0xa8, 0x99}
	// Key fingerprint of "Replicated Encrypted Pastebin" 0xDCA2BA86 OpenPGP key (repeated)
	determGen = [32]byte{0x0E, 0x5C, 0xA1, 0x97, 0xD5, 0x63, 0xFC, 0xCB, 0x90, 0xC0, 0xBF, 0x10, 0x29, 0x75, 0x8D, 0xE1, 0xDC, 0xA2, 0xBA, 0x86, 0x0E, 0x5C, 0xA1, 0x97, 0xD5, 0x63, 0xFC, 0xCB, 0x90, 0xC0, 0xBF, 0x10}
	syncbit   = byte(0x40)
	hiddenbit = byte(0x80)
)

var randSource = rand.Reader

// Curve25519Key is a curve25519 public or private key
type Curve25519Key [Curve25519KeySize]byte

// KeyPack contains one side of keys
type KeyPack struct {
	ConstantPubKey   *Curve25519Key // Public key: SenderConstantPubKey
	ConstantPrivKey  *Curve25519Key // Private key of ConstantPubKey
	TemporaryPubKey  *Curve25519Key // Public key: SenderTemporaryPubKey
	TemporaryPrivKey *Curve25519Key // Private key of TemporaryPubKey
}

// scalarBaseMult is a wrapper around curve25519.ScalarBaseMult with type conversion
func scalarBaseMult(dst, in *Curve25519Key) {
	curve25519.ScalarBaseMult((*[Curve25519KeySize]byte)(dst), (*[Curve25519KeySize]byte)(in))
}

func scalarMult(dst, in, base *Curve25519Key) {
	curve25519.ScalarMult((*[Curve25519KeySize]byte)(dst), (*[Curve25519KeySize]byte)(in), (*[Curve25519KeySize]byte)(base))
}

// FillKeys constructs missing keys in a KeyPack. If deterministic==true then temporary keys are generated from the constant keys
func (kp *KeyPack) FillKeys(deterministic bool) error {
	// Create private keys if constant public key is nil
	if kp.ConstantPrivKey == nil && kp.ConstantPubKey == nil {
		var err error
		kp.ConstantPrivKey = new(Curve25519Key)
		kp.ConstantPrivKey, err = GenLongTermKey(false, true)
		//_, err := io.ReadFull(randSource, kp.ConstantPrivKey[:])
		if err != nil {
			return err
		}
		// Temporary keys are always generated if constant key was missing
		kp.TemporaryPrivKey = nil
		kp.TemporaryPubKey = nil
	}
	if kp.TemporaryPrivKey == nil && kp.TemporaryPubKey == nil {
		kp.TemporaryPrivKey = new(Curve25519Key)
		if !deterministic {
			_, err := io.ReadFull(randSource, kp.TemporaryPrivKey[:])
			if err != nil {
				return err
			}
		} else {
			pka := [Curve25519KeySize]byte(*kp.ConstantPrivKey)
			sum := Curve25519Key(sha256.Sum256(append(pka[:], determGen[:]...)))
			kp.TemporaryPrivKey = &sum
		}
	}
	// Fill in public keys
	if kp.ConstantPubKey == nil && kp.ConstantPrivKey != nil {
		kp.ConstantPubKey = new(Curve25519Key)
		scalarBaseMult(kp.ConstantPubKey, kp.ConstantPrivKey)
	}
	if kp.TemporaryPubKey == nil && kp.TemporaryPrivKey != nil {
		kp.TemporaryPubKey = new(Curve25519Key)
		scalarBaseMult(kp.TemporaryPubKey, kp.TemporaryPrivKey)
	}
	if kp.TemporaryPubKey == nil || kp.ConstantPubKey == nil {
		return ErrIncompleteKeyPack
	}
	return nil
}

// GenKeyPack generate a keypack. If privateConstant is nil, all keys will be ephemeral.
// If deterministic is true then the temporaryPrivKey will be generated from the constantPrivKey
func GenKeyPack(privateConstant *Curve25519Key, deterministic bool) (*KeyPack, error) {
	kp := new(KeyPack)
	kp.ConstantPrivKey = privateConstant
	err := kp.FillKeys(deterministic)
	if err != nil {
		return nil, err
	}
	return kp, nil
}

// GenSenderKeys returns a keypack suitable for the sender. If PrivateConstant is nil, all keys will be ephemeral
func GenSenderKeys(privateConstant *Curve25519Key) (*KeyPack, error) {
	return GenKeyPack(privateConstant, false)
}

// GenLongTermKey returns a private key that is meant for long-term use and adheres to the hidden index rule
// hidden means that the server should not allow index access without authentication
// sync means that the server should not sync the message to other servers (it is not part of the global index)
func GenLongTermKey(hidden bool, sync bool) (*Curve25519Key, error) {
	var priv, pub Curve25519Key
	for {
		_, err := io.ReadFull(randSource, priv[:])
		if err != nil {
			return nil, err
		}
		scalarBaseMult(&pub, &priv)
		if (pub[0]&hiddenbit == hiddenbit) == hidden {
			if (pub[0]&syncbit == syncbit) == sync {
				return &priv, nil
			}
		}
	}
}

// GenRandomKey generates a random key useable for temporary keys
func GenRandomKey() (*Curve25519Key, error) {
	var priv Curve25519Key
	_, err := io.ReadFull(randSource, priv[:])
	if err != nil {
		return nil, err
	}
	return &priv, nil
}

// GenPubKey calculates the public key for a private key
func GenPubKey(priv *Curve25519Key) *Curve25519Key {
	pub := new(Curve25519Key)
	scalarBaseMult(pub, priv)
	return pub
}

// KeyIsHidden returns true if key index should be hidden
func KeyIsHidden(k *Curve25519Key) bool {
	return k[0]&hiddenbit == hiddenbit
}

// KeyIsSync returns true if messages for this key should be synced
func KeyIsSync(k *Curve25519Key) bool {
	return k[0]&syncbit == syncbit
}

// CalcPub generates a public key from a private key
func CalcPub(privateKey *Curve25519Key) *Curve25519Key {
	var pub Curve25519Key
	scalarBaseMult(&pub, privateKey)
	return &pub
}

// GenReceiveKeys generates a KeyPack for the recipient unless public keys are available
func GenReceiveKeys(publicConstant, publicTemporary *Curve25519Key) (*KeyPack, error) {
	if publicConstant == nil {
		return GenKeyPack(nil, true)
	}
	if publicTemporary == nil {
		return nil, ErrMissingKey
	}
	keys := new(KeyPack)
	keys.ConstantPubKey = publicConstant
	keys.TemporaryPubKey = publicTemporary
	return keys, nil
}

// GenNonce returns NonceSize random bytes
func GenNonce() (*[NonceSize]byte, error) {
	nonce := new([NonceSize]byte)
	_, err := io.ReadFull(randSource, nonce[:])
	if err != nil {
		return nil, err
	}
	return nonce, nil
}

// CalcSharedSecret does a triple DH from the keypacks to generate the shared secret. myKeys needs to contain private keys, peerKeys only needs public keys
// Sending determines if one is sender or recipient of a message
func CalcSharedSecret(myKeys, peerKeys *KeyPack, nonce *[NonceSize]byte, sending bool) (sharedSecret [SharedKeySize]byte) {
	preKey := make([]byte, 3*Curve25519KeySize+NonceSize) // three keys plus nonce
	key1, key2, key3 := new(Curve25519Key), new(Curve25519Key), new(Curve25519Key)
	log.Secretf("TemporaryPrivKey: %x\n", *myKeys.TemporaryPrivKey)
	log.Secretf("ConstantPrivKey: %x\n", *myKeys.ConstantPrivKey)
	log.Secretf("TemporaryPubKey: %x\n", *peerKeys.TemporaryPubKey)
	log.Secretf("ConstantPubKey: %x\n", *peerKeys.ConstantPubKey)
	scalarMult(key1, myKeys.TemporaryPrivKey, peerKeys.TemporaryPubKey)
	log.Secretf("Key1: %x\n", *key1)
	scalarMult(key2, myKeys.ConstantPrivKey, peerKeys.TemporaryPubKey)
	scalarMult(key3, myKeys.TemporaryPrivKey, peerKeys.ConstantPubKey)
	preKey = append(preKey, key1[:]...)
	if sending {
		log.Secretf("Key2: %x\n", *key2)
		log.Secretf("Key3: %x\n", *key3)
		preKey = append(preKey, key2[:]...)
		preKey = append(preKey, key3[:]...)
	} else { // Swap for receiver
		log.Secretf("Key2: %x\n", *key3)
		log.Secretf("Key3: %x\n", *key2)
		preKey = append(preKey, key3[:]...)
		preKey = append(preKey, key2[:]...)

	}
	log.Secretf("Nonce: %x\n", *nonce)
	preKey = append(preKey, nonce[:]...)
	return sha512.Sum512(preKey)
}

// CalcKeys generates the keys for hmac calculation and symmetric encryption from the shared secret
func CalcKeys(sharedSecret [SharedKeySize]byte) (hmacKey *[HMACKeySize]byte, symmetricKey *[SymmetricKeySize]byte) {
	var hmacres, symmres []byte
	hmaccalc := hmac.New(sha256.New, sharedSecret[:])
	hmaccalc.Write(hmacGen[:])
	hmacres = hmaccalc.Sum(hmacres)

	symmcalc := hmac.New(sha256.New, sharedSecret[:])
	symmcalc.Write(symmGen[:])
	symmres = symmcalc.Sum(symmres)

	hmacKey = new([HMACKeySize]byte)
	symmetricKey = new([SymmetricKeySize]byte)

	copy(hmacKey[:], hmacres)
	copy(symmetricKey[:], symmres)
	return
}

// GenIV generates the IV by sha256 d. d should be the slice containing the pubkeys and nonce from the message
func GenIV(d []byte) *[IVSize]byte {
	var iv [IVSize]byte
	if d == nil { // this guarantees that the message will never be decrypted outside of tests
		_, err := io.ReadFull(randSource, iv[:])
		if err != nil {
			panic(err) // This shouldnt happen in production anyways
		}
		return &iv
	}
	t := sha256.Sum256(d)
	copy(iv[:], t[:])
	return &iv
}

// MatchPrivate tries to match given private key(s) to the KeyPack. Keys will be added to the keypack if a match is found
func (kp *KeyPack) MatchPrivate(constantPrivKey, temporaryPrivKey *Curve25519Key) bool {
	var constantPubKey, temporaryPubKey Curve25519Key
	scalarBaseMult(&constantPubKey, constantPrivKey)
	if temporaryPrivKey == nil {
		// try deterministic fill
		pka := [Curve25519KeySize]byte(*constantPrivKey)
		sum := Curve25519Key(sha256.Sum256(append(pka[:], determGen[:]...)))
		temporaryPrivKey = &sum
	}
	scalarBaseMult(&temporaryPubKey, temporaryPrivKey)
	if *kp.ConstantPubKey == constantPubKey && *kp.TemporaryPubKey == temporaryPubKey {
		kp.ConstantPrivKey = constantPrivKey
		kp.TemporaryPrivKey = temporaryPrivKey
		return true
	}
	return false
}
