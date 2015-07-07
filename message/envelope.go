package message

// Envelope: Encode messages in base64, and do selective reading of base64 message

import (
	"encoding/base64"
	"errors"
)

const (
	signHeaderMaxBase64 = 184 // a signatureheader encoded in base64 stdencoding is at most 184 bytes long
	signHeaderMaxBuffer = 138 // base64 encoded data of size 184 contains at most 138 bytes decoded data
)

var (
	// ErrEnvelopeShort is returned if the envelope is too short to even contain a signature header.
	ErrEnvelopeShort = errors.New("message: Envelope too short")
)

// Base64Message is a message encoded in Base64.
type Base64Message []byte

// GetSignHeader returns the signature header from an message.
func (bm Base64Message) GetSignHeader() (*[SignHeaderSize]byte, error) {
	var ret [SignHeaderSize]byte
	if len(bm) < signHeaderMaxBase64 {
		return nil, ErrEnvelopeShort
	}
	dst := make([]byte, signHeaderMaxBuffer)
	_, err := base64.StdEncoding.Decode(dst, bm[:signHeaderMaxBase64])
	if err != nil {
		return nil, err
	}
	copy(ret[:], dst[:SignHeaderSize])
	return &ret, nil
}

// Decode a base64 encoded message.
func (bm Base64Message) Decode() ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(bm)))
	n, err := base64.StdEncoding.Decode(dst, bm)
	if err != nil && n < 2048 { // We only signal errors when decodes are too short. Otherwise we accept some errors
		return nil, err
	}
	return dst[:n], nil
}

// EncodeBase64 encodes a message to base64.
func EncodeBase64(message []byte) Base64Message {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(dst, message)
	return Base64Message(dst)
}
