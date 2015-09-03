package sql

import (
	"database/sql"
	"time"

	"github.com/repbin/repbin/message"
)

// AddToGlobalIndex adds a message to the global index
func (db *MessageDB) AddToGlobalIndex(id uint64) error {
	now := time.Now().Unix()
	// globalIndexAdd: INSERT INTO globalindex (Message, EntryTime) VALUES (?, ?);
	return updateConvertNilError(db.globalIndexAddQ.Exec(id, now))
}

// GetKeyIndex returns the index for key index starting with start and at most count entries
func (db *MessageDB) GetKeyIndex(index *message.Curve25519Key, start int64, count int64) ([][]byte, int, error) {
	return genIndex(db.getKeyIndexQ.Query(toHex(index[:]), start, count))
}

// GetGlobalIndex returns the global index starting with start and at most count entries
func (db *MessageDB) GetGlobalIndex(start, count int64) ([][]byte, int, error) {
	return genIndex(db.getGlobalIndexQ.Query(start, count))
}

func genIndex(rows *sql.Rows, err error) ([][]byte, int, error) {
	var ret [][]byte
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		_, str, err := scanMessage(rows)
		if err != nil {
			return nil, 0, err
		}
		ret = append(ret, str.Encode().Fill())
		i++
	}
	return ret, i, nil
}
