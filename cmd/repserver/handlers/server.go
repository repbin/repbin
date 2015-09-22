package handlers

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"net/http"
	"time"

	"github.com/agl/ed25519"
	"github.com/repbin/repbin/cmd/repserver/messagestore"
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/keyauth"
)

// Version of this release
const Version = "0.0.3 very alpha"

const (
	// MaxPostSize maximum size for posts
	MaxPostSize = 96000
	// MinPostSize for posts
	MinPostSize = (message.SignHeaderSize + message.KeyHeaderSize + 1) * 4
	// MinHashCashBits minimum hashcash bits
	MinHashCashBits = 24
	// DefaultMaxSleep is the maximum pause per sleep (info / post handlers)
	DefaultMaxSleep = int64(time.Second) * 3 / 2
	// DefaultIndexGlobal is the maximum number of entries to fetch from the global index per call
	DefaultIndexGlobal = 1000
	// DefaultIndexKey is the maximum number of entries to fetch from a key, per call
	DefaultIndexKey = 23
	// DefaultTimeGrace is the grace time given for authentication
	DefaultTimeGrace = uint64(60)
	// DefaultAuthTokenAge is the maximum age a keyproof token may have
	DefaultAuthTokenAge = 3600 * 3
	// DefaultNotifyDuration is the time to wait between notification sends
	DefaultNotifyDuration = 600 // every 10min
	// DefaultFetchDuration is the time between fetch attempts
	DefaultFetchDuration = 600 // every 10min
	// DefaultFetchMax is the number of messages to fetch from a peer per call
	DefaultFetchMax = 30
	// DefaultListenPort is the port on which to listen for HTTP connections
	DefaultListenPort = 8080
	// DefaultStepLimit is the extra number of bits to overcome for boost to apply in hashcash limits calculation
	DefaultStepLimit = 2
	// DefaultExpireDuration is the time between expire runs
	DefaultExpireDuration = 3600
	// DefaultMaxTimeSkew is the maximum time skew to allow and use
	DefaultMaxTimeSkew = 86400
	// DefaultMinStoreTime is the minimum time to store a message, in seconds
	DefaultMinStoreTime = 86400
	// DefaultMaxStoreTime is the maximum time to ever store a message, in seconds
	DefaultMaxStoreTime = 2592000
	// DefaultAddToPeer determines if the server adds itself to the peerlist
	DefaultAddToPeer = true
	// DefaultMaxAgeSigners defines when to delete signers that are not active anymore
	DefaultMaxAgeSigners = int64(31536000)
	// DefaultMaxAgeRecipients defines when to delete recipients that are not active anymore
	DefaultMaxAgeRecipients = int64(31536000)
)

var (
	// ErrBadMessageID .
	ErrBadMessageID = errors.New("server: MessageID unexpected")
	// ErrNoMore .
	ErrNoMore = errors.New("fileback: No more entries")
)

// Workers defines how many parallel index access goroutines may exist without locking.
var Workers = 100
var enforceTimeOuts = true

// CurrentTime returns the current time in UTC
var CurrentTime = func() int64 { return time.Now().UTC().Unix() }

// MessageServer provides handlers.
type MessageServer struct {
	DB                   *messagestore.Store
	path                 string // root of data
	URL                  string // my own URL
	AddToPeer            bool   // if to add oneself to the peerlist
	MaxPostSize          int64  // Maximum post size
	MinPostSize          int    // Minimum post size
	MinHashCashBits      byte   // Minimum hashcash bits required
	MaxTimeSkew          int64  // Maximum timeskew to use and allow
	authPrivKey          *message.Curve25519Key
	TokenPrivKey         *[ed25519.PrivateKeySize]byte
	AuthPubKey           *message.Curve25519Key       // Used for hidden key index access
	TokenPubKey          *[ed25519.PublicKeySize]byte // Used for peer identification
	InfoStruct           *ServerInfo
	TimeSkew             int64  // TimeSkew changes the returned time by a constant
	TimeGrace            uint64 // Grace time for authentication
	MaxSleep             int64  // MaxSleep is the maximum number of nano seconds of a sleep call
	MaxIndexGlobal       int64  // Maximum entries from global index
	MaxIndexKey          int64  // Maximum entries from key index
	MaxAuthTokenAge      int64  // Maximum age of peer authentication token
	PeerFile             string // File containing the peer information
	NotifyDuration       int64  // Time between notifications
	FetchDuration        int64  // Time between fetches
	FetchMax             int    // Maximum messages to fetch per call to peer
	ExpireDuration       int64  // Time between expire runs
	SocksProxy           string // Socks5 proxy
	EnableDeleteHandler  bool   // should the delete handler be enabled?
	EnableOneTimeHandler bool   // should the one-time message handler be enabled?
	EnablePeerHandler    bool   // show the peer handler be offered?
	HubOnly              bool   // should this server act only as hub?
	StepLimit            int    // Boost limit
	ListenPort           int    // what port to listen on for http
	MinStoreTime         int    // minimum time in seconds for storage
	MaxStoreTime         int    // maximum time in seconds for storage
	Stat                 bool   // calculate and show server usage statistics
	MaxAgeSigners        int64
	MaxAgeRecipients     int64

	notifyChan chan bool // Notification channel. Write to notify system about new message
}

// ServerInfo contains the public server information.
type ServerInfo struct {
	Time            int64
	AuthPubKey      string
	AuthChallenge   string
	MaxPostSize     int64    // Maximum post size
	MinPostSize     int      // Minimum post size
	MinHashCashBits byte     // Minimum hashcash bits required
	Peers           []string // list of known peers
}

// New returns a MessageServer.
func New(driver, url, path string, pubKey, privKey []byte) (*MessageServer, error) {
	var err error
	ms := new(MessageServer)
	ms.DB = messagestore.New(driver, url, path, Workers)
	if err != nil {
		return nil, err
	}
	ms.path = path
	ms.PeerFile = path + "peers.config"
	ms.AddToPeer = DefaultAddToPeer
	ms.MaxTimeSkew = DefaultMaxTimeSkew
	ms.MinPostSize = MinPostSize
	ms.MaxPostSize = MaxPostSize
	ms.MinHashCashBits = MinHashCashBits
	ms.MaxSleep = DefaultMaxSleep
	ms.MaxIndexGlobal = DefaultIndexGlobal
	ms.MaxIndexKey = DefaultIndexKey
	ms.TimeGrace = DefaultTimeGrace
	ms.MaxAuthTokenAge = DefaultAuthTokenAge
	ms.NotifyDuration = DefaultNotifyDuration
	ms.FetchDuration = DefaultFetchDuration
	ms.FetchMax = DefaultFetchMax
	ms.ExpireDuration = DefaultExpireDuration
	ms.StepLimit = DefaultStepLimit
	ms.ListenPort = DefaultListenPort
	ms.MinStoreTime = DefaultMinStoreTime
	ms.MaxStoreTime = DefaultMaxStoreTime
	ms.MaxAgeSigners = DefaultMaxAgeSigners
	ms.MaxAgeRecipients = DefaultMaxAgeRecipients
	messagestore.MaxAgeRecipients = DefaultMaxAgeRecipients
	messagestore.MaxAgeSigners = DefaultMaxAgeSigners
	ms.EnablePeerHandler = true

	ms.authPrivKey, err = message.GenLongTermKey(true, false)
	if err != nil {
		return nil, err
	}
	ms.AuthPubKey = message.GenPubKey(ms.authPrivKey)
	if pubKey != nil && privKey != nil {
		ms.TokenPubKey = new([ed25519.PublicKeySize]byte)
		ms.TokenPrivKey = new([ed25519.PrivateKeySize]byte)
		copy(ms.TokenPubKey[:], pubKey)
		copy(ms.TokenPrivKey[:], privKey)
	} else {
		ms.TokenPubKey, ms.TokenPrivKey, err = ed25519.GenerateKey(rand.Reader)
		log.Printf("Peer authentication public key: %s\n", utils.B58encode(ms.TokenPubKey[:]))
		log.Printf("Peer authentication private key: %s\n", utils.B58encode(ms.TokenPrivKey[:]))
		if err != nil {
			return nil, err
		}
	}
	ms.InfoStruct = new(ServerInfo)
	ms.InfoStruct.MinHashCashBits = ms.MinHashCashBits
	ms.InfoStruct.MinPostSize = ms.MinPostSize
	ms.InfoStruct.MaxPostSize = ms.MaxPostSize
	ms.InfoStruct.AuthPubKey = utils.B58encode(ms.AuthPubKey[:])
	return ms, nil
}

func (ms MessageServer) calcLimits(hashcashbits byte) (maxPost, maxRetain, expire uint64) {
	extra := int(hashcashbits - ms.MinHashCashBits)
	if extra < 0 {
		return 0, 0, 0
	}
	if extra < ms.StepLimit {
		return 1, 1, uint64(ms.MinStoreTime) // Minimum bits only buy 2h of storage, and 2 messages
	}
	extra -= ms.StepLimit
	// Boost extra bits
	raiseFactor := uint64(math.Ceil(math.Pow(2, float64(extra)*1.33)))
	maxPost = raiseFactor
	maxRetain = raiseFactor
	if uint64(ms.MinStoreTime)*raiseFactor > uint64(ms.MaxStoreTime) { // Never more than 30 days of storage
		return raiseFactor + 2, raiseFactor + 2, uint64(ms.MaxStoreTime)
	}
	return raiseFactor + 2, raiseFactor + 2, uint64(ms.MinStoreTime) * raiseFactor
}

func randomNumber(max int64) int64 {
	sleep, err := rand.Int(rand.Reader, big.NewInt(max))
	if err == nil {
		return sleep.Int64()
	}
	return max / 2
}

// RandomSleep makes the go routine sleep for a random time up to ms.MaxSleep.
func (ms MessageServer) RandomSleep() {
	time.Sleep(time.Duration(randomNumber(ms.MaxSleep)))
}

// ServeID returns server information.
func (ms MessageServer) ServeID(w http.ResponseWriter, r *http.Request) {
	ms.RandomSleep()
	now := CurrentTime() + ms.TimeSkew
	privK := [32]byte(*ms.authPrivKey)
	_, pubkey, challenge := keyauth.GenTempKeyTime(uint64(now), &privK)
	info := &ServerInfo{
		Time:            now,
		AuthPubKey:      utils.B58encode(pubkey[:]),
		AuthChallenge:   utils.B58encode(challenge[:]),
		MaxPostSize:     int64(messagestore.MaxMessageSize),
		MinPostSize:     ms.InfoStruct.MinPostSize,
		MinHashCashBits: ms.InfoStruct.MinHashCashBits,
	}
	if ms.EnablePeerHandler {
		info.Peers = ms.getPeerURLs()
	}
	ms.RandomSleep()
	w.Header().Set("Content-Type", "text/json")
	b, err := json.Marshal(info)
	if err != nil {
		log.Debugf("Info: %s\n", err)
	}
	w.Write(b)
	w.Write([]byte("\n"))
	return
}
