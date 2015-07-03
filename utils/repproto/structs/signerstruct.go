package structs

import (
	"bytes"
	"github.com/repbin/repbin/hashcash"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"strconv"
)

const (
	// SignerStructSize is the maximum size of a signer-struct represented as human-readable []byte
	// SignerpubKey      45 // base58 encoding is variable
	// Nonce             12 // base58 encoding is variable
	// Bits               3
	// 5 * 20           100 // uints
	// 7                  7 // field separators
	// -----------------------------------------
	//                  167
	SignerStructSize = 167
	// SignerStructMin is the minimum size
	SignerStructMin = 17
)

// SignerStruct describes a signer
type SignerStruct struct {
	PublicKey           [message.SignerPubKeySize]byte // Public key of signer
	Nonce               [hashcash.NonceSize]byte       // HashCash nonce
	Bits                byte                           // Hashcash bits. Verified
	MessagesPosted      uint64                         // Total messages posted, always increment
	MessagesRetained    uint64                         // Messages retained
	MaxMessagesPosted   uint64                         // Maximum messages that may be posted
	MaxMessagesRetained uint64                         // Maximum messages that may be retained
	ExpireTarget        uint64                         // Calculated from Bits
}

// SignerStructEncoded represents an encoded SignerStruct
type SignerStructEncoded []byte

// Fill fills an encoded Signerstruct to maximum size
func (sse SignerStructEncoded) Fill() SignerStructEncoded {
	if len(sse) < SignerStructSize {
		return append(sse, bytes.Repeat([]byte(" "), SignerStructSize-len(sse))...)
	}
	return sse
}

// Encode a SignerStruct to human readable representation
func (ss SignerStruct) Encode() SignerStructEncoded {
	out := make([]byte, 0, SignerStructSize)
	out = append(out, []byte(utils.B58encode(ss.PublicKey[:])+" ")...)
	out = append(out, []byte(utils.B58encode(ss.Nonce[:])+" ")...)
	out = append(out, []byte(strconv.FormatUint(uint64(ss.Bits), 10)+" ")...)
	out = append(out, []byte(strconv.FormatUint(ss.MessagesPosted, 10)+" ")...)
	out = append(out, []byte(strconv.FormatUint(ss.MessagesRetained, 10)+" ")...)
	out = append(out, []byte(strconv.FormatUint(ss.MaxMessagesPosted, 10)+" ")...)
	out = append(out, []byte(strconv.FormatUint(ss.MaxMessagesRetained, 10)+" ")...)
	out = append(out, []byte(strconv.FormatUint(ss.ExpireTarget, 10)+" ")...)
	return out[:len(out)]
}

// SignerStructDecode decodes an encoded signerstruct
func SignerStructDecode(d SignerStructEncoded) *SignerStruct {
	if len(d) < SignerStructMin {
		return nil
	}
	fields := bytes.Fields(d)
	if len(fields) != 8 {
		return nil
	}
	cur := 0
	ss := new(SignerStruct)
	t := utils.B58decode(string(fields[cur]))
	copy(ss.PublicKey[:], t)
	cur++
	t = utils.B58decode(string(fields[cur]))
	copy(ss.Nonce[:], t)
	cur++
	tb, _ := strconv.ParseUint(string(fields[cur]), 10, 64)
	ss.Bits = byte(tb)
	cur++
	ss.MessagesPosted, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	ss.MessagesRetained, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	ss.MaxMessagesPosted, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	ss.MaxMessagesRetained, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	ss.ExpireTarget, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	return ss
}
