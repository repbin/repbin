package fileback

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

func TestNewRolling(t *testing.T) {
	indexName := make([]byte, 20)
	binary.BigEndian.PutUint64(indexName, uint64(time.Now().UnixNano()))
	testdata1 := []byte("Testing one")
	testdata2 := []byte("Testing two")
	testdata3 := []byte("Testing three")
	testdata4 := []byte("Testing four")
	testdata5 := []byte("Testing five")
	list := NewRolling("/tmp/testing/rolling", 510, 3, byte(' '), []byte("\n"), 1).Index(indexName)
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
	data := list.GetLast()
	if data == nil {
		t.Error("GetLast failed")
	}
	if !bytes.Equal(data[:len(testdata1)], testdata1) {
		t.Error("GetLast bad read")
	}
	// Append
	err = list.Append(testdata2)
	if err != nil {
		t.Fatalf("Append failed: %s", err)
	}
	entries = list.Entries()
	if entries != 2 {
		t.Error("Entries failed 1")
	}
	data = list.GetLast()
	if data == nil {
		t.Error("GetLast failed")
	}
	if !bytes.Equal(data[:len(testdata2)], testdata2) {
		t.Error("GetLast bad read")
	}
	if list.EntryExists(2) {
		t.Error("entry should not exist")
	}
	if !list.EntryExists(1) {
		t.Error("entry should exist")
	}
	// Append
	err = list.Append(testdata3)
	if err != nil {
		t.Fatalf("Append failed: %s", err)
	}
	entries = list.Entries()
	if entries != 3 {
		t.Error("Entries failed 2")
	}
	data = list.GetLast()
	if data == nil {
		t.Error("GetLast failed")
	}
	if !bytes.Equal(data[:len(testdata3)], testdata3) {
		t.Error("GetLast bad read")
	}
	// Append
	err = list.Append(testdata4)
	if err != nil {
		t.Fatalf("Append failed: %s", err)
	}
	entries = list.Entries()
	if entries != 4 {
		t.Error("Entries failed 3")
	}
	data = list.GetLast()
	if data == nil {
		t.Error("GetLast failed")
	}
	if !bytes.Equal(data[:len(testdata4)], testdata4) {
		t.Error("GetLast bad read")
	}
	// Append
	err = list.Append(testdata5)
	if err != nil {
		t.Fatalf("Append failed: %s", err)
	}
	entries = list.Entries()
	if entries != 5 {
		t.Errorf("Entries failed 4: %d", entries)
	}
	data = list.GetLast()
	if data == nil {
		t.Error("GetLast failed")
	}
	if !bytes.Equal(data[:len(testdata5)], testdata5) {
		t.Error("GetLast bad read")
	}
	entry, _ := list.ReadEntry(0)
	if entry == nil {
		t.Error("Read entry 0 failed")
	}
	if !bytes.Equal(entry[:len(testdata1)], testdata1) {
		t.Errorf("GetLast bad read: 0.")
	}
	entry, _ = list.ReadEntry(1)
	if entry == nil {
		t.Error("Read entry 1 failed")
	}
	if !bytes.Equal(entry[:len(testdata2)], testdata2) {
		t.Errorf("GetLast bad read: 1. %s", string(entry))
	}
	entry, _ = list.ReadEntry(2)
	if entry == nil {
		t.Error("Read entry 2 failed")
	}
	if !bytes.Equal(entry[:len(testdata3)], testdata3) {
		t.Errorf("GetLast bad read: 2. %s", string(entry))
	}
	entry, _ = list.ReadEntry(3)
	if entry == nil {
		t.Error("Read entry 3 failed")
	}
	if !bytes.Equal(entry[:len(testdata4)], testdata4) {
		t.Errorf("GetLast bad read: 3. %s", string(entry))
	}
	entry, _ = list.ReadEntry(4)
	if entry == nil {
		t.Error("Read entry 4 failed")
	}
	if !bytes.Equal(entry[:len(testdata5)], testdata5) {
		t.Errorf("GetLast bad read: 4. %s", string(entry))
	}
	// Change
	err = list.Change(0, testdata2)
	if err != nil {
		t.Fatalf("Change failed: %s", err)
	}
	entry, _ = list.ReadEntry(0)
	if entry == nil {
		t.Error("ChangeRead entry 4 failed")
	}
	if !bytes.Equal(entry[:len(testdata2)], testdata2) {
		t.Errorf("ChangeGetLast bad read: 4. %s", string(entry))
	}
	err = list.Change(4, testdata2)
	if err != nil {
		t.Fatalf("Change failed: %s", err)
	}
	entry, _ = list.ReadEntry(4)
	if entry == nil {
		t.Error("ChangeRead entry 4 failed")
	}
	if !bytes.Equal(entry[:len(testdata2)], testdata2) {
		t.Errorf("ChangeGetLast bad read: 4. %s", string(entry))
	}
	countEntries := list.Entries()
	if countEntries != 5 {
		t.Error("Entries count failed")
	}
	// ReadRandom
	_, entry = list.ReadRandom()
	if err != nil {
		t.Fatalf("ReadRandom failed: %s", err)
	}
	if entry == nil {
		t.Error("ReadRandom failed")
	}
	// Delete
	list.Delete()
	if list.Exists() {
		t.Error("File may not exist after delete")
	}
	// Create
	err = NewRolling("/tmp/testing/rolling", 510, 3, byte(' '), []byte("\n"), 1).Index(indexName).CreateAppend(testdata1)
	if err != nil {
		t.Fatalf("AppendCreate failed: %s", err)
	}
	if !list.Exists() {
		t.Error("AppendCreate failed")
	}
	list = NewRolling("/tmp/testing/rolling", 510, 3, byte(' '), []byte("\n"), 1).Index(indexName)
	if !list.Exists() {
		t.Error("NewRolling failed")
	}
	err = list.Append(testdata3)
	if err != nil {
		t.Fatalf("Append failed: %s", err)
	}
	entry, _ = list.ReadEntry(0)
	if entry == nil {
		t.Error("Read entry 0-2 failed")
	}
	if !bytes.Equal(entry[:len(testdata1)], testdata1) {
		t.Errorf("GetLast bad read: 0-2. %s", string(entry))
	}
	entry, _ = list.ReadEntry(1)
	if entry == nil {
		t.Error("Read entry 1-2 failed")
	}
	if !bytes.Equal(entry[:len(testdata3)], testdata3) {
		t.Errorf("GetLast bad read: 1-2. %s", string(entry))
	}
	// Truncate
	list.Truncate()
	if !list.Exists() {
		t.Error("Must exist after truncate")
	}
	err = list.Create(testdata3)
	if err == nil {
		t.Fatal("Create after Truncate must fail")
	}
}
