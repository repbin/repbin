package sql

import (
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/repbin/repbin/utils/repproto/structs"
)

var testIndexMessage = &structs.MessageStruct{
	ReceiverConstantPubKey: *testReceiverPubKey,
	MessageID: *sliceToMessageID([]byte(
		strconv.Itoa(
			int(
				time.Now().Unix(),
			),
		) + "MessageINDEX",
	)),
	SignerPub: *sliceToEDPublicKey([]byte(
		strconv.Itoa(
			int(
				time.Now().Unix(),
			),
		) + "Signer",
	)),
	PostTime:      10,
	ExpireTime:    uint64(time.Now().Unix() + 1000),
	ExpireRequest: 291090912,
	Distance:      2,
	OneTime:       false,
	Sync:          true,
	Hidden:        false,
}
var testIndexMessage2 = &structs.MessageStruct{
	ReceiverConstantPubKey: *testReceiverPubKey,
	MessageID: *sliceToMessageID([]byte(
		strconv.Itoa(
			int(
				time.Now().Unix(),
			),
		) + "Message2INDEX",
	)),
	SignerPub: *sliceToEDPublicKey([]byte(
		strconv.Itoa(
			int(
				time.Now().Unix(),
			),
		) + "Signer",
	)),
	PostTime:      11,
	ExpireTime:    uint64(time.Now().Unix() + 1000),
	ExpireRequest: 291090912,
	Distance:      2,
	OneTime:       false,
	Sync:          true,
	Hidden:        false,
}
var testIndexMessage3 = &structs.MessageStruct{
	ReceiverConstantPubKey: *testReceiverPubKey,
	MessageID: *sliceToMessageID([]byte(
		strconv.Itoa(
			int(
				time.Now().Unix(),
			),
		) + "Message3INDEX",
	)),
	SignerPub: *sliceToEDPublicKey([]byte(
		strconv.Itoa(
			int(
				time.Now().Unix(),
			),
		) + "Signer",
	)),
	PostTime:      12,
	ExpireTime:    uint64(time.Now().Unix() + 1000),
	ExpireRequest: 291090912,
	Distance:      2,
	OneTime:       false,
	Sync:          true,
	Hidden:        false,
}

func TestIndexMysql(t *testing.T) {
	if !testing.Short() {
		dir := path.Join(os.TempDir(), "repbinmsg")
		db, err := New("mysql", "root:root@/repbin", dir, 100)
		if err != nil {
			t.Fatalf("New Mysql: %s", err)
		}
		defer db.Close()
		id, err := db.InsertMessage(testIndexMessage)
		if err != nil {
			t.Errorf("InsertMessage: %s", err)
		}
		err = db.AddToGlobalIndex(id)
		if err != nil {
			t.Errorf("AddToGlobalIndex: %s", err)
		}
		id, _ = db.InsertMessage(testIndexMessage2)
		db.AddToGlobalIndex(id)
		id, _ = db.InsertMessage(testIndexMessage3)
		db.AddToGlobalIndex(id)
		l, i, err := db.GetKeyIndex(&testIndexMessage.ReceiverConstantPubKey, 0, 10)
		if err != nil {
			t.Errorf("GetKeyIndex: %s", err)
		}
		if i < 1 {
			t.Error("GetKeyIndex: None found!!!")
		}
		l, i, err = db.GetGlobalIndex(0, 10)
		if err != nil {
			t.Errorf("GetGlobalIndex: %s", err)
		}
		if i < 1 {
			t.Error("GetGlobalIndex: None found!!!")
		}
		// spew.Dump(l)
		_ = l
	}

}

func TestIndexSQLite(t *testing.T) {
	dir := path.Join(os.TempDir(), "repbinmsg")
	dbFile := path.Join(os.TempDir(), "db.test-index")
	db, err := New("sqlite3", dbFile, dir, 100)
	if err != nil {
		t.Fatalf("New sqlite3: %s", err)
	}
	defer os.Remove(dbFile)
	defer db.Close()
	id, err := db.InsertMessage(testIndexMessage)
	if err != nil {
		t.Errorf("InsertMessage: %s", err)
	}
	err = db.AddToGlobalIndex(id)
	if err != nil {
		t.Errorf("AddToGlobalIndex: %s", err)
	}
	id, _ = db.InsertMessage(testIndexMessage2)
	db.AddToGlobalIndex(id)
	id, _ = db.InsertMessage(testIndexMessage3)
	db.AddToGlobalIndex(id)
	l, i, err := db.GetKeyIndex(&testIndexMessage.ReceiverConstantPubKey, 0, 10)
	if err != nil {
		t.Errorf("GetKeyIndex: %s", err)
	}
	if i < 1 {
		t.Error("GetKeyIndex: None found!!!")
	}
	l, i, err = db.GetGlobalIndex(0, 10)
	if err != nil {
		t.Errorf("GetGlobalIndex: %s", err)
	}
	if i < 1 {
		t.Error("GetGlobalIndex: None found!!!")
	}
	_ = l

}
