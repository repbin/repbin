package messagestore

import (
	"github.com/agl/ed25519"
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/utils/keyproof"
	"github.com/repbin/repbin/utils/repproto/structs"
)

// TouchPeer creates a peer entry if it does not exist yet
func (store Store) TouchPeer(pubkey *[ed25519.PublicKeySize]byte) {
	err := store.db.TouchPeer(pubkey)
	if err != nil {
		log.Errorf("TouchPeer: %s, %x\n", err, *pubkey)
	}
}

// UpdatePeerFetchStat writes fetch-specific data
func (store Store) UpdatePeerFetchStat(pubkey *[ed25519.PublicKeySize]byte, lastFetch, lastPos, lastErrors uint64) {
	err := store.db.UpdatePeerStats(pubkey, lastFetch, lastPos, lastErrors)
	if err != nil {
		log.Errorf("UpdatePeerStats: %s, %x\n", err, *pubkey)
	}
}

// UpdatePeerNotification updates the peer stat after notification send
func (store Store) UpdatePeerNotification(pubkey *[ed25519.PublicKeySize]byte, hasError bool) {
	err := store.db.UpdatePeerNotification(pubkey, hasError)
	if err != nil {
		log.Errorf("UpdatePeerNotification: %s, %x\n", err, *pubkey)
	}
}

// GetPeerStat returns the last entry of peer statistics for pubkey
func (store Store) GetPeerStat(pubkey *[ed25519.PublicKeySize]byte) *structs.PeerStruct {
	st, err := store.db.SelectPeer(pubkey)
	if err != nil {
		log.Errorf("GetPeerStat: %s, %x\n", err, *pubkey)
		return nil
	}
	return st
}

// UpdatePeerAuthToken updates the peer record when a new auth token has been received
func (store Store) UpdatePeerAuthToken(senderPubKey *[ed25519.PublicKeySize]byte, signedToken *[keyproof.ProofTokenSignedSize]byte) {
	err := store.db.UpdatePeerToken(senderPubKey, signedToken)
	if err != nil {
		log.Errorf("UpdatePeerAuthToken: %s, %x\n", err, *senderPubKey)
	}
}
