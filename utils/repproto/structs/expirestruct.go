package structs

import (
	"bytes"
	"strconv"

	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
)

const (
	// ExpireStructSize is the maximum size of an expirestruct when encoded
	// Expiretime   20
	// MessageID    45
	// Signer       90
	// field sep     2
	// ---------------
	//             157
	ExpireStructSize = 157
	// ExpireStructMin is the minimum size of an encoded expirestruct
	ExpireStructMin = 5
)

// ExpireStruct represents expire information for a message
type ExpireStruct struct {
	ExpireTime   uint64                         // When does the message expire
	MessageID    [message.MessageIDSize]byte    // The messageID
	SignerPubKey [message.SignerPubKeySize]byte // The Signer Pubkey
}

// ExpireStructEncoded represents an encoded ExpireStruct
type ExpireStructEncoded []byte

// Fill fills an encoded messagestruct to maximum size
func (ese ExpireStructEncoded) Fill() ExpireStructEncoded {
	if len(ese) < ExpireStructSize {
		return append(ese, bytes.Repeat([]byte(" "), ExpireStructSize-len(ese))...)
	}
	return ese
}

// Encode an expireStruct
func (es ExpireStruct) Encode() ExpireStructEncoded {
	out := make([]byte, 0, ExpireStructSize)
	out = append(out, []byte(strconv.FormatUint(es.ExpireTime, 10)+" ")...)
	out = append(out, []byte(utils.B58encode(es.MessageID[:])+" ")...)
	out = append(out, []byte(utils.B58encode(es.SignerPubKey[:])+" ")...)
	return out[:len(out)]
}

// ExpireStructDecode decodes an encoded ExpireStruct
func ExpireStructDecode(d ExpireStructEncoded) *ExpireStruct {
	if len(d) < ExpireStructMin {
		return nil
	}
	fields := bytes.Fields(d)
	if len(fields) < 3 {
		return nil
	}
	cur := 0
	es := new(ExpireStruct)
	es.ExpireTime, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	t := utils.B58decode(string(fields[cur]))
	copy(es.MessageID[:], t)
	cur++
	t = utils.B58decode(string(fields[cur]))
	copy(es.SignerPubKey[:], t)
	return es
}
