// Package messagestore implements storage for messages
package messagestore

import (
	"errors"
	"github.com/repbin/repbin/fileback"
	"github.com/repbin/repbin/utils/repproto/structs"
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

// ExpireRun defines the time in seconds between expire runs
var ExpireRun = 3600

// FSExpire is the time a filesystem entry must expire when not changed
var FSExpire = 2678400
var globalindex = []byte("globalindex")

// Store implements a message store
type Store struct {
	messages    fileback.PagedList // Messages themselves, indexed by MessageID
	signers     fileback.PagedList // Signer data, indexed by SignerPubKey
	keyindex    fileback.PagedList // Key indices. Pubkey is index. Special case: globalindex
	expireindex fileback.PagedList // Messages to expire, indexed by expire time
	peersindex  fileback.PagedList // peer information indexed by peer public key
}

// New Create a new message store at directory dir. Workers is the maximum concurrent access to an index
func New(dir string, workers int) *Store {
	s := new(Store)
	s.messages = fileback.NewBlob(dir+messageDir, int64(MaxMessageSize+structs.MessageStructSize), workers)
	s.signers = fileback.NewRoundRobin(dir+signerDir, structs.SignerStructSize, 10, ' ', []byte("\n"), workers)       // We keep signer history, last 10 entries
	s.keyindex = fileback.NewRolling(dir+keyindexDir, structs.MessageStructSize, 2048, ' ', []byte("\n"), workers)    // Maximum 2048 entries per file
	s.expireindex = fileback.NewContinue(dir+expireDir, structs.ExpireStructSize, 102400, ' ', []byte("\n"), workers) // Don't expire more than 100k
	s.peersindex = fileback.NewRoundRobin(dir+peersDir, structs.PeerStructSize, 10, ' ', []byte("\n"), workers)
	return s
}
