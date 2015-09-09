package sql

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/repbin/repbin/message"
)

// MessageBlob contains a message blob
type MessageBlob struct {
	ID              uint64                         // numeric ID, local unique
	MessageID       [message.MessageIDSize]byte    // MessageID (global unique)
	SignerPublicKey [message.SignerPubKeySize]byte // Signer's public key
	OneTime         bool                           // If message should be burnt after reading
	Data            []byte                         // Message
}

// Convert messageID to directory/filename
func (db *MessageDB) messageIDToFilename(messageID *[message.MessageIDSize]byte) (dirname, filename string) {
	midH := toHex(messageID[:])
	dirname = path.Join(db.dir, midH[0:3], midH[3:6])
	filename = path.Join(dirname, midH[6:])
	return
}

// InsertBlob writes a blob
func (db *MessageDB) InsertBlob(id uint64, messageID *[message.MessageIDSize]byte, signer *[message.SignerPubKeySize]byte, onetime bool, data []byte) error {
	if db.dir == "" {
		return db.InsertBlobDB(id, messageID, signer, onetime, data)
	}
	return db.InsertBlobFS(messageID, data)
}

// InsertBlobDB inserts a blob into the database
func (db *MessageDB) InsertBlobDB(id uint64, messageID *[message.MessageIDSize]byte, signer *[message.SignerPubKeySize]byte, onetime bool, data []byte) error {
	_, err := db.messageBlobInsertQ.Exec(id, toHex(messageID[:]), toHex(signer[:]), onetime, data)
	return err
}

// InsertBlobFS writes a blob to database
func (db *MessageDB) InsertBlobFS(messageID *[message.MessageIDSize]byte, data []byte) error {
	dirname, filename := db.messageIDToFilename(messageID)
	err := ioutil.WriteFile(filename, data, 0600)
	if err != nil {
		if err := os.MkdirAll(dirname, 0700); err != nil {
			return err
		}
		err = ioutil.WriteFile(filename, data, 0600)
	}
	return err
}

// InsertBlobStruct inserts a MessageBlob
func (db *MessageDB) InsertBlobStruct(mb *MessageBlob) error {
	if db.dir == "" {
		return db.InsertBlobDB(mb.ID, &mb.MessageID, &mb.SignerPublicKey, mb.OneTime, mb.Data)
	}
	return db.InsertBlobFS(&mb.MessageID, mb.Data)
}

// DeleteBlob deletes a blob by messageid
func (db *MessageDB) DeleteBlob(messageID *[message.MessageIDSize]byte) error {
	if db.dir == "" {
		return db.DeleteBlobDB(messageID)
	}
	return db.DeleteBlobFS(messageID)
}

// DeleteBlobDB deletes a message blob by MessageID from the database
func (db *MessageDB) DeleteBlobDB(messageID *[message.MessageIDSize]byte) error {
	return updateConvertNilError(db.messageBlobDeleteQ.Exec(toHex(messageID[:])))
}

// DeleteBlobFS deletes a blob from filesystem
func (db *MessageDB) DeleteBlobFS(messageID *[message.MessageIDSize]byte) error {
	_, filename := db.messageIDToFilename(messageID)
	return os.Remove(filename)
}

// GetBlob returns a blob struct
func (db *MessageDB) GetBlob(messageID *[message.MessageIDSize]byte) (*MessageBlob, error) {
	if db.dir == "" {
		return db.GetBlobDB(messageID)
	}
	return db.GetBlobFS(messageID)
}

// GetBlobFS reads a blob from the filesystem
func (db *MessageDB) GetBlobFS(messageID *[message.MessageIDSize]byte) (*MessageBlob, error) {
	var err error
	mb := new(MessageBlob)
	id, st, err := db.SelectMessageByID(messageID)
	if err != nil {
		return nil, err
	}
	_, filename := db.messageIDToFilename(messageID)
	mb.Data, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	mb.MessageID = st.MessageID
	mb.SignerPublicKey = st.SignerPub
	mb.OneTime = st.OneTime
	mb.ID = id
	return mb, nil
}

// GetBlobDB returns the blob identified by messageID from the database
func (db *MessageDB) GetBlobDB(messageID *[message.MessageIDSize]byte) (*MessageBlob, error) {
	var messageIDT, signerPubT string
	var onetimeT int
	mb := new(MessageBlob)
	err := db.messageBlobSelectQ.QueryRow(toHex(messageID[:])).Scan(
		&mb.ID,
		&messageIDT,
		&signerPubT,
		&onetimeT,
		&mb.Data,
	)
	if err != nil {
		return nil, err
	}
	mb.MessageID = *sliceToMessageID(fromHex(messageIDT))
	mb.SignerPublicKey = *sliceToEDPublicKey(fromHex(signerPubT))
	mb.OneTime = intToBool(onetimeT)
	return mb, nil
}
