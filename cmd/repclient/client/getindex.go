package client

import (
	"fmt"
	"os"

	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/repproto"
	"github.com/repbin/repbin/utils/repproto/structs"
)

// CmdIndex returns the index for a private key
func CmdIndex() int {
	var server string
	var err error
	var pubkey *message.Curve25519Key
	var privkey message.Curve25519Key
	var messages []*structs.MessageStruct
	var moreMessages bool
	if OptionsVar.Server == "" {
		getPeers(false)
	}
	if OptionsVar.Outdir != "" && !isDir(OptionsVar.Outdir) {
		log.Fatalf("outdir does not exist or is no directory: %s", OptionsVar.Outdir)
		return 1
	}

	// privkey must be set
	privkeystr := selectPrivKey(OptionsVar.Privkey, GlobalConfigVar.PrivateKey, "tty")
	if privkeystr == "" {
		log.Fatal("Private key missing: --privkey\n")
		return 1
	}
	privT := utils.B58decode(privkeystr)
	copy(privkey[:], privT)
	pubkey = message.CalcPub(&privkey)
	if OptionsVar.Server != "" {
		server = OptionsVar.Server
	}

	proto := repproto.New(OptionsVar.Socksserver, OptionsVar.Server)
	log.Dataf("STATUS (Process):\tLIST\n")

	if server != "" {
		fmt.Println("Specific server")
		messages, moreMessages, err = proto.ListSpecific(server, pubkey[:], privkey[:], OptionsVar.Start, OptionsVar.Count)
	} else {
		fmt.Println("Any server")
		server, messages, moreMessages, err = proto.List(pubkey[:], privkey[:], OptionsVar.Start, OptionsVar.Count)
	}
	if err != nil {
		log.Fatalf("List error: %s\n", err)
		return 1
	}
	if len(messages) > 0 {
		fmt.Print("Index\t\tMessageID\n")
		fmt.Print("------------------------------------------------------------\n")
	}
	for _, msg := range messages {
		log.Dataf("STATUS (MessageList):\t%d %s %s %d %d\n", msg.Counter, utils.B58encode(msg.MessageID[:]), utils.B58encode(msg.SignerPub[:]), msg.PostTime, msg.ExpireTime)
		fmt.Printf("%d\t\t%s\n", msg.Counter, utils.B58encode(msg.MessageID[:]))
	}
	log.Dataf("STATUS (ListResult):\t%d %d %t\n", OptionsVar.Start, OptionsVar.Count, moreMessages)
	if len(messages) > 0 {
		fmt.Print("------------------------------------------------------------\n")
	}
	if moreMessages {
		fmt.Print("More messages may be available\n")
	} else {
		fmt.Print("Listing complete\n")
	}
	if OptionsVar.Outdir != "" {
		log.Dataf("STATUS (Process):\tFETCHMANY\n")
		fmt.Print("Fetching messages...")
		hasErrors := 0
		for _, msg := range messages {
			log.Dataf("STATUS (Fetch):\t%s\n", utils.B58encode(msg.MessageID[:]))
			err = loadStoreMessage(server, msg.MessageID[:], OptionsVar.Outdir+string(os.PathSeparator)+utils.B58encode(msg.MessageID[:])) // use index server for download, store with name messageID
			if err != nil {
				fmt.Print(".F")
				hasErrors++
				log.Dataf("STATUS (FetchError):\t%s\n", utils.B58encode(msg.MessageID[:]))
			} else {
				fmt.Print(".o")
				log.Dataf("STATUS (FetchComplete):\t%s\n", utils.B58encode(msg.MessageID[:]))
			}
		}
		if hasErrors > 0 {
			fmt.Print(". Some errors during download.\n")
			log.Dataf("STATUS (FetchResult):\tERROS %d\n", hasErrors)
		} else {
			fmt.Print(". Download complete.\n")
			log.Dataf("STATUS (FetchResult):\tOK\n")
		}
	}

	log.Sync()
	// server can be set
	return 0
}
