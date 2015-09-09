package handlers

import (
	"io"
	"net/http"
	"time"

	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/keyproof"
)

// GetNotify receives notifications.
func (ms MessageServer) GetNotify(w http.ResponseWriter, r *http.Request) {
	var proof [keyproof.ProofTokenSize]byte
	w.Header().Set("Content-Type", "text/plain; charset=us-ascii")
	getValues := r.URL.Query()
	if getValues != nil {
		if v, ok := getValues["auth"]; ok {
			if len(v[0]) > keyproof.ProofTokenMax {
				io.WriteString(w, "ERROR: Bad Param\n")
				return
			}
			auth := utils.B58decode(v[0])
			if auth == nil || len(auth) > keyproof.ProofTokenSize {
				io.WriteString(w, "ERROR: Bad Param\n")
				return
			}
			copy(proof[:], auth)
			ok, timeStamp, senderPubKey := keyproof.VerifyProofToken(&proof, ms.TokenPubKey)
			if !ok {
				io.WriteString(w, "ERROR: Authentication failure\n")
				if senderPubKey == nil {
					log.Errorf("VerifyProofToken failed: (proof) %x\n", proof)
				} else {
					log.Errorf("VerifyProofToken failed: (pubkey) %x\n", *senderPubKey)
				}
				return
			}
			// verify that we know the peer
			if !ms.PeerKnown(senderPubKey) {
				io.WriteString(w, "ERROR: Bad peer\n")
				log.Errorf("Notify, bad peer: %x\n", *senderPubKey)
				return
			}
			now := CurrentTime()
			// Test too old, too young
			if enforceTimeOuts && (now > timeStamp+DefaultAuthTokenAge+ms.MaxTimeSkew || now < timeStamp-DefaultAuthTokenAge-ms.MaxTimeSkew) {
				io.WriteString(w, "ERROR: Authentication expired\n")
				log.Errorf("VerifyProofToken replay by %x\n", *senderPubKey)
				return
			}
			ok, signedToken := keyproof.CounterSignToken(&proof, ms.TokenPubKey, ms.TokenPrivKey)
			if !ok {
				io.WriteString(w, "ERROR: Authentication failure\n")
				return
			}
			ms.DB.UpdatePeerAuthToken(senderPubKey, signedToken)
			log.Debugf("Notified by %x\n", *senderPubKey)
			io.WriteString(w, "SUCCESS: Notified\n")
			return
		}
	}
	io.WriteString(w, "ERROR: Missing Param\n")
	return
}
