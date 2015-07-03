package client

import (
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
)

// CmdGenKey generates a long-term public key
func CmdGenKey() int {
	privkey, err := message.GenLongTermKey(OptionsVar.Hidden, OptionsVar.Sync)
	if err != nil {
		log.Errorf("Key generation error:%s\n", err)
		return 1
	}
	pubkey := message.GenPubKey(privkey)
	log.Dataf("STATUS (PrivateKey):\t%s\n", utils.B58encode(privkey[:]))
	// log.Dataf("STATUS (PublicKey):\t%s\n", utils.B58encode(pubkey[:]))
	log.Printf("PRIVATE key: %s\n\n", utils.B58encode(privkey[:]))
	// log.Printf("Public key: %s\n", utils.B58encode(pubkey[:]))
	_ = pubkey
	return 0
}

// CmdGenTempKey generates a temporary key for a given private key
func CmdGenTempKey() int {
	var privkey message.Curve25519Key
	privkeystr := selectPrivKey(OptionsVar.Privkey, GlobalConfigVar.PrivateKey, "tty")
	if privkeystr == "" {
		log.Fatal("No private key given (-privkey)")
		return 1
	}
	copy(privkey[:], utils.B58decode(privkeystr))
	pubkey := message.GenPubKey(&privkey)
	privkeytemp, err := message.GenRandomKey()
	if err != nil {
		log.Errorf("Key generation error:%s\n", err)
		return 1
	}
	pubkeytemp := message.GenPubKey(privkeytemp)
	log.Dataf("STATUS (PrivateKey):\t%s_%s\n", utils.B58encode(privkey[:]), utils.B58encode(privkeytemp[:]))
	log.Dataf("STATUS (PublicKey):\t%s_%s\n", utils.B58encode(pubkey[:]), utils.B58encode(pubkeytemp[:]))
	log.Printf("PRIVATE key: %s_%s\n\n", utils.B58encode(privkey[:]), utils.B58encode(privkeytemp[:]))
	log.Printf("Public key: %s_%s\n", utils.B58encode(pubkey[:]), utils.B58encode(pubkeytemp[:]))
	return 0
}
