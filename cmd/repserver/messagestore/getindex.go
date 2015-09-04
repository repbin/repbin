package messagestore

import (
	"github.com/repbin/repbin/message"
)

// GetIndex returns the index for a key
func (store Store) GetIndex(index *message.Curve25519Key, start int64, count int64) ([][]byte, int, error) {
	ret, i, err := store.db.GetKeyIndex(index, start, count)
	if err != nil {
		return nil, 0, ErrNotFound
	}
	return ret, i, nil
}

// GetGlobalIndex returns the global index
func (store Store) GetGlobalIndex(start int64, count int64) ([][]byte, int, error) {
	ret, i, err := store.db.GetGlobalIndex(start, count)
	if err != nil {
		return nil, 0, ErrNotFound
	}
	return ret, i, nil
}
