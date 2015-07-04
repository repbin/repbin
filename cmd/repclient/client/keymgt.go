package client

import (
	"os"
	"strconv"
	"strings"

	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
)

// KeyCallBack implements a callback function to request keys from file-descriptor
func KeyCallBack(keyMgtFd int) (*os.File, func(*message.Curve25519Key) *message.Curve25519Key) {
	knownKeys := make(map[message.Curve25519Key]message.Curve25519Key)
	fd := os.NewFile(uintptr(keyMgtFd), "fd/"+strconv.Itoa(keyMgtFd))
	return fd, func(pubkey *message.Curve25519Key) *message.Curve25519Key {
		// KeyCallBack func(*Curve25519Key) *Curve25519Key
		log.Sync()
		if v, ok := knownKeys[*pubkey]; ok { // Return from cache if we can
			return &v
		}
		b := make([]byte, 120)
		log.Dataf("STATUS (KeyMGTRequest):\t%s\n", utils.B58encode(pubkey[:]))
		log.Sync()
		n, _ := fd.Read(b)
		if n == 0 {
			log.Datas("STATUS (KeyMGT):\tREAD FAILURE\n")
			return nil
		}
		log.Datas("STATUS (KeyMGT):\tREAD DONE\n")
		k1, k2 := utils.ParseKeyPair(strings.Trim(string(b[:n]), " \t\r\n"))
		if k1 != nil {
			pub1 := message.GenPubKey(k1) // Add to cache
			knownKeys[*pub1] = *k1
		}
		if k2 != nil {
			pub2 := message.GenPubKey(k2) // Add to cache
			knownKeys[*pub2] = *k2
		}
		if v, ok := knownKeys[*pubkey]; ok { // Return from cache if we can
			return &v
		}
		return nil
	}
}
