package handlers

import (
	"fmt"
	"io"
	"net/http"
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/repproto/structs"
	"strconv"
)

func deferVerify(d []byte) (*message.Curve25519Key, *[message.MessageIDSize]byte, error) {
	var keyHeader [message.KeyHeaderSize]byte
	msg, err := message.Base64Message(d).Decode()
	if err != nil {
		return nil, nil, err
	}
	messageID := message.CalcMessageID(msg)
	copy(keyHeader[:], msg[message.SignHeaderSize:message.SignHeaderSize+message.KeyHeaderSize])
	_, recKeys, _ := message.ParseKeyHeader(&keyHeader)
	return recKeys.ConstantPubKey, messageID, nil
}

// ProcessPost verifies and adds a post to the database
func (ms MessageServer) ProcessPost(postdata io.ReadCloser, oneTime bool, expireRequest uint64) string {
	data, err := utils.MaxRead(ms.MaxPostSize, postdata)
	if err != nil {
		return "ERROR: Message too big\n"
	}
	if len(data) < ms.MinPostSize {
		return "ERROR: Message too small\n"
	}
	signheader, err := message.Base64Message(data).GetSignHeader()
	if err != nil {
		log.Debugf("Post:GetSignHeader: %s\n", err)
		return "ERROR: Sign Header\n"
	}
	details, err := message.VerifySignature(*signheader, ms.MinHashCashBits)
	if err != nil {
		log.Debugf("Post:VerifySignature: %s\n", err)
		return "ERROR: HashCash\n"
	}
	constantRecipientPub, MessageID, err := deferVerify(data)
	if err != nil {
		log.Debugf("Post:deferVerify: %s\n", err)
		return "ERROR: Verify\n"
	}
	if *MessageID != details.MsgID {
		log.Debugs("Post:MessageID\n")
		return "ERROR: MessageID\n"
	}
	msgStruct := &structs.MessageStruct{
		MessageID:              *MessageID,
		ReceiverConstantPubKey: *constantRecipientPub,
		SignerPub:              details.PublicKey,
		OneTime:                oneTime,
		Sync:                   false,
		Hidden:                 false,
		ExpireRequest:          expireRequest,
	}
	if !oneTime {
		if message.KeyIsSync(constantRecipientPub) {
			msgStruct.Sync = true
		}
	} else {
		msgStruct.Sync = false
	}
	if message.KeyIsHidden(constantRecipientPub) {
		msgStruct.Hidden = true
	}

	sigStruct := &structs.SignerStruct{
		PublicKey: details.PublicKey,
		Nonce:     details.HashCashNonce,
		Bits:      details.HashCashBits,
	}
	sigStruct.MaxMessagesPosted, sigStruct.MaxMessagesRetained, sigStruct.ExpireTarget = ms.calcLimits(details.HashCashBits)
	_, _ = msgStruct, sigStruct
	ms.RandomSleep()
	// err = ms.DB.Put(msgStruct, sigStruct, data)
	err = ms.DB.PutNotify(msgStruct, sigStruct, data, ms.notifyChan)
	ms.RandomSleep()
	if err != nil {
		log.Debugf("Post:MessageDB: %s\n", err)
		return fmt.Sprintf("ERROR: %s\n", err)
	}
	log.Debugf("Post:Added: %x\n", MessageID[:12])
	return "SUCCESS: Connection close\n"
}

// GenPostHandler returns a handler for message posting
func (ms MessageServer) GenPostHandler(oneTime bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var expireRequest uint64
		w.Header().Set("Content-Type", "text/plain; charset=us-ascii")
		if r.Method != "POST" {
			io.WriteString(w, "ERROR: Bad Method\n")
			return
		}
		getValues := r.URL.Query()
		if getValues != nil {
			if v, ok := getValues["expire"]; ok {
				expire, err := strconv.Atoi(v[0])
				if err != nil {
					expireRequest = uint64(expire)
				}
			}
		}
		res := ms.ProcessPost(r.Body, oneTime, expireRequest)
		io.WriteString(w, res)
		return
	}
}
