package client

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/repproto"
)

// CmdDecrypt implements decryption functions
func CmdDecrypt() int {
	var inData, decMessage []byte
	var err error
	var meta *message.MetaDataRecieve
	var privkeystr string
	var nullKey message.Curve25519Key
	if OptionsVar.Server == "" {
		getPeers(false)
	}
	if OptionsVar.Stmdir != "" && !isDir(OptionsVar.Stmdir) {
		log.Fatalf("stmdir does not exist or is no directory: %s\n", OptionsVar.Stmdir)
		return 1
	}

	// Read input data
	maxInData := int64(GlobalConfigVar.BodyLength+message.KeyHeaderSize+message.SignHeaderSize) * 4
	// Read data from stdin or file
	if len(flag.Args()) == 0 {
		inData, err = inputData(OptionsVar.Infile, maxInData)
		if err != nil {
			log.Fatalf("No input data: %s\n", err)
			return 1
		}
	} else {
		// read data from server (Get). CMDLine is  [server/]messageid[_privatekey]
		var messageidcl []byte
		var server string
		server, messageidcl, privkeystr = cmdlineURLparse(flag.Args()...)
		if server == "" {
			server = OptionsVar.Server
		}
		proto := repproto.New(OptionsVar.Socksserver, OptionsVar.Server, GlobalConfigVar.PasteServers...)
		if server == "" {
			server, inData, err = proto.Get(messageidcl)
		} else {
			inData, err = proto.GetSpecific(server, messageidcl)
		}
		if err != nil {
			log.Fatalf("Fetch error: %s\n", err)
			return 1
		}
		log.Dataf("STATUS (FetchServer):\t%s\n", server)
	}
	if len(inData) < (message.KeyHeaderSize+message.SignHeaderSize)*4 {
		log.Fatals("No input data.\n")
		return 1
	}

	// Set up receiver parameters
	receiver := message.Receiver{
		HashCashBits: GlobalConfigVar.MinHashCash,
	}

	if OptionsVar.Senderkey != "" {
		receiver.SenderPublicKey, _ = utils.ParseKeyPair(OptionsVar.Senderkey)
	}

	// Select private key to use
	if OptionsVar.Keymgt < 0 {
		if privkeystr == "" { // might have been set from commandline
			privkeystr = selectPrivKey(OptionsVar.Privkey, GlobalConfigVar.PrivateKey, "tty")
		}
		// Parse privkey
		if privkeystr != "" {
			receiver.ReceiveConstantPrivateKey, receiver.ReceiveTemporaryPrivateKey = utils.ParseKeyPair(privkeystr)
		}
	} else {
		// Register callback, OptionsVar.keymgt == fd
		keyMgtFile, callback := KeyCallBack(OptionsVar.Keymgt)
		defer keyMgtFile.Close()
		receiver.KeyCallBack = callback
	}

	log.Datas("STATUS (Process):\tREAD\n")

	// Decrypt
	decMessage, meta, err = receiver.Decrypt(inData)
	if err != nil {
		log.Fatalf("%s\n", err)
		return 1
	}
	log.Dataf("STATUS (MessageID):\t%s\n", utils.B58encode(meta.MessageID[:]))
	log.Dataf("STATUS (RecPubKey):\t%s\n", utils.B58encode(meta.ReceiveConstantPublicKey[:]))
	log.Dataf("STATUS (SenderPubKey):\t%s\n", utils.B58encode(meta.SenderConstantPublicKey[:]))

	// Get replyKeys
	embedConstant, embedTemporary := utils.DecodeEmbedded(decMessage[:message.Curve25519KeySize*2])
	if *embedConstant != nullKey {
		log.Dataf("STATUS (EmbedPublicKey):\t%s_%s\n", utils.B58encode(embedConstant[:]), utils.B58encode(embedTemporary[:]))
	}
	// If messageType list: print list to DATA
	if meta.MessageType == message.MsgTypeList {
		log.Datas("STATUS (Process):\tLIST\n")
		err := utils.VerifyListContent(decMessage[message.Curve25519KeySize*2:])
		if err != nil {
			log.Fatalf("%s\n", err)
			return 1
		}
		lines := bytes.Split(decMessage[message.Curve25519KeySize*2:], []byte("\n"))
		for _, l := range lines {
			log.Dataf("STATUS (ListItem):\t%s\n", string(l))
		}
		return 0
	}
	decMessage = decMessage[message.Curve25519KeySize*2:]
	//meta.MessageType
	if meta.MessageType == message.MsgTypeRepost {
		log.Datas("STATUS (Process):\tREPOST\n")
		// If messageType repost: get padkey,mindelay,maxdelay. Repad
		padkey, minDelay, maxDelay := utils.DecodeRepostHeader(decMessage[:utils.RepostHeaderSize])
		timePoint := utils.STM(int(minDelay), int(maxDelay))
		log.Dataf("STATUS (STM):\t%d %d %d\n", minDelay, maxDelay, timePoint)
		repostMsgt := message.RePad(decMessage[utils.RepostHeaderSize:], padkey, GlobalConfigVar.BodyLength)
		signHeader := new([message.SignHeaderSize]byte)
		copy(signHeader[:], repostMsgt[:message.SignHeaderSize])
		details, err := message.VerifySignature(*signHeader, GlobalConfigVar.MinHashCash)
		if err != nil {
			log.Fatalf("%s\n", err)
			return 1
		}
		msgIDMsg := message.CalcMessageID(repostMsgt)
		if *msgIDMsg != details.MsgID {
			log.Fatals("MessageID conflict\n")
			return 1
		}
		log.Dataf("STATUS (MessageIDSig):\t%s\n", utils.B58encode(details.MsgID[:]))
		log.Dataf("STATUS (PubKeySig):\t%s\n", utils.B58encode(details.PublicKey[:]))
		log.Dataf("STATUS (NonceSig):\t%x\n", details.HashCashNonce[:])
		log.Dataf("STATUS (BitsSig):\t%d\n", details.HashCashBits)
		decMessage = message.EncodeBase64(repostMsgt)
		if OptionsVar.Stmdir != "" { // Exist/Dir test done early
			filename := fmt.Sprintf("%s%s%d.%s", OptionsVar.Stmdir, string(os.PathSeparator), timePoint, utils.B58encode(msgIDMsg[:]))
			log.Dataf("STATUS (STMFile):\t%s\n", filename)
			err := utils.WriteNewFile(filename, decMessage)
			if err != nil {
				log.Fatalf("%s\n", err)
				return 1
			}
			return 0
		}
	}
	err = outputData(OptionsVar.Outfile, decMessage)
	if err != nil {
		log.Fatalf("Output failed: %s\n", err)
		return 1
	}
	return 0
}
