package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/fileback"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/keyproof"
)

// GetGlobalIndex returns the global index
func (ms MessageServer) GetGlobalIndex(w http.ResponseWriter, r *http.Request) {
	var pubKey *message.Curve25519Key
	start := int64(0)
	count := int64(10)
	w.Header().Set("Content-Type", "text/plain; charset=us-ascii")
	getValues := r.URL.Query()
	if getValues != nil {
		if v, ok := getValues["start"]; ok {
			t, err := strconv.Atoi(v[0])
			if err == nil {
				start = int64(t)
			}
		}
		if v, ok := getValues["count"]; ok {
			t, err := strconv.Atoi(v[0])
			if err == nil {
				count = int64(t)
				if count > ms.MaxIndexGlobal {
					count = ms.MaxIndexGlobal
				}
			}
		}
		if v, ok := getValues["auth"]; ok {
			err := ms.AuthenticatePeer(v[0])
			if err != nil {
				io.WriteString(w, fmt.Sprintf("Error: %s", err))
				return
			}
		} else {
			io.WriteString(w, "ERROR: Missing param\n")
			return
		}
	} else {
		io.WriteString(w, "ERROR: Missing param\n")
		return
	}
	messages, found, err := ms.DB.GetGlobalIndex(start, count)
	if err != nil && err != fileback.ErrNoMore {
		log.Debugf("List:GetIndex: %s\n", err)
		log.Debugf("List:GetIndex: Key %x\n", pubKey)
		io.WriteString(w, "ERROR: List failed\n")
		return
	}
	io.WriteString(w, "SUCCESS: Data follows\n")
	for _, msg := range messages {
		io.WriteString(w, "IDX: "+strings.Trim(string(msg), " \t\n\r")+"\n")
	}
	if int64(found) < count {
		io.WriteString(w, "CMD: Exceeded\n")
	} else {
		io.WriteString(w, "CMD: Continue\n")
	}
}

// AuthenticatePeer verifies an existing authStr and matches it to the known peers
func (ms MessageServer) AuthenticatePeer(authStr string) error {
	var counterSig [keyproof.ProofTokenSignedSize]byte
	var auth []byte
	if len(authStr) > keyproof.ProofTokenSignedMax {
		return fmt.Errorf("Bad param")
	}
	auth = utils.B58decode(authStr)
	if len(auth) > keyproof.ProofTokenSignedSize {
		return fmt.Errorf("Bad param")
	}
	copy(counterSig[:], auth)
	ok, timestamp := keyproof.VerifyCounterSig(&counterSig, ms.TokenPubKey)
	if !ok {
		log.Debugs("List:Auth no verify\n")
		return fmt.Errorf("Authentication failed: No match")
	}
	now := time.Now().Unix()
	if enforceTimeOuts && (timestamp < now-ms.MaxAuthTokenAge-ms.MaxTimeSkew || timestamp > now+ms.MaxAuthTokenAge+ms.MaxTimeSkew) {
		return fmt.Errorf("Authentication failed: Timeout")
	}
	return nil
}
