package messagestore

import (
	"time"

	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils/repproto/structs"
)

// MessageExists returns true if the message exists
func (store Store) MessageExists(messageID [message.MessageIDSize]byte) bool {
	return store.db.MessageKnown(&messageID)
}

// Put stores a message in the message store WITHOUT notifying the notify backend
func (store Store) Put(msgStruct *structs.MessageStruct, signerStruct *structs.SignerStruct, message []byte) error {
	// Check if message exists
	if store.db.MessageKnown(&msgStruct.MessageID) {
		return ErrDuplicate
	}
	// Check if signer exists, load last from signer
	_, signerLoaded, _ := store.db.SelectSigner(&signerStruct.PublicKey)
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
	err := store.db.InsertOrUpdateSigner(signerStruct)
	if err != nil {
		log.Errorf("messagestore, update signer: %s", err)
	}
	// Write message. Message is prefixed by encoded messageStruct
	storeID, err := store.db.InsertMessage(msgStruct)
	if err != nil {
		log.Errorf("messagestore, write message (DB): %s", err)
		return err
	}
	store.db.LearnMessage(&msgStruct.MessageID)
	err = store.db.InsertBlob(storeID, &msgStruct.MessageID, &signerStruct.PublicKey, msgStruct.OneTime, message)
	if err != nil {
		log.Errorf("messagestore, write message (Blob): %s", err)
		return err
	}
	if !msgStruct.OneTime && msgStruct.Sync {
		err := store.db.AddToGlobalIndex(storeID)
		if err != nil {
			log.Errorf("messagestore, globalindex append: %s", err)
		}
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
