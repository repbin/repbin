package client

import (
	"os"

	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/repproto"
)

//CmdEncrypt encrypts data
func CmdEncrypt() int {
	var privkey, embedConstantPrivKey, embedTemporaryPrivKey, embedConstantPubKey, embedTemporaryPubKey *message.Curve25519Key
	var err error
	var inData, encMessage []byte
	var meta *message.MetaDataSend
	var signKeyPair *message.SignKeyPair
	var removeFile string
	if OptionsVar.Server == "" {
		getPeers(false)
	}
	// 	OptionsVar.anonymous . disable private key and signkey
	if OptionsVar.Anonymous {
		OptionsVar.Signkey = ""
		OptionsVar.Privkey = ""
		GlobalConfigVar.PrivateKey = ""
	}
	mindelay := uint32(OptionsVar.Mindelay)
	maxdelay := uint32(OptionsVar.Maxdelay)

	// Read input data
	maxInData := int64(GlobalConfigVar.BodyLength - (message.Curve25519KeySize * 2) - innerHeader)
	//maxInData := int64(MsgSizeLimit)
	if OptionsVar.Repost {
		maxInData -= utils.RepostHeaderSize - message.KeyHeaderSize - message.SignHeaderSize
	}
	log.Debugf("Size limit: %d\n", maxInData)
	inData, err = inputData(OptionsVar.Infile, maxInData)
	if err != nil {
		log.Fatalf("No input data: %s\n", err)
		return 1
	}

	// Verify list contents
	if OptionsVar.MessageType == message.MsgTypeList {
		err := utils.VerifyListContent(inData)
		if err != nil {
			log.Errorf("Could not verify list input: %s\n", err)
			return 1
		}
	}

	// Select private key to use
	privkeystr := selectPrivKey(OptionsVar.Privkey, GlobalConfigVar.PrivateKey, "")
	// Parse privkey
	if privkeystr != "" {
		privkey = new(message.Curve25519Key)
		copy(privkey[:], utils.B58decode(privkeystr))
	}
	// Generate embedded keypairs
	if OptionsVar.Embedkey {
		if OptionsVar.Notrace || privkey == nil {
			embedConstantPrivKey, err = message.GenLongTermKey(OptionsVar.Hidden, OptionsVar.Sync)
			if err != nil {
				log.Errorf("Key generation error:%s\n", err)
				return 1
			}
		} else {
			embedConstantPrivKey = privkey
		}
		embedConstantPubKey = message.GenPubKey(embedConstantPrivKey)
		embedTemporaryPrivKey, err = message.GenRandomKey()
		if err != nil {
			log.Errorf("Key generation error:%s\n", err)
			return 1
		}
		embedTemporaryPubKey = message.GenPubKey(embedTemporaryPrivKey)
	}
	embedded := utils.EncodeEmbedded(embedConstantPubKey, embedTemporaryPubKey)
	recipientConstantPubKey, recipientTemporaryPubKey := utils.ParseKeyPair(OptionsVar.Recipientkey)

	// Find a signature keypair if we can
	signKeyDir := ""
	if GlobalConfigVar.KeyDir != "" {
		signKeyDir = GlobalConfigVar.KeyDir
	}
	if OptionsVar.Signdir != "" {
		signKeyDir = OptionsVar.Signdir
	}

	if OptionsVar.Signkey != "" {
		var d []byte
		d, err = utils.MaxReadFile(2048, OptionsVar.Signkey)
		if err == nil {
			signKeyPair = new(message.SignKeyPair)
			kp, err := signKeyPair.Unmarshal(d)
			if err != nil {
				log.Errorf("Sign keypair decode error: %s\n", err)
				signKeyPair = nil
			} else {
				signKeyPair = kp
			}
		} else {
			log.Errorf("Sign keypair read error: %s\n", err)
		}
	}
	if signKeyPair == nil && signKeyDir != "" {
		var d []byte
		d, removeFile, err = utils.ReadRandomFile(signKeyDir, 2048)
		if err == nil {
			signKeyPair = new(message.SignKeyPair)
			kp, err := signKeyPair.Unmarshal(d)
			if err != nil {
				log.Errorf("Sign keypair decode error: %s\n", err)
				signKeyPair = nil
				removeFile = ""
			} else {
				signKeyPair = kp
			}
		} else {
			log.Errorf("Sign keypair read error: %s\n", err)
		}
	}

	// Set up sender parameters
	sender := message.Sender{
		Signer:                    signKeyPair,
		SenderPrivateKey:          privkey,
		ReceiveConstantPublicKey:  recipientConstantPubKey,
		ReceiveTemporaryPublicKey: recipientTemporaryPubKey,
		TotalLength:               GlobalConfigVar.BodyLength,
		PadToLength:               GlobalConfigVar.PadToLength,
		HashCashBits:              GlobalConfigVar.MinHashCash,
	}
	// We want encryption output in realtime
	log.Sync()
	inData = append(embedded, inData...)
	if OptionsVar.Repost {
		// Generate a repost-message
		encMessage, meta, err = sender.EncryptRepost(byte(OptionsVar.MessageType), inData)
		if err == nil {
			rph := utils.EncodeRepostHeader(meta.PadKey, mindelay, maxdelay)
			encMessage = append(rph[:], encMessage...)
			log.Datas("STATUS (Process):\tPREPOST\n")
			log.Dataf("STATUS (RepostSettings):\t%d %d\n", mindelay, maxdelay)
		}
	} else {
		// Generate a normal message
		encMessage, meta, err = sender.Encrypt(byte(OptionsVar.MessageType), inData)
		if err == nil {
			log.Datas("STATUS (Process):\tPOST\n")
		}
	}
	if err != nil {
		log.Fatalf("Encryption failed: %s\n", err)
		return 1
	}
	log.Sync()

	// Output. repost is only written to stdout or file
	if OptionsVar.Outfile == "-" || (OptionsVar.Repost && OptionsVar.Outfile == "") {
		err = utils.WriteStdout(encMessage)
		// Display data as necessary
		if err == nil {
			log.Dataf("STATUS (RecPubKey):\t%s\n", utils.B58encode(meta.ReceiverConstantPubKey[:]))
			if OptionsVar.Embedkey {
				log.Dataf("STATUS (EmbedPublicKey):\t%s_%s\n", utils.B58encode(embedConstantPubKey[:]), utils.B58encode(embedTemporaryPubKey[:]))
				log.Dataf("STATUS (EmbedPrivateKey):\t%s_%s\n", utils.B58encode(embedConstantPrivKey[:]), utils.B58encode(embedTemporaryPrivKey[:]))
			}
			if meta.MessageKey != nil {
				log.Dataf("STATUS (ListInput):\tNULL %s %s\n", utils.B58encode(meta.MessageID[:]), utils.B58encode(meta.MessageKey[:]))
				log.Dataf("STATUS (Message):\t%s_%s\n", utils.B58encode(meta.MessageID[:]), utils.B58encode(meta.MessageKey[:]))
			} else {
				log.Dataf("STATUS (ListInput):\tNULL %s NULL\n", utils.B58encode(meta.MessageID[:]))
				log.Dataf("STATUS (MessageID):\t%s\n", utils.B58encode(meta.MessageID[:]))
			}
		}
	} else if OptionsVar.Outfile != "" || OptionsVar.Repost {
		err = utils.WriteNewFile(OptionsVar.Outfile, encMessage)
		// Display data as necessary
		log.Dataf("STATUS (RecPubKey):\t%s\n", utils.B58encode(meta.ReceiverConstantPubKey[:]))
		if err == nil {
			if OptionsVar.Embedkey {
				log.Dataf("STATUS (EmbedPublicKey):\t%s_%s\n", utils.B58encode(embedConstantPubKey[:]), utils.B58encode(embedTemporaryPubKey[:]))
				log.Dataf("STATUS (EmbedPrivateKey):\t%s_%s\n", utils.B58encode(embedConstantPrivKey[:]), utils.B58encode(embedTemporaryPrivKey[:]))
			}
			if meta.MessageKey != nil {
				log.Dataf("STATUS (ListInput):\tNULL %s %s\n", utils.B58encode(meta.MessageID[:]), utils.B58encode(meta.MessageKey[:]))
				log.Dataf("STATUS (Message):\t%s_%s\n", utils.B58encode(meta.MessageID[:]), utils.B58encode(meta.MessageKey[:]))
				log.Printf("Pastebin Address:\t%s_%s\n", utils.B58encode(meta.MessageID[:]), utils.B58encode(meta.MessageKey[:]))
			} else {
				log.Dataf("STATUS (ListInput):\tNULL %s NULL\n", utils.B58encode(meta.MessageID[:]))
				log.Dataf("STATUS (MessageID):\t%s\n", utils.B58encode(meta.MessageID[:]))
				log.Printf("Pastebin Address:\t%s\n", utils.B58encode(meta.MessageID[:]))
			}
		}
	} else {
		// Post to server
		server := OptionsVar.Server
		proto := repproto.New(OptionsVar.Socksserver, OptionsVar.Server, GlobalConfigVar.PasteServers...)
		if server == "" {
			server, err = proto.Post(meta.MessageID[:], encMessage)
		} else {
			err = proto.PostSpecific(server, encMessage)
		}
		sep := "/"
		if server != "" && server[len(server)-1] == '/' {
			sep = ""
		}
		if err == nil {
			if meta.MessageKey != nil {
				log.Dataf("STATUS (URL):\t%s/%s_%s\n", server, utils.B58encode(meta.MessageID[:]), utils.B58encode(meta.MessageKey[:]))
			}
			if OptionsVar.Embedkey {
				log.Dataf("STATUS (EmbedPublicKey):\t%s_%s\n", utils.B58encode(embedConstantPubKey[:]), utils.B58encode(embedTemporaryPubKey[:]))
				log.Dataf("STATUS (EmbedPrivateKey):\t%s_%s\n", utils.B58encode(embedConstantPrivKey[:]), utils.B58encode(embedTemporaryPrivKey[:]))
			}
			if meta.MessageKey != nil {
				log.Dataf("STATUS (ListInput):\t%s %s %s\n", server, utils.B58encode(meta.MessageID[:]), utils.B58encode(meta.MessageKey[:]))
				log.Dataf("STATUS (Message):\t%s_%s\n", utils.B58encode(meta.MessageID[:]), utils.B58encode(meta.MessageKey[:]))
				log.Printf("Pastebin Address:\t%s%s%s_%s\n", server, sep, utils.B58encode(meta.MessageID[:]), utils.B58encode(meta.MessageKey[:]))
			} else {
				log.Dataf("STATUS (ListInput):\t%s %s NULL\n", server, utils.B58encode(meta.MessageID[:]))
				log.Dataf("STATUS (MessageID):\t%s\n", utils.B58encode(meta.MessageID[:]))
				log.Printf("Pastebin Address:\t%s%s%s\n", server, sep, utils.B58encode(meta.MessageID[:]))
			}
		}
	}
	if err != nil {
		log.Fatalf("Output failed: %s\n", err)
		log.Sync()
		return 1
	}
	if removeFile != "" {
		// Operation has been successful, remove signer keyfile (if any)
		os.Remove(removeFile)
	}
	return 0
}
