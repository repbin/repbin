package client

import (
	"errors"
)

// Version of client functions
const Version = "0.0.2 very alpha"

// PeerUPdateDuration is the maximum time to wait until peer updates are forced
const PeerUpdateDuration = 259200
const innerHeader = 332 // inner header size

var (
	// ErrNoPeers is returned if the client could not find new peers
	ErrNoPeers = errors.New("client: No peers")
	// ErrNoConfig is returned when no config file could be named
	ErrNoConfig = errors.New("client: No config file")
)

// Options are options used in the client
type Options struct {
	KEYVERB bool // switch on key verbosity
	Verbose bool // switch on verbosity
	Appdata bool // switch on appdata output

	Sync   bool // key is synced
	Hidden bool // key is hidden

	Privkey string // private key

	Signkey string // signature key file
	Signdir string // signature directory

	Configfile string // load configuration from file

	Infile  string // read input data from file
	Outfile string // write output data to file
	Stmdir  string // STM dir
	Outdir  string // output directory for batch index download

	Senderkey    string // public key for recipient
	Recipientkey string // public key for recipient
	Embedkey     bool   // embed a new/fresh public key
	Notrace      bool   // embed keys do not depend on private key
	Anonymous    bool   // disable private key and previous signerkeys
	Repost       bool   // create a repost message (will not be posted)
	Mindelay     int    // minimum repost delay
	Maxdelay     int    // maximum repost delay

	Keymgt int // key management file descriptor

	Start int // start position for index fetch
	Count int // maximum number of index entries to fetch

	Socksserver string // socks server to use
	Server      string // server to use

	MessageType int // message type of message
}

// ConfigVariables are settings read from config file
type ConfigVariables struct {
	BodyLength    int
	PadToLength   int
	MinHashCash   byte
	PrivateKey    string
	KeyDir        string   // directory for signKeys
	SocksServer   string   // url of socks server (if any)
	PeerUpdate    int64    // when did we update the peers last?
	BootStrapPeer string   // What peer to bootstrap from
	PasteServers  []string // urls of pastebins
}

// OptionsVar .
var OptionsVar Options

// GobalconfigVar
var GlobalConfigVar ConfigVariables
