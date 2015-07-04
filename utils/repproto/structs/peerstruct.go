package structs

import (
	"bytes"
	"strconv"

	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/keyproof"
)

const (
	// PeerStructSize is the maximum size of an encoded PeerStruct
	// Uints 20*4     100
	// field sep        6
	// ------------------
	//                106 + AuthToken
	PeerStructSize = 106 + keyproof.ProofTokenSignedMax
	// PeerStructMin is the mimimum size of a peer struct
	PeerStructMin = 13
	// PeerPubKeySize is the size of a peer's public key when encoded
	PeerPubKeySize = 44
)

// PeerStruct represents a peer
type PeerStruct struct {
	AuthToken      [keyproof.ProofTokenSignedSize]byte // Keyproof token, countersigned
	LastNotifySend uint64                              // When did we send a notify last
	LastNotifyFrom uint64                              // When did we receive a notify last
	LastFetch      uint64                              // When did we last fetch from this peer
	ErrorCount     uint64                              // Number of errors occured
	LastPosition   uint64                              // Last successful position in download
}

// PeerStructEncoded represents an encoded PeerStruct
type PeerStructEncoded []byte

// Fill fills an encoded peerstruct to maximum size
func (pse PeerStructEncoded) Fill() PeerStructEncoded {
	if len(pse) < PeerStructSize {
		return append(pse, bytes.Repeat([]byte(" "), PeerStructSize-len(pse))...)
	}
	return pse
}

// Encode peerstruct into human readable form
func (ps PeerStruct) Encode() PeerStructEncoded {
	out := make([]byte, 0, PeerStructSize)
	out = append(out, []byte(utils.B58encode(ps.AuthToken[:])+" ")...)
	out = append(out, []byte(strconv.FormatUint(ps.LastNotifySend, 10)+" ")...)
	out = append(out, []byte(strconv.FormatUint(ps.LastNotifyFrom, 10)+" ")...)
	out = append(out, []byte(strconv.FormatUint(ps.LastFetch, 10)+" ")...)
	out = append(out, []byte(strconv.FormatUint(ps.ErrorCount, 10)+" ")...)
	out = append(out, []byte(strconv.FormatUint(ps.LastPosition, 10)+" ")...)
	return out[:len(out)]
}

// PeerStructDecode decodes an encoded peer struct
func PeerStructDecode(d PeerStructEncoded) *PeerStruct {
	if len(d) < PeerStructMin {
		return nil
	}
	fields := bytes.Fields(d)
	if len(fields) != 6 {
		return nil
	}
	ps := new(PeerStruct)
	cur := 0
	t := utils.B58decode(string(fields[cur]))
	copy(ps.AuthToken[:], t)
	cur++
	ps.LastNotifySend, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	ps.LastNotifyFrom, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	ps.LastFetch, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	ps.ErrorCount, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	cur++
	ps.LastPosition, _ = strconv.ParseUint(string(fields[cur]), 10, 64)
	return ps
}
