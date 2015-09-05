package handlers

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	"github.com/agl/ed25519"
	"github.com/repbin/repbin/cmd/repserver/messagestore"
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/keyproof"
	"github.com/repbin/repbin/utils/repproto"
	"github.com/repbin/repbin/utils/repproto/structs"
)

const (
	exampleURL = "http://example.onion/"
)

// Peer is a single peer as loaded from file.
type Peer struct {
	PubKey [ed25519.PublicKeySize]byte // Peer's public key for ping authentication
	URL    string                      // Peer's URL
	IsHub  bool                        // Is the peer a hub?
}

// PeerEncoded is an encoded peer.
type PeerEncoded struct {
	PubKey string
	URL    string
	IsHub  bool
}

// PeerList contains all peers.
type PeerList map[[ed25519.PublicKeySize]byte]Peer

// PeerListEncoded is an encoded peer list.
type PeerListEncoded []PeerEncoded

var peerURLs []string
var peerURLsMutex *sync.Mutex    // access to peerURLs
var systemPeersMutex *sync.Mutex // access mutex for systemPeers
var systemPeers PeerList         // Current peers known
var timeout = 91                 // notify timeout
var debug = false                // Debugging workarounds

func init() {
	systemPeersMutex = new(sync.Mutex)
	peerURLsMutex = new(sync.Mutex)
}

// LoadPeers from file.
func (ms MessageServer) LoadPeers() {
	var prePeerlist PeerListEncoded
	var myPeerURLs []string
	myPeers := make(PeerList)
	peersEncoded, err := ioutil.ReadFile(ms.PeerFile)
	if err != nil {
		example := make(PeerListEncoded, 1)
		example[0].PubKey = "Example ed25519 public key encoded in base58"
		example[0].URL = exampleURL
		dat, _ := json.MarshalIndent(example, "", "    ")
		_ = dat
		err := ioutil.WriteFile(ms.PeerFile, dat, 0600)
		if err != nil {
			log.Errorf("Could not write peerfile: %s\n", err)
			return
		}
		log.Debugs("Example peerfile written\n")
		return
	}
	err = json.Unmarshal(peersEncoded, &prePeerlist)
	if err != nil {
		log.Errorf("Could not read peerfile: %s\n", err)
		return
	}
	if ms.AddToPeer && !ms.HubOnly {
		myPeerURLs = append(myPeerURLs, ms.URL)
	}
	for _, p := range prePeerlist {
		if p.URL != exampleURL {
			var tkey [ed25519.PublicKeySize]byte
			t := utils.B58decode(p.PubKey)
			copy(tkey[:], t)
			if *ms.TokenPubKey != tkey || debug {
				myPeers[tkey] = Peer{
					URL:    p.URL,
					PubKey: tkey,
				}
				if !p.IsHub {
					myPeerURLs = append(myPeerURLs, p.URL)
				}
				// Create Peer Index File
				ms.DB.TouchPeer(&tkey)
			}
		}
	}
	defer ms.setPeerURLs(myPeerURLs)
	systemPeersMutex.Lock()
	defer systemPeersMutex.Unlock()
	systemPeers = myPeers
	log.Debugs("Peers loaded\n")
}

func (ms MessageServer) setPeerURLs(urls []string) {
	peerURLsMutex.Lock()
	defer peerURLsMutex.Unlock()
	peerURLs = urls
}

func (ms MessageServer) getPeerURLs() []string {
	peerURLsMutex.Lock()
	defer peerURLsMutex.Unlock()
	ret := make([]string, len(peerURLs))
	for k, v := range peerURLs {
		ret[k] = v
	}
	return ret
}

// PeerKnown tests if a peer is known.
func (ms MessageServer) PeerKnown(PubKey *[ed25519.PublicKeySize]byte) bool {
	systemPeersMutex.Lock()
	defer systemPeersMutex.Unlock()
	_, ok := systemPeers[*PubKey]
	return ok
}

// NotifyPeers runs notification for all peers.
func (ms MessageServer) NotifyPeers() {
	systemPeersMutex.Lock()
	defer systemPeersMutex.Unlock()
	for peerPubKey, peer := range systemPeers {
		peerPubKeyTemp := new([ed25519.PublicKeySize]byte) // We need to create a new pointer
		copy(peerPubKeyTemp[:], peerPubKey[:])
		go ms.notifyPeer(peerPubKeyTemp, peer.URL)
	}
}

func (ms MessageServer) notifyPeer(PubKey *[ed25519.PublicKeySize]byte, url string) {
	rand.Seed(time.Now().UnixNano())
	maxSleep := ms.NotifyDuration - int64(timeout)
	if maxSleep > 0 {
		sleeptime := rand.Int63() % maxSleep
		time.Sleep(time.Duration(sleeptime) * time.Second)
	}
	now := time.Now().Unix() + ms.TimeSkew
	token := utils.B58encode(keyproof.SignProofToken(now, PubKey, ms.TokenPubKey, ms.TokenPrivKey)[:])
	// Socks call
	proto := repproto.New(ms.SocksProxy, "")
	err := proto.Notify(url, token)
	// Write result
	if err != nil {
		log.Debugf("Notify error: %s\n", err)
		ms.DB.UpdatePeerNotification(PubKey, true)
	} else {
		log.Debugf("Notified peer: %x\n", *PubKey)
		ms.DB.UpdatePeerNotification(PubKey, false)
	}
}

// FetchPeers checks the peers for new messages, and downloads them.
func (ms MessageServer) FetchPeers() {
	systemPeersMutex.Lock()
	defer systemPeersMutex.Unlock()
	for peerPubKey, peer := range systemPeers {
		peerPubKeyTemp := new([ed25519.PublicKeySize]byte) // We need to create a new pointer
		copy(peerPubKeyTemp[:], peerPubKey[:])
		go ms.fetchPeer(peerPubKeyTemp, peer.URL)
	}
}

// fetchPeer downloads new messages from peer.
func (ms MessageServer) fetchPeer(PubKey *[ed25519.PublicKeySize]byte, url string) {
	// Get token, lastpos from database
	var doUpdate bool
	log.Debugf("fetch from peer: %x\n", *PubKey)
	peerStat := ms.DB.GetPeerStat(PubKey)
	if peerStat == nil { // Errors are ignored
		log.Debugf("fetch from peer: not found %x\n", *PubKey)
		return
	}
	if peerStat.LastNotifyFrom == 0 { // We never heard from him
		log.Debugf("fetch from peer: no notification %x\n", *PubKey)
		return
	}
	startDate := time.Now().Unix()
	// Check that LastNotifyFrom > LastFetch
	if ms.HubOnly {
		// Pre-Emptive fetch in hub-mode: every 4 times FetchDuration or when notified
		if peerStat.LastFetch != 0 && !(peerStat.LastFetch < peerStat.LastNotifyFrom || peerStat.LastFetch < uint64(startDate-(ms.FetchDuration*4))) {
			log.Debugf("fetch from peer: no fetch-force and no notifications since last fetch %x\n", *PubKey)
			log.Debugf("LastFetch: %d  LastNotifyFrom: %d\n", peerStat.LastFetch, peerStat.LastNotifyFrom)
			return // nothing new to gain
		}
	} else {
		if peerStat.LastFetch > peerStat.LastNotifyFrom && peerStat.LastFetch != 0 {
			log.Debugf("fetch from peer: no notifications since last fetch %x\n", *PubKey)
			log.Debugf("LastFetch: %d  LastNotifyFrom: %d\n", peerStat.LastFetch, peerStat.LastNotifyFrom)
			return // nothing new to gain
		}
	}
	rand.Seed(time.Now().UnixNano())
	maxSleep := ms.NotifyDuration - int64(timeout)
	if maxSleep > 0 {
		sleeptime := rand.Int63() % maxSleep
		log.Debugf("fetch from peer: sleeping %d %x\n", sleeptime, *PubKey)
		time.Sleep(time.Duration(sleeptime) * time.Second)
	}
	doUpdate = true
FetchLoop:
	for {
		// Make GetIndex call
		proto := repproto.New(ms.SocksProxy, "")
		nextPosition := int(peerStat.LastPosition)
		if nextPosition != 0 {
			nextPosition++
		}
		log.Debugf("GlobalIndex fetch: %s next: %d max: %d\n", url, nextPosition, ms.FetchMax)
		authtoken := utils.B58encode(peerStat.AuthToken[:])
		messages, more, err := proto.GetGlobalIndex(url, authtoken, nextPosition, int(ms.FetchMax))
		if err != nil {
			peerStat.ErrorCount++
			log.Debugf("GlobalIndex err: %s\n", err)
			// Only hit on proxy errors
			errstr := err.Error()
			if len(errstr) >= 6 && errstr[0:6] == "proxy:" {
				doUpdate = false
			}
			break FetchLoop
		}
		for _, msg := range messages {
			log.Debugf("Index: %d  Message: %s\n", msg.Counter, utils.B58encode(msg.MessageID[:]))
		}
	MessageLoop:
		for _, msg := range messages {
			log.Debugf("Fetching.  %d  %s\n", msg.Counter, utils.B58encode(msg.MessageID[:]))
			if startDate+ms.FetchDuration <= time.Now().Unix() { // We worked for long enough
				log.Debugf("FetchDuration timeout\n")
				break MessageLoop
			}
			// Check if message exists
			if ms.DB.MessageExists(msg.MessageID) {
				peerStat.LastPosition = msg.Counter
				log.Debugf("fetch from peer: exists %x %x\n", msg.MessageID, *PubKey)
				continue MessageLoop // Message exists.
			}
			// Add message
			err := ms.FetchPost(url, authtoken, msg.MessageID, msg.ExpireTime)
			if err == nil || err == messagestore.ErrDuplicate {
				// Reduce fetch.ErrorCount when downloads are successful
				log.Debugf("fetch from peer: exists now %x %x\n", msg.MessageID, *PubKey)
				if peerStat.ErrorCount >= 2 {
					peerStat.ErrorCount -= 2
				}
			} else if err != nil {
				peerStat.ErrorCount++
				log.Debugf("Fetch err: %s %s\n", url, err)
				doUpdate = false
				// LastPosition will only advance if future downloads work
				continue MessageLoop
			}
			log.Debugf("fetch from peer: added %x %x\n", msg.MessageID, *PubKey)
			peerStat.LastPosition = msg.Counter
			ms.notifyChan <- true
		}
		if !more { // No more messages
			log.Debugf("Sync done. No More.\n")
			break FetchLoop
		}
		if startDate+ms.FetchDuration <= time.Now().Unix() { // We worked for long enough
			log.Debugf("Sync done. FetchDuration timeout.\n")
			break FetchLoop
		}
	}
	// Write peer update
	log.Debugf("fetch from peer: cycle done %x\n", *PubKey)
	if doUpdate {
		ms.DB.UpdatePeerFetchStat(PubKey, uint64(time.Now().Unix()), peerStat.LastPosition, peerStat.ErrorCount)
	} else {
		ms.DB.UpdatePeerFetchStat(PubKey, peerStat.LastFetch, peerStat.LastPosition, peerStat.ErrorCount)
	}
}

// FetchPost fetches a post from a peer and adds it.
func (ms MessageServer) FetchPost(url, auth string, msgID [message.MessageIDSize]byte, expireRequest uint64) error {
	// Fetch the post
	proto := repproto.New(ms.SocksProxy, "")
	data, err := proto.GetSpecificAuth(url, auth, msgID[:])
	// data, err := proto.GetSpecific(url, msgID[:])
	if err != nil {
		return err
	}
	// Verify and use it
	signheader, err := message.Base64Message(data).GetSignHeader()
	if err != nil {
		log.Debugf("Bad fetch:GetSignHeader: %s\n", err)
		return err
	}
	details, err := message.VerifySignature(*signheader, ms.MinHashCashBits)
	if err != nil {
		log.Debugf("Bad fetch:VerifySignature: %s\n", err)
		return err
	}
	constantRecipientPub, MessageID, err := deferVerify(data)
	if err != nil {
		log.Debugf("Bad fetch:deferVerify: %s\n", err)
		return err
	}
	if *MessageID != details.MsgID {
		log.Debugs("Bad fetch:MessageID\n")
		return ErrBadMessageID
	}
	msgStruct := &structs.MessageStruct{
		MessageID:              *MessageID,
		ReceiverConstantPubKey: *constantRecipientPub,
		SignerPub:              details.PublicKey,
		OneTime:                false,
		Sync:                   false,
		Hidden:                 false,
		ExpireRequest:          expireRequest,
	}
	if message.KeyIsSync(constantRecipientPub) {
		msgStruct.Sync = true
	}
	if message.KeyIsHidden(constantRecipientPub) {
		msgStruct.Hidden = true
	}

	sigStruct := &structs.SignerStruct{
		PublicKey: details.PublicKey,
		Nonce:     details.HashCashNonce,
		Bits:      details.HashCashBits,
	}
	sigStruct.MaxMessagesPosted, sigStruct.MaxMessagesRetained, sigStruct.ExpireTarget = ms.calcLimits(details.HashCashBits)
	return ms.DB.Put(msgStruct, sigStruct, data)
}
