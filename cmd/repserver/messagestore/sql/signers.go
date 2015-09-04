package sql

import (
	"database/sql"
	"time"

	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils/repproto/structs"
)

// InsertSigner writes a new signer to database and returns its unique ID
func (db *MessageDB) InsertSigner(signerStruct *structs.SignerStruct) (int64, error) {
	// mdb.signerInsertQ(PublicKey,Nonce,Bits,MaxMessagesPosted,MaxMessagesRetained,ExpireTarget)
	res, err := db.signerInsertQ.Exec(
		toHex(signerStruct.PublicKey[:]),
		toHex(signerStruct.Nonce[:]),
		signerStruct.Bits,
		signerStruct.MaxMessagesPosted,
		signerStruct.MaxMessagesRetained,
		signerStruct.ExpireTarget,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateSigner updates the signer
func (db *MessageDB) UpdateSigner(signerStruct *structs.SignerStruct) error {
	return updateConvertNilError(db.signerUpdateQ.Exec(
		toHex(signerStruct.Nonce[:]),
		signerStruct.Bits,
		signerStruct.MaxMessagesPosted,
		signerStruct.MaxMessagesRetained,
		signerStruct.ExpireTarget,
		toHex(signerStruct.PublicKey[:]),
	))
}

// InsertOrUpdateSigner inserts or updates a signer
func (db *MessageDB) InsertOrUpdateSigner(signerStruct *structs.SignerStruct) error {
	var err error
	if db.driver == "mysql" {
		// Use single query
		return updateConvertNilError(db.signerUpdateInsertQ.Exec(
			toHex(signerStruct.PublicKey[:]),
			toHex(signerStruct.Nonce[:]),
			signerStruct.Bits,
			signerStruct.MaxMessagesPosted,
			signerStruct.MaxMessagesRetained,
			signerStruct.ExpireTarget,
		))
	}
	// Do update, insert on error
	if err = db.UpdateSigner(signerStruct); err == ErrNoModify {
		_, err = db.InsertSigner(signerStruct)
	}
	return err
}

func parseSigner(row *sql.Row) (int64, *structs.SignerStruct, error) {
	var dbID int64
	var pubkeyT, nonceT string
	st := new(structs.SignerStruct)
	if err := row.Scan(
		&dbID,
		&pubkeyT,
		&nonceT,
		&st.Bits,
		&st.MessagesPosted,
		&st.MessagesRetained,
		&st.MaxMessagesPosted,
		&st.MaxMessagesRetained,
		&st.ExpireTarget,
	); err != nil {
		return 0, nil, err
	}
	st.PublicKey = *sliceToSignerPubKey(fromHex(pubkeyT))
	st.Nonce = *sliceToNonce(fromHex(nonceT))
	return dbID, st, nil
}

// SelectSigner returns the signer identified by its public key. Returns unique ID, struct or error
func (db *MessageDB) SelectSigner(pk *[message.SignerPubKeySize]byte) (int64, *structs.SignerStruct, error) {
	return parseSigner(db.signerSelectPublicKeyQ.QueryRow(toHex(pk[:])))
}

// SelectSignerByID returns the signer identified by its public key. Returns unique ID, struct or error
func (db *MessageDB) SelectSignerByID(id int64) (int64, *structs.SignerStruct, error) {
	return parseSigner(db.signerSelectIDQ.QueryRow(id))
}

// AddMessage adds a message to the signer stats
func (db *MessageDB) AddMessage(pk *[message.SignerPubKeySize]byte) error {
	return updateConvertNilError(db.signerAddMessageQ.Exec(toHex(pk[:])))
}

// DelMessage deletes a message from the signer stats
func (db *MessageDB) DelMessage(pk *[message.SignerPubKeySize]byte) error {
	return updateConvertNilError(db.signerDelMessageQ.Exec(toHex(pk[:])))
}

// ExpireSigners expires signers. Returns number of prepared and deleted entries
func (db *MessageDB) ExpireSigners(maxAge int64) (int64, int64, error) {
	var prepared, deleted int64
	var err error
	now := time.Now().Unix()
	res, err := db.signerPrepareExpireQ.Exec(now)
	if err != nil {
		return 0, 0, err
	}
	prepared, err = res.RowsAffected()
	if err != nil {
		return prepared, 0, err
	}
	res, err = db.signerExpireQ.Exec(now - maxAge)
	if err != nil {
		return prepared, 0, err
	}
	deleted, err = res.RowsAffected()
	if err != nil {
		return prepared, deleted, err
	}
	return prepared, deleted, nil
}
