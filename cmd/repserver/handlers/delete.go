package handlers

import (
	"io"
	"net/http"

	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
)

// Delete implements the delete call for messages.
func (ms MessageServer) Delete(w http.ResponseWriter, r *http.Request) {
	var messageID *[message.MessageIDSize]byte
	var privateKey *message.Curve25519Key
	for i := 0; i < 4; i++ {
		ms.RandomSleep() // Let's not give instant gratification here
	}
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
		if v, ok := getValues["privkey"]; ok {
			t := utils.B58decode(v[0])
			if len(t) < message.Curve25519KeySize || len(t) > message.Curve25519KeySize {
				io.WriteString(w, "ERROR: Bad parameter\n")
				return
			}
			privateKey = new(message.Curve25519Key)
			copy(privateKey[:], t)
		}
		if privateKey != nil && messageID != nil {
			publicKey := message.GenPubKey(privateKey)
			err := ms.DB.PreExpire(messageID, publicKey)
			if err == nil {
				log.Errorf("Message censored: %x, asshole.\n", *messageID)
				io.WriteString(w, "SUCCESS: If you want to call it that\n")
				return
			}
			log.Errorf("Message censoring failed! %s\n", err)
		}
	}
	io.WriteString(w, "ERROR: Missing parameters\n")
	return
}
