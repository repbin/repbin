package structs

import (
	"bytes"
	"strconv"

	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
)

// Version of this release
const Version = "0.0.1 very alpha"

const (
	// MessageStructSize is the maximum size a message struct takes converted to byte/string
	// 10 field separators 	    10
	// 5 * 20 bytes (uints)	   100
	// curve25519 -> 45 	    45 // base58 encoding is variable
	// messageID -> 45 		    45 // base58 encoding is variable
	// signerPub -> 90 		    90 // base58 encoding is variable
	// 3 * 5 bytes bools 	    15
	// ---------------------------
	//                         295
	MessageStructSize = 274
	// MessageStructMin Minimum size to start parsing
	MessageStructMin = 32
)

// MessageStruct describes a message
type MessageStruct struct {
	Counter                uint64                         // is zero unless from a list
	PostTime               uint64                         // When was the message posted
	ExpireTime             uint64                         // When does the message expire. Calculate from ExpireRequest
	ExpireRequest          uint64                         // Requested expire time
	MessageID              [message.MessageIDSize]byte    // The messageID
	ReceiverConstantPubKey message.Curve25519Key          // The ReceiverConstantPubKey
	SignerPub              [message.SignerPubKeySize]byte // The Signer Pubkey
	Distance               uint64                         // The distance to origin. Increased by one when fetching
	OneTime                bool                           // Mark message as one-time. Message is deleted on fetch
	Sync                   bool                           // Message will be synced (0x00==no,0x01==yes)
	Hidden                 bool                           // Message is hidden (0x00==no,0x01==yes)
}

// MessageStructEncoded represents an encoded MessageStruct
type MessageStructEncoded []byte

// Fill fills an encoded messagestruct to maximum size
func (mse MessageStructEncoded) Fill() MessageStructEncoded {
	if len(mse) < MessageStructSize {
		return append(mse, bytes.Repeat([]byte(" "), MessageStructSize-len(mse))...)
	}
	return mse
}

// Encode encodes a MessageStruct in human-readable format for storage
func (ms MessageStruct) Encode() MessageStructEncoded {
	out := make([]byte, 0, MessageStructSize)
	out = append(out, []byte(strconv.FormatUint(ms.Counter, 10)+" ")...)
	out = append(out, []byte(strconv.FormatUint(ms.PostTime, 10)+" ")...)
	out = append(out, []byte(strconv.FormatUint(ms.ExpireTime, 10)+" ")...)
	out = append(out, []byte(strconv.FormatUint(ms.ExpireRequest, 10)+" ")...)
	out = append(out, []byte(utils.B58encode(ms.MessageID[:])+" ")...)
	out = append(out, []byte(utils.B58encode(ms.ReceiverConstantPubKey[:])+" ")...)
	out = append(out, []byte(utils.B58encode(ms.SignerPub[:])+" ")...)
	out = append(out, []byte(strconv.FormatUint(ms.Distance, 10)+" ")...)
	out = append(out, []byte(BoolToString(ms.OneTime)+" ")...)
	out = append(out, []byte(BoolToString(ms.Sync)+" ")...)
	out = append(out, []byte(BoolToString(ms.Hidden))...)
	return out[:len(out)]
}

// MessageStructDecode decodes bytes to MessageStruct
func MessageStructDecode(d MessageStructEncoded) *MessageStruct {
	fields := bytes.Fields(d)
	l := len(fields)
	if l < 10 || l > 11 || len(d) < MessageStructMin { // with or without counter
		return nil
	}
	cur := 0
	ms := new(MessageStruct)
	if l == 11 { // with counter\
		ms.Counter, _ = strconv.ParseUint(string(fields[0]), 10, 64)
		cur++
	}
	ms.PostTime, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	ms.ExpireTime, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	ms.ExpireRequest, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	t := utils.B58decode(string(fields[cur]))
	copy(ms.MessageID[:], t)
	cur++
	t = utils.B58decode(string(fields[cur]))
	copy(ms.ReceiverConstantPubKey[:], t)
	cur++
	t = utils.B58decode(string(fields[cur]))
	copy(ms.SignerPub[:], t)
	cur++
	ms.Distance, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	ms.OneTime = StringToBool(string(fields[cur]))
	cur++
	ms.Sync = StringToBool(string(fields[cur]))
	cur++
	ms.Hidden = StringToBool(string(fields[cur]))
	return ms
}

/*
func ParseInt(s string, base int, bitSize int) (i int64, err error)
func ParseUint(s string, base int, bitSize int) (n uint64, err error)
func FormatInt(i int64, base int) string
func FormatUint(i uint64, base int) string
*/
