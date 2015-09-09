package sql

import (
	"time"

	"github.com/repbin/repbin/message"
)

// LearnMessage records a message to be known
func (db *MessageDB) LearnMessage(mid *[message.MessageIDSize]byte) error {
	now := time.Now().Unix()
	return updateConvertNilError(db.messageExistInsertQ.Exec(toHex(mid[:]), now))
}

// MessageKnown returns true if we know a message already
func (db *MessageDB) MessageKnown(mid *[message.MessageIDSize]byte) bool {
	var messageID string
	err := db.messageExistSelectQ.QueryRow(toHex(mid[:])).Scan(&messageID)
	if err != nil || messageID == "" {
		return false
	}
	return true
}

// ForgetMessages returns true if we know a message already
func (db *MessageDB) ForgetMessages(expireTime int64) error {
	return updateConvertNilError(db.messageExistExpireQ.Exec(expireTime))
}
