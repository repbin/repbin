package messagestore

import (
	"github.com/agl/ed25519"
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/fileback"
	"github.com/repbin/repbin/utils/keyproof"
	"github.com/repbin/repbin/utils/repproto/structs"
	"time"
)

// TouchPeer creates a peer entry if it does not exist yet
func (store Store) TouchPeer(pubkey *[ed25519.PublicKeySize]byte) {
	if !store.peersindex.Index(pubkey[:]).Exists() {
		s := structs.PeerStruct{
			AuthToken: [keyproof.ProofTokenSignedSize]byte{0x00},
		}
		store.peersindex.Index(pubkey[:]).Create(s.Encode().Fill())
	}
}

// UpdatePeerFetchStat writes fetch-specific data
func (store Store) UpdatePeerFetchStat(pubkey *[ed25519.PublicKeySize]byte, lastFetch, lastPos, lastErrors uint64) {
	if store.peersindex.Index(pubkey[:]).Exists() {
		store.peersindex.Index(pubkey[:]).Update(func(tx fileback.Tx) error {
			last := tx.GetLast()
			if last == nil {
				log.Errorf("Peer last nil: %x\n", *pubkey)
				return nil
			}
			peerStat := structs.PeerStructDecode(last)
			if peerStat == nil {
				log.Errorf("Peer stat nil: %x\n", *pubkey)
				return nil
			}
			peerStat.LastFetch = lastFetch
			peerStat.LastPosition = lastPos
			peerStat.ErrorCount = lastErrors
			tx.Append(peerStat.Encode().Fill())
			return nil
		})
	}
}

// UpdatePeerNotification updates the peer stat after notification send
func (store Store) UpdatePeerNotification(pubkey *[ed25519.PublicKeySize]byte, hasError bool) {
	if store.peersindex.Index(pubkey[:]).Exists() {
		store.peersindex.Index(pubkey[:]).Update(func(tx fileback.Tx) error {
			last := tx.GetLast()
			if last == nil {
				log.Errorf("Peer last nil: %x\n", *pubkey)
				return nil
			}
			peerStat := structs.PeerStructDecode(last)
			if peerStat == nil {
				log.Errorf("Peer stat nil: %x\n", *pubkey)
				return nil
			}
			peerStat.LastNotifySend = uint64(time.Now().Unix())
			if hasError {
				peerStat.ErrorCount++
			}
			tx.Append(peerStat.Encode().Fill())
			return nil
		})
	}
}

// GetPeerStat returns the last entry of peer statistics for pubkey
func (store Store) GetPeerStat(pubkey *[ed25519.PublicKeySize]byte) *structs.PeerStruct {
	if store.peersindex.Index(pubkey[:]).Exists() {
		last := store.peersindex.Index(pubkey[:]).GetLast()
		if last == nil {
			log.Errorf("Peer last nil: %x\n", *pubkey)
			return nil
		}
		peerStat := structs.PeerStructDecode(last)
		if peerStat == nil {
			log.Errorf("Peer stat nil: %x\n", *pubkey)
			return nil
		}
		return peerStat
	}
	return nil
}

// UpdatePeerAuthToken updates the peer record when a new auth token has been received
func (store Store) UpdatePeerAuthToken(senderPubKey *[ed25519.PublicKeySize]byte, signedToken *[keyproof.ProofTokenSignedSize]byte) {
	if store.peersindex.Index(senderPubKey[:]).Exists() {
		store.peersindex.Index(senderPubKey[:]).Update(func(tx fileback.Tx) error {
			last := tx.GetLast()
			if last == nil {
				log.Errorf("Peer last nil: %x\n", *senderPubKey)
				return nil
			}
			peerStat := structs.PeerStructDecode(last)
			if peerStat == nil {
				log.Errorf("Peer stat nil: %x\n", *senderPubKey)
				return nil
			}
			peerStat.LastNotifyFrom = uint64(time.Now().Unix())
			peerStat.AuthToken = *signedToken
			tx.Append(peerStat.Encode().Fill())
			return nil
		})
	}
}
