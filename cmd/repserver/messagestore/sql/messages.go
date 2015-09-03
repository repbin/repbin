package sql

import (
	"database/sql"
	"time"

	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils/repproto/structs"
)

// messageNextCounter returns the next message counter for the receiver
func (db *MessageDB) messageNextCounter(receiver *message.Curve25519Key) (uint64, error) {
	var id sql.NullInt64
	db.LockShard(receiver[:])
	defer db.UnlockShard(receiver[:])
	now := time.Now().Unix()
	rec := toHex(receiver[:])
	err := db.nextMessageCounterQ.QueryRow(rec).Scan(&id)
	if err != nil {
		return 0, err
	}
	if id.Valid {
		_, err = db.incrMessageCounterQ.Exec(now, rec)
		if err != nil {
			return 0, err
		}
		return uint64(id.Int64) + 1, nil
	}
	_, err = db.insertMessageCounterQ.Exec(rec, now)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

// ExpireMessageCounter expires all messagecounters (and thus resets key indices) that are older than
// maxAge seconds
func (db *MessageDB) ExpireMessageCounter(maxAge uint64) error {
	expireDate := uint64(time.Now().Unix()) - maxAge
	_, err := db.expireMessageCounterQ.Exec(expireDate)
	return err
}

// InsertMessage inserts a message struct into the database
func (db *MessageDB) InsertMessage(msg *structs.MessageStruct) (uint64, error) {
	var err error
	msg.Counter, err = db.messageNextCounter(&msg.ReceiverConstantPubKey)
	res, err := db.insertMessageQ.Exec(
		msg.Counter,
		toHex(msg.MessageID[:]),
		toHex(msg.ReceiverConstantPubKey[:]),
		toHex(msg.SignerPub[:]),
		msg.PostTime,
		msg.ExpireTime,
		msg.ExpireRequest,
		msg.Distance,
		boolToInt(msg.OneTime),
		boolToInt(msg.Sync),
		boolToInt(msg.Hidden),
	)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, ErrNoModify
	}
	n, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(n), nil
}

// scanAble is an interface that describes both sql.Row as well as sql.Rows (note the s)
type scanAble interface {
	Scan(dest ...interface{}) error
}

func scanMessage(a scanAble) (uint64, *structs.MessageStruct, error) {
	var messageIDT, receiverConstantPubKeyT, signerPubT string
	var oneTimeT, syncT, hiddenT int
	var id uint64
	s := new(structs.MessageStruct)
	if err := a.Scan(
		&id,
		&s.Counter,
		&messageIDT,
		&receiverConstantPubKeyT,
		&signerPubT,
		&s.PostTime,
		&s.ExpireTime,
		&s.ExpireRequest,
		&s.Distance,
		&oneTimeT,
		&syncT,
		&hiddenT,
	); err != nil {
		return 0, nil, err
	}
	s.MessageID = *sliceToMessageID(fromHex(messageIDT))
	s.ReceiverConstantPubKey = *sliceToCurve25519Key(fromHex(receiverConstantPubKeyT))
	s.SignerPub = *sliceToEDPublicKey(fromHex(signerPubT))
	s.OneTime = intToBool(oneTimeT)
	s.Sync = intToBool(syncT)
	s.Hidden = intToBool(hiddenT)
	return id, s, nil
}

// SelectMessageByID returns message data for the messageid
func (db *MessageDB) SelectMessageByID(mid *[message.MessageIDSize]byte) (uint64, *structs.MessageStruct, error) {
	return scanMessage(db.selectMessageQ.QueryRow(toHex(mid[:])))
}

// DeleteMessageByID deletes a message by messageid
func (db *MessageDB) DeleteMessageByID(mid *[message.MessageIDSize]byte) error {
	return updateConvertNilError(db.deleteMessageQ.Exec(toHex(mid[:])))
}

// SetMessageExpireByID changes the expire time for the message identified by messageid
func (db *MessageDB) SetMessageExpireByID(mid *[message.MessageIDSize]byte, expire int64) error {
	return updateConvertNilError(db.updateExpireMessageQ.Exec(expire, toHex(mid[:])))
}

// ExpireMessage contains data necessary for expiring a message
type ExpireMessage struct {
	MessageID [message.MessageIDSize]byte
	SignerPub [message.SignerPubKeySize]byte
}

// SelectMessageExpire returns a list of messages that have expired
func (db *MessageDB) SelectMessageExpire(now int64) ([]ExpireMessage, error) {
	var res []ExpireMessage
	rows, err := db.selectExpireMessageQ.Query(now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var messageIDT, signerPubT string
		err := rows.Scan(&messageIDT, &signerPubT)
		if err != nil {
			return nil, err
		}
		res = append(res, ExpireMessage{
			MessageID: *sliceToMessageID(fromHex(messageIDT)),
			SignerPub: *sliceToEDPublicKey(fromHex(signerPubT)),
		})
	}
	return res, nil
}
