package sql

import (
	"time"

	"github.com/agl/ed25519"
	"github.com/repbin/repbin/utils/keyproof"
	"github.com/repbin/repbin/utils/repproto/structs"
)

// structs.PeerStruct

// InsertPeer writes a new peer, returns error if exists
func (db *MessageDB) InsertPeer(pubkey *[ed25519.PublicKeySize]byte) error {
	return updateConvertNilError(db.peerInsertQ.Exec(toHex(pubkey[:])))
}

func (db *MessageDB) suppressDuplicateError(err error) error {
	if db.driver == "mysql" {
		if err.Error()[0:27] == "Error 1062: Duplicate entry" {
			return nil
		}
	} else if db.driver == "sqlite3" {
		if err.Error()[0:25] == "UNIQUE constraint failed:" {
			return nil
		}
	}
	return err
}

// TouchPeer inserts a peer but ignores duplicate errors
func (db *MessageDB) TouchPeer(pubkey *[ed25519.PublicKeySize]byte) error {
	return db.suppressDuplicateError(db.InsertPeer(pubkey))
}

// UpdatePeerStats updates the peer statistics
func (db *MessageDB) UpdatePeerStats(pubkey *[ed25519.PublicKeySize]byte, lastFetch, lastPos, lastErrors uint64) error {
	return updateConvertNilError(db.peerUpdateStatQ.Exec(lastFetch, lastPos, lastErrors, toHex(pubkey[:])))
}

// UpdatePeerNotification records the peer's notification ping time
func (db *MessageDB) UpdatePeerNotification(pubkey *[ed25519.PublicKeySize]byte, hasError bool) error {
	var errorDif int
	if hasError {
		errorDif = 1
	}
	return updateConvertNilError(db.peerUpdateNotifyQ.Exec(time.Now().Unix(), errorDif, toHex(pubkey[:])))
}

// UpdatePeerToken records the next authentication token for this peer
func (db *MessageDB) UpdatePeerToken(pubkey *[ed25519.PublicKeySize]byte, signedToken *[keyproof.ProofTokenSignedSize]byte) error {
	return updateConvertNilError(db.peerUpdateTokenQ.Exec(time.Now().Unix(), toHex(signedToken[:]), toHex(pubkey[:])))
}

// SelectPeer returns information about the peer
func (db *MessageDB) SelectPeer(pubkey *[ed25519.PublicKeySize]byte) (*structs.PeerStruct, error) {
	var authtokenT string
	r := new(structs.PeerStruct)
	err := db.peerSelectQ.QueryRow(toHex(pubkey[:])).Scan(
		&authtokenT,
		&r.LastNotifySend,
		&r.LastNotifyFrom,
		&r.LastFetch,
		&r.ErrorCount,
		&r.LastPosition,
	)
	if err != nil {
		return nil, err
	}
	r.AuthToken = *sliceToProofTokenSigned(fromHex(authtokenT))
	return r, nil
}
