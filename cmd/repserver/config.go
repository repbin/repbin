package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/agl/ed25519"
	"github.com/repbin/repbin/cmd/repserver/handlers"
	"github.com/repbin/repbin/cmd/repserver/messagestore"
	"github.com/repbin/repbin/utils"
)

// ServerConfig are configurable parameters
type ServerConfig struct {
	URL                  string // my own URL
	AddToPeer            bool   // if to add self to peers
	MaxTimeSkew          int64  // Maximum time skew to use and allow
	MinHashCashBits      byte   // Minimum hashcash bits required
	NotifyDuration       int64  // Time between notifications
	FetchDuration        int64  // Time between fetches
	ExpireDuration       int64  // Time between expire runs
	ExpireFSDuration     int64  // Time between filesystem expire runs
	SocksProxy           string // Socks5 proxy
	EnableDeleteHandler  bool   // should the delete handler be enabled?
	EnableOneTimeHandler bool   // should the one-time message handler be enabled?
	EnablePeerHandler    bool   // should the peer handler (peer discovery for clients) be enabled?
	HubOnly              bool   // should this server only act as hub?
	StepLimit            int    // Boost limit
	ListenPort           int    // what port to listen on for http
	StoragePath          string // Where to store files
	MinStoreTime         int    // Minimum time to store a message
	MaxStoreTime         int    // Maximum time to store a message
	PeeringPublicKey     string // public key for peering authentication
	PeeringPrivateKey    string // private key for peering authentication
	DBDriver             string // Database driver of the storage backend (mysql, sqlite3)
	DBURL                string // database access URL, user:password@server/database
	MaxAgeSigners        int64
	MaxAgeRecipients     int64
}

var defaultSettings = &ServerConfig{
	AddToPeer:            handlers.DefaultAddToPeer,
	MaxTimeSkew:          handlers.DefaultMaxTimeSkew,
	MinHashCashBits:      handlers.MinHashCashBits,
	NotifyDuration:       handlers.DefaultNotifyDuration,
	FetchDuration:        handlers.DefaultFetchDuration,
	ExpireDuration:       handlers.DefaultExpireDuration,
	ExpireFSDuration:     handlers.DefaultExpireFSDuration,
	StepLimit:            handlers.DefaultStepLimit,
	ListenPort:           handlers.DefaultListenPort,
	EnableDeleteHandler:  false,
	EnableOneTimeHandler: false,
	EnablePeerHandler:    true,
	HubOnly:              false,
	SocksProxy:           "socks5://127.0.0.1:9050/",
	StoragePath:          "",
	MinStoreTime:         handlers.DefaultMinStoreTime,
	MaxStoreTime:         handlers.DefaultMaxStoreTime,
	PeeringPublicKey:     "",
	PeeringPrivateKey:    "",
	DBDriver:             "mysql",
	DBURL:                "repbin:repbin@/repbin",
	MaxAgeSigners:        handlers.DefaultMaxAgeSigners,
	MaxAgeRecipients:     handlers.DefaultMaxAgeRecipients,
}

// showConfig shows current (default) config
func showConfig() {
	if defaultSettings.PeeringPrivateKey == "" {
		pubkey, privkey, err := ed25519.GenerateKey(rand.Reader)
		if err == nil {
			defaultSettings.PeeringPrivateKey = utils.B58encode(privkey[:])
			defaultSettings.PeeringPublicKey = utils.B58encode(pubkey[:])
		}
	}
	config, _ := json.MarshalIndent(defaultSettings, "", "    ")
	fmt.Println(string(config))
}

func loadConfig(filename string) error {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(d, defaultSettings)
	if err != nil {
		return err
	}
	return nil
}

func applyConfig(ms *handlers.MessageServer) {
	ms.AddToPeer = defaultSettings.AddToPeer
	ms.URL = defaultSettings.URL
	ms.MaxTimeSkew = defaultSettings.MaxTimeSkew
	ms.MinHashCashBits = defaultSettings.MinHashCashBits
	ms.NotifyDuration = defaultSettings.NotifyDuration
	ms.FetchDuration = defaultSettings.FetchDuration
	ms.ExpireDuration = defaultSettings.ExpireDuration
	ms.ExpireFSDuration = defaultSettings.ExpireFSDuration
	ms.StepLimit = defaultSettings.StepLimit
	ms.ListenPort = defaultSettings.ListenPort
	ms.MinStoreTime = defaultSettings.MinStoreTime
	ms.MaxStoreTime = defaultSettings.MaxStoreTime
	ms.EnableDeleteHandler = defaultSettings.EnableDeleteHandler
	ms.EnableOneTimeHandler = defaultSettings.EnableOneTimeHandler
	ms.EnablePeerHandler = defaultSettings.EnablePeerHandler
	ms.HubOnly = defaultSettings.HubOnly
	ms.SocksProxy = defaultSettings.SocksProxy
	ms.MaxAgeSigners = defaultSettings.MaxAgeSigners
	ms.MaxAgeRecipients = defaultSettings.MaxAgeRecipients
	messagestore.MaxAgeSigners = defaultSettings.MaxAgeSigners
	messagestore.MaxAgeRecipients = defaultSettings.MaxAgeSigners
}
