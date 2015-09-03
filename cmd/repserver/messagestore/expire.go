package messagestore

import (
	"mutesrc/log"
	"time"
)

// ExpireFromFS expires data based on filesystem last change
func (store Store) ExpireFromFS() {
	store.ExpireFromIndex(0)
}

// ExpireFromIndex reads the expire index and expires messages as they are recorded
func (store Store) ExpireFromIndex(cycles int) {
	// ExpireRun
	delMessages, err := store.db.SelectMessageExpire(time.Now().Unix())
	if err != nil {
		log.Errorf("ExpireFromIndex, SelectMessageExpire: %s", err)
		return
	}
	for _, msg := range delMessages {
		err := store.db.DeleteBlob(&msg.MessageID)
		if err != nil {
			log.Errorf("ExpireFromIndex, DeleteBlob: %s %x", err, msg.MessageID)
			continue
		}
		err = store.db.DeleteMessageByID(&msg.MessageID)
		if err != nil {
			log.Errorf("ExpireFromIndex, DeleteMessageByID: %s %x", err, msg.MessageID)
		}
	}
	_, _, err = store.db.ExpireSigners(MaxAgeSigners)
	if err != nil {
		log.Errorf("ExpireFromIndex, ExpireSigners: %s", err)
	}
	err = store.db.ExpireMessageCounter(MaxAgeRecipients)
	if err != nil {
		log.Errorf("ExpireFromIndex, ExpireMessageCounter: %s", err)
	}
}
