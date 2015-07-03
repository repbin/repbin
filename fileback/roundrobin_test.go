package fileback

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

func TestNewRoundRobin(t *testing.T) {
	indexName := make([]byte, 20)
	binary.BigEndian.PutUint64(indexName, uint64(time.Now().UnixNano()))
	testdata := []byte("Testing one")
	testdata2 := []byte("Testing two")
	testdata3 := []byte("Testing three")
	testdata4 := []byte("Testing four")
	testdata5 := []byte("Testing five")
	list := NewRoundRobin("/tmp/testing/roundrobin", 510, 4, byte(' '), []byte("\n"), 1).Index(indexName)
	if list.Exists() {
		t.Error("list should not exist")
	}
	entries := list.Entries()
	if entries != 0 {
		t.Error("Entries failed")
	}
	err := list.Create(testdata)
	if err != nil {
		t.Fatalf("Create failed: %s", err)
	}
	data := list.GetLast()
	if data == nil {
		t.Error("GetLast failed")
	}
	if !bytes.Equal(data[:len(testdata)], testdata) {
		t.Error("GetLast bad read")
	}
	entries = list.Entries()
	if entries != 1 {
		t.Error("Entries failed")
	}
	err = list.Append(testdata2)
	if err != nil {
		t.Fatalf("Append failed: %s", err)
	}
	entries = list.Entries()
	if entries != 2 {
		t.Error("Entries failed")
	}
	data = list.GetLast()
	if data == nil {
		t.Error("GetLast failed")
	}
	if !bytes.Equal(data[:len(testdata2)], testdata2) {
		t.Error("GetLast bad read")
	}
	data, _ = list.ReadEntry(0)
	if data == nil {
		t.Error("ReadEntry failed")
	}
	if !bytes.Equal(data[:len(testdata)], testdata) {
		t.Error("ReadEntry bad read")
	}
	data, _ = list.ReadEntry(1)
	if data == nil {
		t.Error("ReadEntry failed")
	}
	if !bytes.Equal(data[:len(testdata2)], testdata2) {
		t.Error("ReadEntry bad read")
	}
	if list.EntryExists(2) {
		t.Error("entry should not exist")
	}
	if !list.EntryExists(1) {
		t.Error("entry should exist")
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
	err = list.Append(testdata3)
	if err != nil {
		t.Fatalf("Append failed: %s", err)
	}
	entries = list.Entries()
	if entries != 3 {
		t.Error("Entries failed")
	}
	err = list.Append(testdata4)
	if err != nil {
		t.Fatalf("Append failed: %s", err)
	}
	entries = list.Entries()
	if entries != 4 {
		t.Error("Entries failed")
	}
	err = list.Append(testdata5)
	if err != nil {
		t.Fatalf("Append failed: %s", err)
	}
	entries = list.Entries()
	if entries != 4 {
		t.Error("Append 5 Entries failed")
	}
	data, _ = list.ReadEntry(0)
	if data == nil {
		t.Error("ReadEntry failed")
	}
	if !bytes.Equal(data[:len(testdata5)], testdata5) {
		t.Error("ReadEntry bad read")
	}

	err = list.Append(testdata)
	if err != nil {
		t.Fatalf("Append failed: %s", err)
	}
	entries = list.Entries()
	if entries != 4 {
		t.Error("Entries failed")
	}
	data, _ = list.ReadEntry(0)
	if data == nil {
		t.Error("ReadEntry failed")
	}
	if !bytes.Equal(data[:len(testdata5)], testdata5) {
		t.Error("ReadEntry bad read")
	}
	data, _ = list.ReadEntry(1)
	if data == nil {
		t.Error("ReadEntry failed")
	}
	if !bytes.Equal(data[:len(testdata)], testdata) {
		t.Error("ReadEntry bad read")
	}
	data = list.GetLast()
	if data == nil {
		t.Error("GetLast failed")
	}
	if !bytes.Equal(data[:len(testdata)], testdata) {
		t.Error("GetLast bad read")
	}
	entryList, err := list.ReadRange(0, 4)
	if err != nil {
		t.Errorf("ReadRange failed: %s", err)
	}
	if !bytes.Equal(entryList[0][:len(testdata)], testdata) {
		t.Errorf("ReadRange bad read: %s", entryList[0])
	}
	if !bytes.Equal(entryList[1][:len(testdata3)], testdata3) {
		t.Errorf("ReadRange bad read: %s", entryList[1])
	}
	if !bytes.Equal(entryList[2][:len(testdata4)], testdata4) {
		t.Errorf("ReadRange bad read: %s", entryList[2])
	}
	if !bytes.Equal(entryList[3][:len(testdata5)], testdata5) {
		t.Errorf("ReadRange bad read: %s", entryList[3])
	}
	entryList, err = list.ReadRange(1, 4)
	if err != nil {
		t.Errorf("ReadRange failed: %s", err)
	}
	if !bytes.Equal(entryList[0][:len(testdata3)], testdata3) {
		t.Errorf("ReadRange bad read: %s", entryList[0])
	}
	if !bytes.Equal(entryList[1][:len(testdata4)], testdata4) {
		t.Errorf("ReadRange bad read: %s", entryList[1])
	}
	if !bytes.Equal(entryList[2][:len(testdata5)], testdata5) {
		t.Errorf("ReadRange bad read: %s", entryList[2])
	}
	if !bytes.Equal(entryList[3][:len(testdata)], testdata) {
		t.Errorf("ReadRange bad read: %s", entryList[3])
	}
	list.Delete()
	if list.Exists() {
		t.Error("list should not exist")
	}
	err = list.Create(testdata)
	if err != nil {
		t.Fatalf("Create failed: %s", err)
	}
	if !list.Exists() {
		t.Error("list should exist")
	}
	list.Truncate()
	if !list.Exists() {
		t.Error("list should exist")
	}
	err = list.Create(testdata)
	if err == nil {
		t.Fatal("Truncated list may not be re-created")
	}
}
