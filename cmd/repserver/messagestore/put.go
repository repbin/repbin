package messagestore

import (
	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/fileback"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/repproto/structs"
	"time"
)

// MessageExists returns true if the message exists
func (store Store) MessageExists(messageID [message.MessageIDSize]byte) bool {
	if store.messages.Index(messageID[:]).Exists() {
		return true
	}
	return false
}

// Put stores a message in the message store WITHOUT notifying the notify backend
func (store Store) Put(msgStruct *structs.MessageStruct, signerStruct *structs.SignerStruct, message []byte) error {
	// Check if message exists
	if store.MessageExists(msgStruct.MessageID) {
		return ErrDuplicate
	}
	// Check if signer exists, load last from signer
	signerLoaded := structs.SignerStructDecode(store.signers.Index(signerStruct.PublicKey[:]).GetLast())
	if signerLoaded != nil {
		// Old signer is better
		if signerLoaded.Bits > signerStruct.Bits {
			signerStruct.Bits = signerLoaded.Bits
			signerStruct.Nonce = signerLoaded.Nonce
			signerStruct.MaxMessagesPosted = signerLoaded.MaxMessagesPosted
			signerStruct.MaxMessagesRetained = signerLoaded.MaxMessagesRetained
		}
		// Update post data
		signerStruct.MessagesPosted = signerLoaded.MessagesPosted
		signerStruct.MessagesRetained = signerLoaded.MessagesRetained
	}
	// Verify signer
	if signerStruct.MessagesPosted >= signerStruct.MaxMessagesPosted {
		return ErrPostLimit
	}
	if signerStruct.MessagesRetained >= signerStruct.MaxMessagesRetained {
		// Retention is changed by expiring messages
		return ErrPostLimit
	}
	msgStruct.PostTime = uint64(time.Now().Unix())
	msgStruct.ExpireTime = uint64(uint64(time.Now().Unix()) + signerStruct.ExpireTarget)
	if msgStruct.ExpireTime < msgStruct.ExpireRequest {
		msgStruct.ExpireTime = msgStruct.ExpireRequest
	}
	// Update signer
	err := store.signers.Index(signerStruct.PublicKey[:]).CreateAppend(signerStruct.Encode()) // Ignore errors on signer update. They don't matter much
	if err != nil {
		log.Errorf("messagestore, update signer: %s", err)
	}
	// Write message. Message is prefixed by encoded messageStruct
	msgStructEncoded := msgStruct.Encode().Fill()
	err = store.messages.Index(msgStruct.MessageID[:]).Create(append(msgStructEncoded, message...))
	if err != nil {
		log.Errorf("messagestore, write message: %s", err)
		return err
	}
	if !msgStruct.OneTime && msgStruct.Sync {
		// TX: Add to global index
		if store.keyindex.Index(globalindex).Exists() {
			err := store.keyindex.Index(globalindex).Update(func(tx fileback.Tx) error { // Ignore errors
				return updateKeyIndex(tx, msgStruct, msgStructEncoded)
			})
			if err != nil {
				log.Errorf("messagestore, globalindex append: %s", err)
			}
		} else {
			err := store.keyindex.Index(globalindex).CreateAppend(msgStructEncoded)
			if err != nil {
				log.Errorf("messagestore, globalindex createAppend: %s", err)
			}
		}
	}
	// TX: Add to key index
	if store.keyindex.Index(msgStruct.ReceiverConstantPubKey[:]).Exists() {
		err := store.keyindex.Index(msgStruct.ReceiverConstantPubKey[:]).Update(func(tx fileback.Tx) error { // Ignore errors
			return updateKeyIndex(tx, msgStruct, msgStructEncoded)
		})
		if err != nil {
			log.Errorf("messagestore, receiverPub append: %s", err)
		}
	} else {
		err := store.keyindex.Index(msgStruct.ReceiverConstantPubKey[:]).CreateAppend(msgStructEncoded)
		if err != nil {
			log.Errorf("messagestore, receiverPub createAppend: %s", err)
		}
	}
	// Add to expire
	expire := &structs.ExpireStruct{
		ExpireTime:   msgStruct.ExpireTime,
		MessageID:    msgStruct.MessageID,
		SignerPubKey: msgStruct.SignerPub,
	}
	expireTime := (msgStruct.ExpireTime / uint64(ExpireRun)) + uint64(ExpireRun)
	err = store.expireindex.Index(utils.EncodeUInt64(expireTime)).CreateAppend(expire.Encode().Fill())
	if err != nil {
		log.Errorf("messagestore, record expire: %s", err)
	}
	return nil
}

// PutNotify runs Put and notifies the backend if no error occured
func (store Store) PutNotify(msgStruct *structs.MessageStruct, signerStruct *structs.SignerStruct, message []byte, notifyChan chan bool) error {
	err := store.Put(msgStruct, signerStruct, message)
	if err != nil {
		return err
	}
	notifyChan <- true
	return nil
}

// updateKeyIndex updates a key index by writing the new struct with an increased counter
func updateKeyIndex(tx fileback.Tx, msgStruct *structs.MessageStruct, msgStructEncoded structs.MessageStructEncoded) error { // Ignore errors
	prev := tx.GetLast()
	if prev == nil {
		log.Errors("messagestore, globalindex getLast: NIL")
		err := tx.Append(msgStructEncoded)
		return err
	}
	prevStr := structs.MessageStructDecode(prev)
	if prevStr == nil {
		log.Errors("messagestore, globalindex decode: NIL")
		err := tx.Append(msgStructEncoded)
		return err
	}
	msgStruct.Counter = prevStr.Counter + 1 // Increase counter
	msgStructEncoded = msgStruct.Encode().Fill()
	err := tx.Append(msgStructEncoded)
	return err
}
