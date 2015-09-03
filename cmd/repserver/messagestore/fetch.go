package messagestore

import "github.com/repbin/repbin/message"

// Fetch a message from storage, delete if it is a one-time message
func (store Store) Fetch(messageID *[message.MessageIDSize]byte) ([]byte, error) {
	mb, err := store.db.GetBlob(messageID)
	if err != nil {
		return nil, ErrNotFound
	}
	if mb.OneTime {
		store.db.DelMessage(&mb.SignerPublicKey)
		store.db.DeleteMessageByID(&mb.MessageID)
		store.db.DeleteBlob(&mb.MessageID)
	}
	return mb.Data, nil
}
