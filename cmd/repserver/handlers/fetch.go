package handlers

import (
	"fmt"
	"io"
	"net/http"
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
)

// Fetch returns a single message
func (ms MessageServer) Fetch(w http.ResponseWriter, r *http.Request) {
	var messageID *[message.MessageIDSize]byte
	w.Header().Set("Content-Type", "text/plain; charset=us-ascii")
	getValues := r.URL.Query()
	if getValues != nil {
		if v, ok := getValues["messageid"]; ok {
			t := utils.B58decode(v[0])
			if len(t) < message.MessageIDSize || len(t) > message.MessageIDSize {
				io.WriteString(w, "ERROR: Bad parameter\n")
				return
			}
			messageID = new([message.MessageIDSize]byte)
			copy(messageID[:], t)
		}
		if ms.HubOnly {
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
		}
	}
	if messageID == nil {
		io.WriteString(w, "ERROR: Missing parameter\n")
		return
	}
	data, err := ms.DB.Fetch(messageID)
	if err != nil {
		log.Debugf("Fetch: %s\n", err)
		io.WriteString(w, "ERROR: No data")
		return
	}
	io.WriteString(w, "SUCCESS: Data follows\n")
	_, err = w.Write(data)
	if err != nil {
		log.Debugf("Write: %s\n", err)
		return
	}
	log.Debugs("Fetch OK\n")
	return
}
