package fileback

import (
	"bytes"
	"encoding/binary"
	// "fmt"
	"testing"
	"time"
)

func TestNewContinue(t *testing.T) {
	indexName := make([]byte, 20)
	binary.BigEndian.PutUint64(indexName, uint64(time.Now().UnixNano()))
	testdata := []byte("Testing one")
	testdata2 := []byte("Testing two")
	list := NewContinue("/tmp/testing/paged", 510, 10, byte(' '), []byte("\n"), 1).Index(indexName)
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
	indices := list.Indices()
	if indices == nil || len(indices) != 1 {
		t.Error("Indices() wrong count")
	} else {
		if !bytes.Equal(indices[0], indexName) {
			t.Error("Wrong indices listed")
		}
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
	list.Delete()
	if list.Exists() {
		t.Error("list must NOT exist after Delete")
	}
	err = list.Create(testdata)
	if err != nil {
		t.Fatalf("Create failed: %s", err)
	}
	if !list.Exists() {
		t.Error("list must exist")
	}
	err = list.Append(testdata2)
	if err != nil {
		t.Fatalf("Append failed: %s", err)
	}
	if list.EntryExists(2) {
		t.Error("entry should not exist")
	}
	if !list.EntryExists(1) {
		t.Error("entry should exist")
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
	data, _ = list.ReadEntry(1)
	if data == nil {
		t.Error("ReadEntry failed")
	}
	if !bytes.Equal(data[:len(testdata2)], testdata2) {
		t.Error("ReadEntry bad read")
	}
	entries = list.Entries()
	if entries != 2 {
		t.Error("Entries failed")
	}
	err = list.Change(0, testdata)
	if err != nil {
		t.Fatalf("Change failed: %s", err)
	}
	data, _ = list.ReadEntry(0)
	if data == nil {
		t.Error("ReadEntry failed")
	}
	if !bytes.Equal(data[:len(testdata)], testdata) {
		t.Error("ReadEntry bad read 2")
	}
	data, _ = list.ReadEntry(1)
	if data == nil {
		t.Error("ReadEntry failed")
	}
	if !bytes.Equal(data[:len(testdata2)], testdata2) {
		t.Error("ReadEntry bad read 3")
	}
	pos, entry := list.ReadRandom()
	if pos == -1 || entry == nil {
		t.Error("ReadRandom failed")
	}
	if pos == 0 && !bytes.Equal(entry[:len(testdata)], testdata) {
		t.Errorf("ReadRandom wrong data 1: %s", string(entry))
	}
	if pos == 1 && !bytes.Equal(data[:len(testdata2)], testdata2) {
		t.Error("ReadRandom wrong data 2")
	}
	if pos > 1 {
		t.Error("Bad ReadRandom position")
	}
	entryList, err := list.ReadRange(0, 2)
	if err != nil {
		t.Error("ReadRange failed")
	}
	if !bytes.Equal(entryList[0][:len(testdata)], testdata) {
		t.Error("ReadRange wrong data")
	}
	if !bytes.Equal(entryList[1][:len(testdata2)], testdata2) {
		t.Error("ReadRange wrong data")
	}
	// Test partial read/write
	part, err := list.ReadEntryPart(0, 3, 8)
	if err != nil {
		t.Errorf("ReadPart failed: %s", err)
	}
	if !bytes.Equal(part, testdata[3:11]) {
		t.Errorf("ReadPart wrong data: '%s' != '%s' ", part, testdata[3:11])
	}
	err = list.ChangePart(0, []byte("PART"), 2)
	if err != nil {
		t.Errorf("ChangePart failed: %s", err)
	}
	part, err = list.ReadEntryPart(0, 2, 4)
	if err != nil {
		t.Errorf("ReadPart failed: %s", err)
	}
	if !bytes.Equal(part, []byte("PART")) {
		t.Errorf("ReadPart wrong data: '%s' != '%s' ", part, []byte("PART"))
	}
	_ = part
	list.Delete()
}
