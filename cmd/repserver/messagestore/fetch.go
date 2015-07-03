package messagestore

import (
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils/repproto/structs"
)

// Fetch a message from storage, delete if it is a one-time message
func (store Store) Fetch(messageID *[message.MessageIDSize]byte) ([]byte, error) {
	data := store.messages.Index(messageID[:]).GetLast()
	if data == nil || len(data) < MaxMessageSize+structs.MessageStructSize {
		return nil, ErrNotFound
	}
	msgStruct := structs.MessageStructDecode(data[:structs.MessageStructSize])
	if msgStruct.OneTime {
		store.messages.Index(messageID[:]).Truncate()
		signerLoaded := structs.SignerStructDecode(store.signers.Index(msgStruct.SignerPub[:]).GetLast())
		if signerLoaded != nil {
			signerLoaded.MessagesRetained--
			store.signers.Index(msgStruct.SignerPub[:]).Append(signerLoaded.Encode().Fill())
		}
	}
	return data[structs.MessageStructSize:], nil
}
