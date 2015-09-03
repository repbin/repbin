package sql

import (
	"bytes"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/repbin/repbin/utils/repproto/structs"
)

var testBlobMessage = &structs.MessageStruct{
	ReceiverConstantPubKey: *testReceiverPubKey,
	MessageID: *sliceToMessageID([]byte(
		strconv.Itoa(
			int(
				time.Now().Unix(),
			),
		) + "MessageBlob",
	)),
	SignerPub: *sliceToEDPublicKey([]byte(
		strconv.Itoa(
			int(
				time.Now().Unix(),
			),
		) + "SignerBlob",
	)),
	PostTime:      10,
	ExpireTime:    uint64(time.Now().Unix() + 1000),
	ExpireRequest: 291090912,
	Distance:      2,
	OneTime:       false,
	Sync:          true,
	Hidden:        false,
}
var testMessageBlob = &MessageBlob{
	// ID              uint64                         // numeric ID, local unique
	MessageID:       testBlobMessage.MessageID,
	SignerPublicKey: testBlobMessage.SignerPub,
	OneTime:         false,
	Data:            bytes.Repeat([]byte{0x00, 0xff, 0x01, 0x03, 0x07, 0x13, 0x23, 0x40}, 65536),
}

func TestBlobMysql(t *testing.T) {
	dir := "" //path.Join(os.TempDir(), "repbinmsg")
	db, err := New("mysql", "root:root@/repbin", dir, 100)
	if err != nil {
		t.Fatalf("New Mysql: %s", err)
	}
	defer db.Close()
	testMessageBlob.ID, err = db.InsertMessage(testBlobMessage)
	if err != nil {
		t.Errorf("InsertMessage: %s", err)
	}
	err = db.InsertBlobStruct(testMessageBlob)
	if err != nil {
		t.Errorf("InsertBlobStruct: %s", err)
	}
	testBlobRes, err := db.GetBlobDB(&testMessageBlob.MessageID)
	if err != nil {
		t.Errorf("GetBlob: %s", err)
	}

	if testMessageBlob.ID != testBlobRes.ID {
		t.Errorf("ID mismatch: %s != %s", testMessageBlob.ID, testBlobRes.ID)
	}
	if testMessageBlob.MessageID != testBlobRes.MessageID {
		t.Errorf("MessageID mismatch: %x != %x", testMessageBlob.MessageID, testBlobRes.MessageID)
	}
	if testMessageBlob.SignerPublicKey != testBlobRes.SignerPublicKey {
		t.Errorf("SignerPublicKey mismatch: %x != %x", testMessageBlob.SignerPublicKey, testBlobRes.SignerPublicKey)
	}
	if testMessageBlob.OneTime != testBlobRes.OneTime {
		t.Errorf("OneTime mismatch: %b != %b", testMessageBlob.OneTime, testBlobRes.OneTime)
	}
	if !bytes.Equal(testMessageBlob.Data, testBlobRes.Data) {
		t.Errorf("Data mismatch: %d != %d", len(testMessageBlob.Data), len(testBlobRes.Data))
	}

	err = db.DeleteBlobDB(&testMessageBlob.MessageID)
	if err != nil {
		t.Errorf("DeleteBlob: %s", err)
	}
	_, err = db.GetBlobDB(&testMessageBlob.MessageID)
	if err == nil {
		t.Error("GetBlob must fail on deleted message")
	}
}

func TestBlobSQLite(t *testing.T) {
	dir := "" //path.Join(os.TempDir(), "repbinmsg")
	dbFile := path.Join(os.TempDir(), "db.test-blob")
	db, err := New("sqlite3", dbFile, dir, 100)
	if err != nil {
		t.Fatalf("New sqlite3: %s", err)
	}
	defer os.Remove(dbFile)
	defer db.Close()
	testMessageBlob.ID, err = db.InsertMessage(testBlobMessage)
	if err != nil {
		t.Errorf("InsertMessage: %s", err)
	}
	err = db.InsertBlobStruct(testMessageBlob)
	if err != nil {
		t.Errorf("InsertBlobStruct: %s", err)
	}
	testBlobRes, err := db.GetBlobDB(&testMessageBlob.MessageID)
	if err != nil {
		t.Errorf("GetBlob: %s", err)
	}

	if testMessageBlob.ID != testBlobRes.ID {
		t.Errorf("ID mismatch: %s != %s", testMessageBlob.ID, testBlobRes.ID)
	}
	if testMessageBlob.MessageID != testBlobRes.MessageID {
		t.Errorf("MessageID mismatch: %x != %x", testMessageBlob.MessageID, testBlobRes.MessageID)
	}
	if testMessageBlob.SignerPublicKey != testBlobRes.SignerPublicKey {
		t.Errorf("SignerPublicKey mismatch: %x != %x", testMessageBlob.SignerPublicKey, testBlobRes.SignerPublicKey)
	}
	if testMessageBlob.OneTime != testBlobRes.OneTime {
		t.Errorf("OneTime mismatch: %b != %b", testMessageBlob.OneTime, testBlobRes.OneTime)
	}
	if !bytes.Equal(testMessageBlob.Data, testBlobRes.Data) {
		t.Errorf("Data mismatch: %d != %d", len(testMessageBlob.Data), len(testBlobRes.Data))
	}

	err = db.DeleteBlobDB(&testMessageBlob.MessageID)
	if err != nil {
		t.Errorf("DeleteBlob: %s", err)
	}
	_, err = db.GetBlobDB(&testMessageBlob.MessageID)
	if err == nil {
		t.Error("GetBlob must fail on deleted message")
	}
}

func TestBlobDirSQLite(t *testing.T) {
	dir := path.Join(os.TempDir(), "repbinmsg")
	dbFile := path.Join(os.TempDir(), "db.test-blobdir")
	db, err := New("sqlite3", dbFile, dir, 100)
	if err != nil {
		t.Fatalf("New sqlite3: %s", err)
	}
	defer os.Remove(dbFile)
	defer db.Close()
	testMessageBlob.ID, err = db.InsertMessage(testBlobMessage)
	if err != nil {
		t.Errorf("InsertMessage: %s", err)
	}
	err = db.InsertBlobStruct(testMessageBlob)
	if err != nil {
		t.Errorf("InsertBlobStruct: %s", err)
	}
	testBlobRes, err := db.GetBlob(&testMessageBlob.MessageID)
	if err != nil {
		t.Errorf("GetBlob: %s", err)
	}

	if testMessageBlob.ID != testBlobRes.ID {
		t.Errorf("ID mismatch: %s != %s", testMessageBlob.ID, testBlobRes.ID)
	}
	if testMessageBlob.MessageID != testBlobRes.MessageID {
		t.Errorf("MessageID mismatch: %x != %x", testMessageBlob.MessageID, testBlobRes.MessageID)
	}
	if testMessageBlob.SignerPublicKey != testBlobRes.SignerPublicKey {
		t.Errorf("SignerPublicKey mismatch: %x != %x", testMessageBlob.SignerPublicKey, testBlobRes.SignerPublicKey)
	}
	if testMessageBlob.OneTime != testBlobRes.OneTime {
		t.Errorf("OneTime mismatch: %b != %b", testMessageBlob.OneTime, testBlobRes.OneTime)
	}
	if !bytes.Equal(testMessageBlob.Data, testBlobRes.Data) {
		t.Errorf("Data mismatch: %d != %d", len(testMessageBlob.Data), len(testBlobRes.Data))
	}

	err = db.DeleteBlobFS(&testMessageBlob.MessageID)
	if err != nil {
		t.Errorf("DeleteBlob: %s", err)
	}
	_, err = db.GetBlob(&testMessageBlob.MessageID)
	if err == nil {
		t.Error("GetBlob must fail on deleted message")
	}
}
