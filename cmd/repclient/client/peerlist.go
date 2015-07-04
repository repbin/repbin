package client

import (
	"math/rand"
	"time"

	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/utils/repproto"
)

// CmdPeerList fetches the peerlist and returns a new configfile
func CmdPeerList() int {
	err := getPeers(true)
	if err != nil {
		log.Fatalf("Peer discovery failed: %s\n", err)
		return 1
	}
	return CmdShowConfig()
}

func getPeers(force bool) error {
	log.Debugf("Peer update called. Last: %d\n", GlobalConfigVar.PeerUpdate)
	if GlobalConfigVar.PeerUpdate == 0 || GlobalConfigVar.PeerUpdate < (time.Now().Unix()-PeerUpdateDuration) {
		log.Debugs("Peer update forced. Prevent by running -peerlist.\n")
		force = true
	}
	if force {
		log.Debugs("Updating peers.\n")
		defer log.Debugs("Peer update done.\n")
		GlobalConfigVar.PeerUpdate = time.Now().Unix()
		server := OptionsVar.Server
		if server == "" && GlobalConfigVar.PasteServers != nil && len(GlobalConfigVar.PasteServers) > 0 { // no server defined, we select one
			servercount := len(GlobalConfigVar.PasteServers)
			rand.Seed(time.Now().UnixNano())
			pos := rand.Int() % servercount
			server = GlobalConfigVar.PasteServers[pos]
		}
		if server == "" {
			server = GlobalConfigVar.BootStrapPeer
		}
		if server == "" {
			return ErrNoPeers
		}
		proto := repproto.New(OptionsVar.Socksserver, OptionsVar.Server, GlobalConfigVar.PasteServers...)
		serverInfo, err := proto.ID(server)
		if err != nil {
			return ErrNoPeers
		}
		if serverInfo.Peers != nil && len(serverInfo.Peers) > 0 {
			GlobalConfigVar.PasteServers = serverInfo.Peers
			GlobalConfigVar.MinHashCash = serverInfo.MinHashCashBits
			err := WriteConfigFile(GlobalConfigVar)
			if err != nil {
				log.Errorf("Error writing config-file: %s\n", err)
			}
		}
	}
	return nil
}
