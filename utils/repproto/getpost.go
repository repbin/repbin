// Package repproto implements the repbin protocol wrappers.
package repproto

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/keyauth"
	"github.com/repbin/repbin/utils/listparse"
	"github.com/repbin/repbin/utils/repproto/structs"
	"github.com/repbin/repbin/utils/socks"
)

// Version of this release
const Version = "0.0.1 very alpha"

var (
	// ErrNoServers is returned if no servers are known
	ErrNoServers = errors.New("rep: No servers")
	// ErrNoResponse is returned if no response could be gotten
	ErrNoResponse = errors.New("rep: No response")
	// ErrBadProto is returned if the protocol is not adhered to
	ErrBadProto = errors.New("rep: Bad proto")
	// ErrNoMoreServers is returned if no more servers are available
	ErrNoMoreServers = errors.New("rep: Servers exhausted")
	// ErrPrivKey is returned if authentication is required but the private key was not given
	ErrPrivKey = errors.New("rep: Private key required but missing")
)

// Proto implements the protocol wrappers
type Proto struct {
	SocksServer string   // the socks server
	Servers     []string // Repbin servers
	// ServerSelector replaces server selection. first string is messageID, remaining are available servers.
	// ServerSelector MUST return ErrNoMoreServers when servers are exhausted or an infinite loop will result
	ServerSelector func([]byte, ...string) (string, error)
	// SelectorReset resets server selection
	SelectorReset func()
	selectorPerm  []int
	selectorPos   int
}

func init() {
	socks.AcceptNoSocks = false
}

// New creates a proto wrapper
func New(socksserver string, firstServer string, servers ...string) *Proto {
	tservers := make([]string, 0, 1)
	if firstServer != "" {
		tservers = append(tservers, firstServer)
	}
	p := &Proto{
		SocksServer: socksserver,
		Servers:     append(tservers, servers...),
	}
	p.selectorReset()
	return p
}

func (proto *Proto) selectorReset() {
	if proto.SelectorReset != nil {
		proto.SelectorReset()
		return
	}
	sc := len(proto.Servers)
	if sc == 1 {
		// Do not permutate single server
		proto.selectorPos = 0
		proto.selectorPerm = []int{0}
		return
	}
	// Generate permutation
	rand.Seed(time.Now().UnixNano())
	proto.selectorPerm = rand.Perm(sc)
	proto.selectorPos = 0
	return
}

// selectServer calls the server select or does round-robin (pre-permutation)
func (proto *Proto) selectServer(messageID []byte) (string, error) {
	if proto.ServerSelector != nil {
		return proto.ServerSelector(messageID, proto.Servers...)
	}
	if len(proto.Servers) == 0 || proto.Servers[0] == "" {
		return "", ErrNoServers
	}
	if proto.selectorPos >= len(proto.Servers) {
		proto.selectorReset()
		return "", ErrNoMoreServers
	}
	server := proto.Servers[proto.selectorPerm[proto.selectorPos]]
	proto.selectorPos++
	return server, nil
}

// constructURL safe URL construction fix. first element may not be ""
func constructURL(parts ...string) string {
	var url string
	qmU := byte('?')
	slU := byte('/')
	eqU := byte('=')
	amU := byte('&')
	query := false
	for i, p := range parts {
		if i == 0 {
			url = p
			if strings.LastIndex(p, "?") != -1 {
				query = true
			}
			continue
		}
		if p == "" {
			continue
		}
		lurl := len(url) - 1
		if lurl <= 0 {
			continue
		}
		if !query {
			if p[0] == qmU {
				url += p
				query = true
				continue
			}
			if url[lurl] == slU && p[0] == slU {
				if len(p) > 1 {
					url += p[1:] // prevent duplication
				}
				// do nothing if it's just a /
			} else if url[lurl] != slU && p[0] != slU {
				url += string(slU) + p // fill missing
			} else {
				url += p // nothing to fix
			}
			if strings.LastIndex(p, "?") != -1 {
				query = true
			}
		} else {
			// if both terminate with =, remove one
			if url[lurl] == eqU && p[0] == eqU {
				if len(p) > 1 {
					url += p[1:]
				}
			} else if url[lurl] == eqU || p[0] == eqU {
				// either terminates with =, add
				url += p
			} else if url[lurl] == amU && p[0] == amU {
				// both terminate with &, remove one
				if len(p) > 1 {
					url += p[1:]
				}
			} else if url[lurl] == amU || p[0] == amU {
				url += p
			} else {
				url += string(amU) + p
			}
		}
	}
	return url
}

// Parse error takes the body of a response and checks the first line for error/succes, returns the remaining body
func parseError(body []byte) ([]byte, error) {
	l := bytes.SplitN(body, []byte("\n"), 2)
	if l == nil || len(l) == 0 {
		return nil, ErrNoResponse
	}
	if len(l[0]) > 8 && string(l[0][:8]) == "SUCCESS:" {
		if len(l) == 2 {
			return l[1], nil
		}
		return nil, nil
	}
	if len(l[0]) > 6 && string(l[0][:6]) == "ERROR:" {
		return nil, fmt.Errorf("Server error: %s", string(l[0][7:]))
	}
	return nil, ErrBadProto
}

// Post a message
func (proto *Proto) Post(messageID []byte, message []byte) (string, error) {
	var err error
	maxLoops := len(proto.Servers)
	for {
		server, selecterr := proto.selectServer(messageID)
		if selecterr == ErrNoMoreServers {
			if err == nil {
				return "", ErrNoMoreServers
			}
			// return last error when servers are all done
			return "", err
		}
		err = proto.PostSpecific(server, message)
		// Return on success
		if err == nil {
			return server, nil
		}
		if maxLoops == 0 {
			// Catches one late. Only to prevent faulty Proto.ServerSelector implementations
			return "", ErrNoMoreServers
		}
		maxLoops--
		// Continue through servers on error
	}
}

// PostSpecific posts a message to a specific server
func (proto *Proto) PostSpecific(server string, message []byte) error {
	body, err := socks.Proxy(proto.SocksServer).LimitPostBytes(constructURL(server, "/post"), "text/text", message, 512000)
	if err != nil {
		return err
	}
	_, err = parseError(body) // we do not care about the body
	return err
}

// Get a message
func (proto *Proto) Get(messageID []byte) (string, []byte, error) {
	var err error
	var body []byte
	maxLoops := len(proto.Servers)
	for {
		server, selecterr := proto.selectServer(messageID)
		if selecterr == ErrNoMoreServers {
			if err == nil {
				return "", nil, ErrNoMoreServers
			}
			// return last error when servers are all done
			return "", nil, err
		}
		body, err = proto.GetSpecific(server, messageID)
		// Return on success
		if err == nil {
			return server, body, nil
		}
		if maxLoops == 0 {
			// Catches one late. Only to prevent faulty Proto.ServerSelector implementations
			return "", nil, ErrNoMoreServers
		}
		maxLoops--
		// Continue through servers on error
	}
}

// GetSpecific fetches a message from a specific server
func (proto *Proto) GetSpecific(server string, messageID []byte) ([]byte, error) {
	messageIDenc := utils.B58encode(messageID)
	body, err := socks.Proxy(proto.SocksServer).LimitGet(constructURL(server, "/fetch", "?messageid=", messageIDenc), 512000)
	if err != nil {
		return nil, err
	}
	return parseError(body)
}

// ServerInfo public server info
type ServerInfo struct {
	Time            int64
	AuthPubKey      string
	AuthChallenge   string
	MaxPostSize     int64    // Maximum post size
	MinPostSize     int      // Minimum post size
	MinHashCashBits byte     // Minimum hashcash bits required
	Peers           []string // Peers of the server, if any
}

// ID returns the ID of a specific server
func (proto *Proto) ID(server string) (*ServerInfo, error) {
	body, err := socks.Proxy(proto.SocksServer).LimitGet(constructURL(server, "/id"), 4096)
	if err != nil {
		return nil, err
	}
	si := new(ServerInfo)
	err = json.Unmarshal(body, si)
	if err != nil {
		return nil, err
	}
	return si, nil
}

// Auth creates an authentication for server and privKey
func (proto *Proto) Auth(server string, privKey []byte) (string, error) {
	var challenge [keyauth.ChallengeSize]byte
	var secret [keyauth.PrivateKeySize]byte
	info, err := proto.ID(server)
	if err != nil {
		return "", err
	}
	challengeS := utils.B58decode(info.AuthChallenge)
	copy(challenge[:], challengeS)
	copy(secret[:], privKey[:])
	answer := keyauth.Answer(&challenge, &secret)
	return utils.B58encode(answer[:]), nil
}

// List messages for pubKey
func (proto *Proto) List(pubKey, privKey []byte, start, count int) (server string, messages []*structs.MessageStruct, more bool, err error) {
	maxLoops := len(proto.Servers)
	for {
		server, selecterr := proto.selectServer(pubKey)
		if selecterr == ErrNoMoreServers {
			if err == nil {
				return "", nil, false, ErrNoMoreServers
			}
			// return last error when servers are all done
			return "", nil, false, err
		}
		messages, more, err = proto.ListSpecific(server, pubKey, privKey, start, count)
		// Return on success
		if err == nil {
			return server, messages, more, nil
		}
		if maxLoops == 0 {
			// Catches one late. Only to prevent faulty Proto.ServerSelector implementations
			return "", nil, false, ErrNoMoreServers
		}
		maxLoops--
		// Continue through servers on error
	}
}

// ListSpecific lists the messages for pubKey from a specific server
func (proto *Proto) ListSpecific(server string, pubKey, privKey []byte, start, count int) (messages []*structs.MessageStruct, more bool, err error) {
	var authStr string
	var myPubKey message.Curve25519Key
	copy(myPubKey[:], pubKey)
	if message.KeyIsHidden(&myPubKey) {
		if privKey == nil {
			return nil, false, ErrPrivKey
		}
		auth, err := proto.Auth(server, privKey)
		if err != nil {
			return nil, false, err
		}
		authStr = "&auth=" + auth
	}
	url := constructURL(server, "/keyindex?key=", utils.B58encode(pubKey[:]), "&start=", strconv.Itoa(start), "count=", strconv.Itoa(count), authStr)
	body, err := socks.Proxy(proto.SocksServer).LimitGet(url, 512000)
	if err != nil {
		return nil, false, err
	}
	return parseListResponse(body)
}

func parseListResponse(body []byte) (messages []*structs.MessageStruct, more bool, err error) {
	listbody, err := parseError(body)
	if err != nil {
		return nil, false, err
	}
	msg, last, err := listparse.ReadMessageList(listbody)
	if err != nil {
		return nil, false, err
	}
	if string(last) == "CMD: Continue" {
		return msg, true, nil
	} else if string(last) == "CMD: Exceeded" {
		return msg, false, nil
	}
	return nil, false, ErrBadProto
}

// Notify a server
func (proto *Proto) Notify(server, auth string) error {
	body, err := socks.Proxy(proto.SocksServer).LimitGet(constructURL(server, "/notify?auth=", auth), 4096)
	if err != nil {
		return err
	}
	_, err = parseError(body)
	return err
}

// GetGlobalIndex returns the global index of a server
func (proto *Proto) GetGlobalIndex(server, auth string, start, count int) (messages []*structs.MessageStruct, more bool, err error) {
	url := constructURL(server, "/globalindex?auth=", auth, "&start=", strconv.Itoa(start), "count=", strconv.Itoa(count))
	body, err := socks.Proxy(proto.SocksServer).LimitGet(url, 5242880)
	if err != nil {
		return nil, false, err
	}
	return parseListResponse(body)
}

// GetSpecificAuth fetches a message from a specific server using authentication
func (proto *Proto) GetSpecificAuth(server, auth string, messageID []byte) ([]byte, error) {
	messageIDenc := utils.B58encode(messageID)
	body, err := socks.Proxy(proto.SocksServer).LimitGet(constructURL(server, "/fetch", "?messageid=", messageIDenc, "&auth=", auth), 512000)
	if err != nil {
		return nil, err
	}
	return parseError(body)
}
