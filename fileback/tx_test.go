package fileback

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestTX(t *testing.T) {
	indexName := make([]byte, 20)
	binary.BigEndian.PutUint64(indexName, uint64(time.Now().UnixNano()))
	testdata1 := []byte("Testing one")
	testdata2 := []byte("Testing two")
	// testdata3 := []byte("Testing three")
	// testdata4 := []byte("Testing four")
	// testdata5 := []byte("Testing five")
	list := NewRolling("/tmp/testing/rollingtx", 510, 3, byte(' '), []byte("\n"), 1).Index(indexName)
	if list.Exists() {
		t.Error("list should not exist")
	}
	entries := list.Entries()
	if entries != 0 {
		t.Error("Entries failed")
	}
	err := list.Create(testdata1)
	if err != nil {
		t.Fatalf("Create failed: %s", err)
	}
	if !list.Exists() {
		t.Error("list must exist")
	}
	entries = list.Entries()
	if entries != 1 {
		t.Error("Entries failed 0")
	}
	err = list.Update(func(tx Tx) error {
		return tx.Append(testdata2)
	})
	if err != nil {
		t.Fatalf("TX Append failed: %s", err)
	}
	entries = list.Entries()
	if entries != 2 {
		t.Error("Entries failed 0")
	}
}

func TestInterface(t *testing.T) {
	intFunc := func(t ListIndex) {
		return
	}
	indexName := make([]byte, 20)
	binary.BigEndian.PutUint64(indexName, uint64(time.Now().UnixNano()))
	list := NewRolling("/tmp/testing/interface", 510, 3, byte(' '), []byte("\n"), 1).Index(indexName)
	intFunc(list)
}
