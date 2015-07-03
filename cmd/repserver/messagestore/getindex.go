package messagestore

import (
	"github.com/repbin/repbin/message"
)

// GetIndex returns the index for a key
func (store Store) GetIndex(index *message.Curve25519Key, start int64, count int64) ([][]byte, int, error) {
	if !store.keyindex.Index(index[:]).Exists() {
		return nil, 0, ErrNotFound
	}
	ret, err := store.keyindex.Index(index[:]).ReadRange(start, count)
	return ret, len(ret), err
}

// GetGlobalIndex returns the global index
func (store Store) GetGlobalIndex(start int64, count int64) ([][]byte, int, error) {
	if !store.keyindex.Index(globalindex).Exists() {
		return nil, 0, ErrNotFound
	}
	ret, err := store.keyindex.Index(globalindex).ReadRange(start, count)
	return ret, len(ret), err
}
