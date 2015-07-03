package handlers

import (
	"io"
	"net/http"
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/fileback"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/keyauth"
	"strconv"
	"strings"
	"time"
)

// GetKeyIndex returns the index for a key
func (ms MessageServer) GetKeyIndex(w http.ResponseWriter, r *http.Request) {
	var pubKey *message.Curve25519Key
	var auth []byte
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
				if count > ms.MaxIndexKey {
					count = ms.MaxIndexKey
				}
			}
		}
		if v, ok := getValues["key"]; ok {
			if len(v[0]) > message.Curve25519KeySize*10 {
				io.WriteString(w, "ERROR: Bad param\n")
				return
			}
			t := utils.B58decode(v[0])
			if len(t) != message.Curve25519KeySize {
				io.WriteString(w, "ERROR: Bad param\n")
				return
			}
			pubKey = new(message.Curve25519Key)
			copy(pubKey[:], t)
			if v, ok := getValues["auth"]; ok {
				if len(v[0]) > keyauth.AnswerSize*10 {
					io.WriteString(w, "ERROR: Bad param\n")
					return
				}
				auth = utils.B58decode(v[0])
				if len(auth) != keyauth.AnswerSize {
					io.WriteString(w, "ERROR: Bad param\n")
					return
				}
			}
		} else {
			io.WriteString(w, "ERROR: Missing param\n")
			return
		}
	} else {
		io.WriteString(w, "ERROR: Missing param\n")
		return
	}
	if pubKey == nil {
		io.WriteString(w, "ERROR: Missing param\n")
		return
	}
	if message.KeyIsHidden(pubKey) {
		if auth == nil {
			log.Debugs("List:Auth missing\n")
			io.WriteString(w, "ERROR: Authentication required\n")
			return
		}
		answer := [keyauth.AnswerSize]byte{}
		copy(answer[:], auth)
		now := uint64(time.Now().Unix() + ms.TimeSkew)
		if !keyauth.VerifyTime(&answer, now, ms.TimeGrace) {
			log.Debugs("List:Auth timeout\n")
			io.WriteString(w, "ERROR: Authentication failed: Timeout\n")
			return
		}
		privK := [32]byte(*ms.authPrivKey)
		testK := [32]byte(*pubKey)
		if !keyauth.Verify(&answer, &privK, &testK) {
			log.Debugs("List:Auth no verify\n")
			io.WriteString(w, "ERROR: Authentication failed: No Match\n")
			return
		}
	}
	messages, found, err := ms.DB.GetIndex(pubKey, start, count)
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
