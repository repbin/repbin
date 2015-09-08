// Package messagestore implements storage for messages.
package messagestore

import (
	"errors"

	"github.com/repbin/repbin/cmd/repserver/messagestore/sql"
)

// Version of this release
const Version = "0.0.1 very alpha"

const (
	messageDir  = "messages"
	signerDir   = "signers"
	keyindexDir = "keyindex"
	expireDir   = "expire"
	peersDir    = "peers"
)

var (
	// ErrDuplicate is returned if a message is a duplicate
	ErrDuplicate = errors.New("messagestore: Duplicate message")
	// ErrPostLimit is returned when a signer has posted more messages than allowed by calculation
	ErrPostLimit = errors.New("messagestore: Signer has reached post limit")
	// ErrNotFound is returned when a message was fetched that didn't exist
	ErrNotFound = errors.New("messagestore: Message not found")
	// MaxMessageSize is the size of an encoded repbin message
	MaxMessageSize = 87776
)

// MaxAgeSigners defines when to delete signers that are not active anymore
var MaxAgeSigners = int64(31536000)

// MaxAgeRecipients defines when to delete recipients that are not active anymore
var MaxAgeRecipients = int64(31536000)

// Store implements a message store
type Store struct {
	db *sql.MessageDB
}

// New Create a new message store at directory dir. Workers is the maximum concurrent access to an index
func New(driver, url, dir string, workers int) *Store {
	var err error
	s := new(Store)
	s.db, err = sql.New(driver, url, dir, workers)
	if err != nil {
		panic(err)
	}
	return s
}
