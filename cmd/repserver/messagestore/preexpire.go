package messagestore

import (
	"time"

	"github.com/repbin/repbin/message"
)

// PreExpire expires a message in the next expire run
func (store Store) PreExpire(messageID *[message.MessageIDSize]byte, pubkey *message.Curve25519Key) error {
	_, message, err := store.db.SelectMessageByID(messageID)
	if err != nil {
		return ErrNotFound
	}
	if message.ReceiverConstantPubKey == *pubkey {
		store.db.SetMessageExpireByID(messageID, int64(time.Now().Unix()))
		return nil
	}
	return ErrNotFound
}
