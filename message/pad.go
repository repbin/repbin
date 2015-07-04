package message

import (
	"crypto/aes"
	"crypto/cipher"
	"io"

	log "github.com/repbin/repbin/deferconsole"
)

// GenPadKey generates a random value suitable for the padding generator
func GenPadKey() (*[PadKeySize]byte, error) {
	var k [PadKeySize]byte
	_, err := io.ReadFull(randSource, k[:])
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// GenPad produces a slice of bytes with length int, filled with pseudo-random numbers generated from key using AES-CTR
func GenPad(key *[PadKeySize]byte, length int) []byte {
	blockcipher, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err) // this should never happens since the key should always come from GenPadKey
	}
	in := make([]byte, (1+length/aes.BlockSize)*aes.BlockSize)           // we encrypt zeros
	ctrcipher := cipher.NewCTR(blockcipher, make([]byte, aes.BlockSize)) // IV for padding is constant. This is no problem since the key is unique
	ctrcipher.XORKeyStream(in, in)                                       // Ciphertext overwrites cleartext without conflict
	return in[0:length]
}

// RePad adds the deterministic padding to a message
func RePad(msg []byte, padKey *[PadKeySize]byte, totalLength int) []byte {
	msgLen := len(msg)
	padLen := totalLength - msgLen
	log.Secretf("RePad Length: %d\n", padLen)
	if padLen <= 0 {
		return msg
	}
	log.Secretf("RePad Key: %x\n", *padKey)
	ret := append(make([]byte, 0), msg[:msgLen-HMACSize]...) // Message body
	ret = append(ret, GenPad(padKey, padLen)...)             // Deterministic padding
	ret = append(ret, msg[msgLen-HMACSize:]...)              // HMAC
	return ret
}
