package fileback

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

func TestNewBlob(t *testing.T) {
	indexName := make([]byte, 20)
	binary.BigEndian.PutUint64(indexName, uint64(time.Now().UnixNano()))
	testdata := []byte("Testing one")
	testdata2 := []byte("Testing two")
	list := NewBlob("/tmp/testing/blob", 510, 10).Index(indexName)
	if list.Exists() {
		t.Error("list should not exist")
	}
	err := list.Create(testdata)
	if err != nil {
		t.Fatalf("Create failed: %s", err)
	}
	if !list.Exists() {
		t.Error("list must exist")
	}
	entries := list.Entries()
	if entries != 1 {
		t.Error("Entries failed")
	}
	data := list.GetLast()
	if data == nil {
		t.Error("GetLast failed")
	}
	if !bytes.Equal(data[:len(testdata)], testdata) {
		t.Error("GetLast bad read")
	}

	err = list.Append(testdata2)
	if err == nil {
		t.Error("Append MUST fail")
	}

	err = list.Change(0, testdata2)
	if err != nil {
		t.Fatalf("Change failed: %s", err)
	}
	data, _ = list.ReadEntry(0)
	if data == nil {
		t.Error("ReadEntry failed")
	}
	if !bytes.Equal(data[:len(testdata2)], testdata2) {
		t.Error("ReadEntry bad read")
	}

	err = list.Change(0, testdata)
	if err != nil {
		t.Fatalf("Change failed: %s", err)
	}

	data, _ = list.ReadEntry(1)
	if data == nil {
		t.Error("ReadEntry MUST fail")
	}

	list.Truncate()
	data = list.GetLast()
	if data != nil {
		t.Error("GetLast after Truncate must fail")
	}
	if !list.Exists() {
		t.Error("list must exist after Truncate")
	}
	entries = list.Entries()
	if entries != 0 {
		t.Error("Entries failed")
	}
	err = list.Create(testdata)
	if err == nil {
		t.Fatal("Creating truncated must fail")
	}
	list.Delete()
	if list.Exists() {
		t.Error("list must NOT exist after Delete")
	}
}
