package message

/*
Message format:
	base64 encoded:
	SignatureHeader:
		Version:    1 byte = 0x01
		PubkeySign: ed25519 public key
		PubkeyCash: hashcash
		Signature:  ed25519 signature over MessageID
		MessageID   hash256(Envelope)
	Body:
		KeyHeader:
			PubKeySendConstant     curve25519 pubkey
			PubKeySendTemporary    curve25519 pubkey
			PubKeyReceiveConstant  curve25519 pubkey
			PubKeyReceiveTemporary curve25519 pubkey
			Nonce 32byte
		EncryptedBody:
			Encrypted (message-key):
				Type: 1 byte. 0x01 == Blob, 0x02 == List, 0x03 == Repost
				Length Content (forget everything beyond, does not include type/length). 2 bytes (65k)
				Content
					Reply PubKeys // can be zero
					List:
						List of MessageID PrivKey\n
					Blob:
						Unformatted data
					Repost:
						PaddingKey. (cut repost content and add padding key before HMAC)
						RepostContent
			RandomPadding
			DeterministicPadding: aes-ctr(0x00,0x00,[]byte(paddinglength),padding-key)
			HMAC(encrypted | padding, hmac-key)
*/

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"

	"github.com/agl/ed25519"
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/hashcash"
)

var (
	// ErrTooLong is returned if the message body is too big
	ErrTooLong = errors.New("message: Content too long")
	// ErrTooShort is returned if the message is smaller than allowed
	ErrTooShort = errors.New("message: Message too short")
	// ErrBadHMAC is returned if the HMAC does not verify
	ErrBadHMAC = errors.New("message: Bad HMAC")
)

const (
	// MsgTypeBlob signals a standard blob of data without special meaning
	MsgTypeBlob = 0x01
	// MsgTypeList signals a list of other messages
	MsgTypeList = 0x02
	// MsgTypeRepost signals a message that is sent to a reposter
	MsgTypeRepost = 0x03
)

const (
	// BodyMaxLength is the maximum length of the body that can be accepted
	BodyMaxLength = 65535
	// MessageIDSize is the size of a MessageID
	MessageIDSize = sha256.Size
	// SignHeaderSize is the size of the signature header
	SignHeaderSize = 1 + ed25519.PublicKeySize + ed25519.SignatureSize + MessageIDSize + hashcash.NonceSize
	// KeyHeaderSize is the size of the key header (four curve25519 public keys for DH and nonce)
	KeyHeaderSize       = 4*Curve25519KeySize + NonceSize
	encryptedHeaderSize = 3
	// Version of this protocol
	Version = 0x01
)

// DecryptBodyDef contains parameters for decryption
type DecryptBodyDef struct {
	IV           [IVSize]byte        // The IV to use. This is overkill since the key is unique anyways... anyways
	SharedSecret [SharedKeySize]byte // The shared secret, calculate HMAC and Symmetric Key from this
}

// EncryptBodyDef contains parameters for body encryption
type EncryptBodyDef struct {
	IV           [IVSize]byte        // The IV to use. This is overkill since the key is unique anyways... anyways
	SharedSecret [SharedKeySize]byte // The shared secret, calculate HMAC and Symmetric Key from this
	TotalLength  int                 // Total size of body including padding
	PadToLength  int                 // PadToLength will pad the body to PadToLength size of random padding before adding deterministic padding, if any
	MessageType  byte
}

// EncryptedBody contains the restuls of an encryption or parsing
type EncryptedBody struct {
	Encrypted            []byte           // First element
	RandomPadding        []byte           // Second element
	DeterministicPadding []byte           // Third element
	PadKey               [PadKeySize]byte // Will be set if deterministic padding is appended
	HMACSum              [HMACSize]byte   // The HMAC
}

// EncryptBody takes data and creates a body out of it
func (bd *EncryptBodyDef) EncryptBody(data []byte) (*EncryptedBody, error) {
	var detSize int
	dataLen := len(data)
	log.Secretf("Data Length: %d\n", dataLen)
	if dataLen > bd.TotalLength-HMACSize-encryptedHeaderSize || dataLen > BodyMaxLength { // HMac is appended, one byte for type, two for length
		return nil, ErrTooLong
	}

	encryptedHeader := make([]byte, aes.BlockSize)
	log.Secretf("Shared Secret: %x\n", bd.SharedSecret)
	hmacKey, symmetricKey := CalcKeys(bd.SharedSecret)
	log.Secretf("HMAC Key: %x\n", *hmacKey)
	log.Secretf("Symmetric Key: %x\n", *symmetricKey)
	eBody := new(EncryptedBody)
	eBody.Encrypted = make([]byte, dataLen+encryptedHeaderSize)
	blockcipher, err := aes.NewCipher(symmetricKey[:])
	if err != nil {
		return nil, err
	}
	ctr := cipher.NewCTR(blockcipher, bd.IV[:aes.BlockSize])
	log.Secretf("Encrypt IV: %x\n", bd.IV[:aes.BlockSize])
	encryptedHeader[0] = bd.MessageType
	binary.BigEndian.PutUint16(encryptedHeader[1:encryptedHeaderSize], uint16(dataLen)) // Length for data

	if len(data) > aes.BlockSize-encryptedHeaderSize {
		// Multiple blocks. Encrypted header first, then remainder
		copy(encryptedHeader[encryptedHeaderSize:], data[0:aes.BlockSize-encryptedHeaderSize])
		ctr.XORKeyStream(eBody.Encrypted[0:aes.BlockSize], encryptedHeader)
		ctr.XORKeyStream(eBody.Encrypted[aes.BlockSize:], data[aes.BlockSize-encryptedHeaderSize:])
	} else {
		// Short block. Only one encryption
		copy(encryptedHeader[encryptedHeaderSize:], data)
		ctr.XORKeyStream(eBody.Encrypted, encryptedHeader[:len(eBody.Encrypted)])
	}

	hmaccalc := hmac.New(sha256.New, hmacKey[:])
	hmaccalc.Write(eBody.Encrypted)

	detSize = dataLen + encryptedHeaderSize + HMACSize
	if bd.PadToLength-detSize > 0 {
		randomPadKey, err := GenPadKey()
		if err != nil {
			return nil, err
		}
		padlength := bd.PadToLength - detSize
		log.Secretf("Random Pad Key: %x\n", *randomPadKey)
		log.Secretf("Random Pad Length: %d\n", padlength)
		eBody.RandomPadding = GenPad(randomPadKey, padlength)
		detSize = bd.PadToLength
		hmaccalc.Write(eBody.RandomPadding)
	}
	if bd.TotalLength > detSize {
		detPadKey, err := GenPadKey()
		if err != nil {
			return nil, err
		}
		padlength := bd.TotalLength - detSize
		log.Secretf("Deterministic Pad Key: %x\n", *detPadKey)
		log.Secretf("Deterministic Pad Length: %d\n", padlength)

		eBody.PadKey = *detPadKey
		eBody.DeterministicPadding = GenPad(detPadKey, bd.TotalLength-detSize)
		hmaccalc.Write(eBody.DeterministicPadding)
	}
	copy(eBody.HMACSum[:], hmaccalc.Sum(nil))
	return eBody, nil
}

// DecryptBody decrypts a body and verifies the hmac
func (bd *DecryptBodyDef) DecryptBody(data []byte) ([]byte, byte, error) {
	var content []byte
	log.Secretf("Shared Secret: %x\n", bd.SharedSecret)
	hmacKey, symmetricKey := CalcKeys(bd.SharedSecret)
	log.Secretf("HMAC Key: %x\n", *hmacKey)
	log.Secretf("Symmetric Key: %x\n", *symmetricKey)
	hmaccalc := hmac.New(sha256.New, hmacKey[:])
	hmaccalc.Write(data[:len(data)-HMACSize])
	hmaccalcSum := hmaccalc.Sum(nil)
	if !hmac.Equal(hmaccalcSum, data[len(data)-HMACSize:]) {
		log.Debugf("Bad HMAC: %x\n", hmaccalcSum)
		return nil, 0, ErrBadHMAC
	}
	blockcipher, err := aes.NewCipher(symmetricKey[:])
	if err != nil {
		return nil, 0, err
	}
	ctr := cipher.NewCTR(blockcipher, bd.IV[:aes.BlockSize])
	// Calculate size of first read. Read one block unless message is too short
	firstRead := aes.BlockSize
	if len(data)-HMACSize < aes.BlockSize {
		firstRead = len(data)
	}
	header := make([]byte, firstRead)
	ctr.XORKeyStream(header, data[:firstRead])
	msgType := header[0]                                             // Decode Message Type
	length := binary.BigEndian.Uint16(header[1:encryptedHeaderSize]) // Decode Message Length
	log.Secretf("Real Length: %d\n", length)
	// Length escapes one block (minus header which is not part of the length)
	if length > aes.BlockSize-encryptedHeaderSize {
		// We have already read aes.BlockSize, but it includes the header
		nlength := length - aes.BlockSize + encryptedHeaderSize
		content = make([]byte, nlength)
		// Decrypt whatever is left
		ctr.XORKeyStream(content, data[aes.BlockSize:aes.BlockSize+nlength])
	}
	//Concat both reads
	content = append(header, content...)
	// Only return after header til end of message
	return content[encryptedHeaderSize : length+encryptedHeaderSize], msgType, nil
}

// Bytes returns the body as byteslice
func (eb EncryptedBody) Bytes() []byte {
	var out []byte
	out = append(out, eb.Encrypted...)
	if eb.RandomPadding != nil && len(eb.RandomPadding) > 0 {
		out = append(out, eb.RandomPadding...)
	}
	if eb.DeterministicPadding != nil && len(eb.DeterministicPadding) > 0 {
		out = append(out, eb.DeterministicPadding...)
	}
	out = append(out, eb.HMACSum[:]...)
	return out
}

// BytesNoPadding returns the body as byteslice omitting deterministic padding
func (eb EncryptedBody) BytesNoPadding() []byte {
	var out []byte
	out = append(out, eb.Encrypted...)
	if eb.RandomPadding != nil && len(eb.RandomPadding) > 0 {
		out = append(out, eb.RandomPadding...)
	}
	out = append(out, eb.HMACSum[:]...)
	return out
}

// PackKeyHeader constructs the header from the KeyPacks and Nonce.
func PackKeyHeader(senderKeys, peerKeys *KeyPack, nonce *[NonceSize]byte) *[KeyHeaderSize]byte {
	var res [KeyHeaderSize]byte
	copy(res[0:Curve25519KeySize], senderKeys.ConstantPubKey[:])
	copy(res[Curve25519KeySize:Curve25519KeySize*2], senderKeys.TemporaryPubKey[:])
	copy(res[Curve25519KeySize*2:Curve25519KeySize*3], peerKeys.ConstantPubKey[:])
	copy(res[Curve25519KeySize*3:Curve25519KeySize*4], peerKeys.TemporaryPubKey[:])
	copy(res[Curve25519KeySize*4:Curve25519KeySize*4+NonceSize], nonce[:])
	return &res
}

// ParseKeyHeader returns the KeyPacks and Nonce of a KeyHeader
func ParseKeyHeader(keyHeader *[KeyHeaderSize]byte) (senderKeys, peerKeys *KeyPack, nonce *[NonceSize]byte) {
	senderKeys, peerKeys = new(KeyPack), new(KeyPack)
	senderKeys.ConstantPubKey = new(Curve25519Key)
	copy(senderKeys.ConstantPubKey[:], keyHeader[0:Curve25519KeySize])
	senderKeys.TemporaryPubKey = new(Curve25519Key)
	copy(senderKeys.TemporaryPubKey[:], keyHeader[Curve25519KeySize:Curve25519KeySize*2])
	peerKeys.ConstantPubKey = new(Curve25519Key)
	copy(peerKeys.ConstantPubKey[:], keyHeader[Curve25519KeySize*2:Curve25519KeySize*3])
	peerKeys.TemporaryPubKey = new(Curve25519Key)
	copy(peerKeys.TemporaryPubKey[:], keyHeader[Curve25519KeySize*3:Curve25519KeySize*4])
	nonce = new([NonceSize]byte)
	copy(nonce[:], keyHeader[Curve25519KeySize*4:Curve25519KeySize*4+NonceSize])
	return
}

// Message is a full message
type Message struct {
	SignatureHeader *[SignHeaderSize]byte // Packet signature header
	Header          *[KeyHeaderSize]byte  // Packed message header
	Body            []byte                // Unpacked body
}

// ParseMessage cuts a byteslice into a Message struct
func ParseMessage(msg []byte) (*Message, error) {
	if len(msg) < SignHeaderSize+KeyHeaderSize+1 {
		return nil, ErrTooShort
	}
	m := &Message{
		SignatureHeader: new([SignHeaderSize]byte),
		Header:          new([KeyHeaderSize]byte),
	}
	copy(m.SignatureHeader[:], msg[0:SignHeaderSize])
	copy(m.Header[:], msg[SignHeaderSize:SignHeaderSize+KeyHeaderSize])
	m.Body = msg[SignHeaderSize+KeyHeaderSize:]
	return m, nil
}

// Bytes converts a message struct into a byte slice
func (msg Message) Bytes() []byte {
	ret := make([]byte, 0, SignHeaderSize+KeyHeaderSize+len(msg.Body))
	ret = append(ret, msg.SignatureHeader[:]...)
	ret = append(ret, msg.Header[:]...)
	ret = append(ret, msg.Body...)
	return ret
}

// CalcMessageID from encrypted body
func (msg Message) CalcMessageID() *[MessageIDSize]byte {
	var ret [MessageIDSize]byte
	h := sha256.New()
	h.Write(msg.Header[:])
	h.Write(msg.Body)
	t := h.Sum(make([]byte, 0))
	copy(ret[:], t)
	return &ret
}

// CalcMessageID from raw message (skip signature part)
func CalcMessageID(msg []byte) *[MessageIDSize]byte {
	var ret [MessageIDSize]byte
	h := sha256.New()
	h.Write(msg[SignHeaderSize:]) // Skip Signature Header
	t := h.Sum(make([]byte, 0))
	copy(ret[:], t)
	return &ret
}
