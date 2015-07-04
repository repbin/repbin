package messagestore

import (
	"time"

	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/repproto/structs"
)

// PreExpire expires a message in the next expire run
func (store Store) PreExpire(messageID *[message.MessageIDSize]byte, pubkey *message.Curve25519Key) error {
	data := store.messages.Index(messageID[:]).GetLast()
	if data == nil || len(data) < MaxMessageSize+structs.MessageStructSize {
		return ErrNotFound
	}
	msgStruct := structs.MessageStructDecode(data[:structs.MessageStructSize])
	if msgStruct.ReceiverConstantPubKey == *pubkey {
		// Add to expire
		expire := &structs.ExpireStruct{
			ExpireTime:   uint64(time.Now().Unix()),
			MessageID:    msgStruct.MessageID,
			SignerPubKey: msgStruct.SignerPub,
		}
		expireTime := (uint64(time.Now().Unix()) / uint64(ExpireRun)) + uint64(ExpireRun)
		return store.expireindex.Index(utils.EncodeUInt64(expireTime)).CreateAppend(expire.Encode().Fill())
	}
	return ErrNotFound
}
