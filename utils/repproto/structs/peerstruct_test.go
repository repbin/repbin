package structs

import (
	"github.com/repbin/repbin/utils/keyproof"
	"testing"
)

func TestPeerStructEncode(t *testing.T) {
	ps := PeerStruct{
		AuthToken:      [keyproof.ProofTokenSignedSize]byte{0x42, 0x23, 0x40},
		LastNotifySend: ^uint64(0),
		LastNotifyFrom: ^uint64(0),
		LastFetch:      ^uint64(0),
		ErrorCount:     ^uint64(0),
	}
	enc := ps.Encode().Fill()
	if len(enc) != PeerStructSize {
		t.Errorf("Fill failed: %d != %d", len(enc), PeerStructSize)
	}
	dec := PeerStructDecode(enc)
	if ps.AuthToken != dec.AuthToken {
		t.Errorf("Decode AuthToken failed: %x", dec.AuthToken)
	}
	if ps.LastNotifySend != dec.LastNotifySend {
		t.Errorf("Decode LastNotifySend failed: %d", dec.LastNotifySend)
	}
	if ps.LastNotifyFrom != dec.LastNotifyFrom {
		t.Errorf("Decode LastNotifyFrom failed: %d", dec.LastNotifyFrom)
	}
	if ps.LastFetch != dec.LastFetch {
		t.Errorf("Decode LastFetch failed: %d", dec.LastFetch)
	}
	if ps.ErrorCount != dec.ErrorCount {
		t.Errorf("Decode ErrorCount failed: %d", dec.ErrorCount)
	}

}
