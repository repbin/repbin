package sql

import (
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/repbin/repbin/utils/repproto/structs"
)

var testReceiverPubKey = sliceToCurve25519Key([]byte(
	strconv.Itoa(
		int(
			0, //time.Now().Unix(),
		),
	) + "Receiver",
))
var testMessage = &structs.MessageStruct{
	ReceiverConstantPubKey: *testReceiverPubKey,
	MessageID: *sliceToMessageID([]byte(
		strconv.Itoa(
			int(
				time.Now().Unix(),
			),
		) + "Message",
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

func TestMessagesMysql(t *testing.T) {
	if !testing.Short() {
		dir := path.Join(os.TempDir(), "repbinmsg")
		db, err := newMySQLForTest(dir, 100)
		if err != nil {
			t.Fatalf("New Mysql: %s", err)
		}
		defer db.Close()
		_, err = db.messageNextCounter(testReceiverPubKey)
		if err != nil {
			t.Errorf("messageNextCounter: %s", err)
		}
		_, err = db.InsertMessage(testMessage)
		if err != nil {
			t.Errorf("InsertMessage: %s", err)
		}
		id2, err := db.messageNextCounter(testReceiverPubKey)
		if err != nil {
			t.Errorf("messageNextCounter: %s", err)
		}
		if id2 < 2 {
			t.Errorf("NextCounter too small: %d < %d", id2, 2)
		}
		_, msgS, err := db.SelectMessageByID(&testMessage.MessageID)
		if err != nil {
			t.Errorf("SelectMessageByID: %s", err)
		}
		if msgS.Counter >= id2 {
			t.Errorf("NextCounter too small: %d < %d", id2, msgS.Counter)
		}
		if msgS.MessageID != testMessage.MessageID {
			t.Error("MessageID no match")
		}
		if msgS.ReceiverConstantPubKey != testMessage.ReceiverConstantPubKey {
			t.Error("ReceiverConstantPubKey no match")
		}
		if msgS.SignerPub != testMessage.SignerPub {
			t.Error("SignerPub no match")
		}
		if msgS.PostTime != testMessage.PostTime {
			t.Error("PostTime no match")
		}
		if msgS.ExpireTime != testMessage.ExpireTime {
			t.Error("ExpireTime no match")
		}
		if msgS.ExpireRequest != testMessage.ExpireRequest {
			t.Error("ExpireRequest no match")
		}
		if msgS.Distance != testMessage.Distance {
			t.Error("Distance no match")
		}
		if msgS.OneTime != testMessage.OneTime {
			t.Error("OneTime no match")
		}
		if msgS.Sync != testMessage.Sync {
			t.Error("Sync no match")
		}
		if msgS.Hidden != testMessage.Hidden {
			t.Error("Hidden no match")
		}
		now := time.Now().Unix()
		expires, err := db.SelectMessageExpire(now)
		if err != nil {
			t.Errorf("SelectMessageExpire: %s", err)
		}
		for _, mid := range expires {
			if mid.MessageID == testMessage.MessageID {
				t.Fatal("Found message at wrong expire!")
			}
		}
		err = db.SetMessageExpireByID(&testMessage.MessageID, now-1)
		if err != nil {
			t.Errorf("SetMessageExpireByID: %s", err)
		}
		expires, err = db.SelectMessageExpire(now)
		if err != nil {
			t.Errorf("SelectMessageExpire: %s", err)
		}
		found := false
		for _, mid := range expires {
			if mid.MessageID == testMessage.MessageID {
				found = true
				if mid.SignerPub != testMessage.SignerPub {
					t.Error("SignerPub on expired message does not match")
				}
			}
		}
		if !found {
			t.Error("Select Expire no result")
		}
		err = db.DeleteMessageByID(&testMessage.MessageID)
		if err != nil {
			t.Errorf("DeleteMessageByID: %s", err)
		}
		_, _, err = db.SelectMessageByID(&testMessage.MessageID)
		if err == nil {
			t.Error("Message was deleted but found anyways")
		}
		if err = db.ExpireMessageCounter(100); err != nil {
			t.Errorf("ExpireMessageCounter: %s", err)
		}
	}
}

func TestMessagesSQLite(t *testing.T) {
	dir := path.Join(os.TempDir(), "repbinmsg")
	dbFile := path.Join(os.TempDir(), "db.test-messages")
	db, err := New("sqlite3", dbFile, dir, 100)
	if err != nil {
		t.Fatalf("New sqlite3: %s", err)
	}
	defer os.Remove(dbFile)
	defer db.Close()
	_, err = db.messageNextCounter(testReceiverPubKey)
	if err != nil {
		t.Errorf("messageNextCounter: %s", err)
	}
	_, err = db.InsertMessage(testMessage)
	if err != nil {
		t.Errorf("InsertMessage: %s", err)
	}
	id2, err := db.messageNextCounter(testReceiverPubKey)
	if err != nil {
		t.Errorf("messageNextCounter: %s", err)
	}
	if id2 < 2 {
		t.Errorf("NextCounter too small: %d < %d", id2, 2)
	}
	_, msgS, err := db.SelectMessageByID(&testMessage.MessageID)
	if err != nil {
		t.Errorf("SelectMessageByID: %s", err)
	}
	if msgS.Counter >= id2 {
		t.Errorf("NextCounter too small: %d < %d", id2, msgS.Counter)
	}
	if msgS.MessageID != testMessage.MessageID {
		t.Error("MessageID no match")
	}
	if msgS.ReceiverConstantPubKey != testMessage.ReceiverConstantPubKey {
		t.Error("ReceiverConstantPubKey no match")
	}
	if msgS.SignerPub != testMessage.SignerPub {
		t.Error("SignerPub no match")
	}
	if msgS.PostTime != testMessage.PostTime {
		t.Error("PostTime no match")
	}
	if msgS.ExpireTime != testMessage.ExpireTime {
		t.Error("ExpireTime no match")
	}
	if msgS.ExpireRequest != testMessage.ExpireRequest {
		t.Error("ExpireRequest no match")
	}
	if msgS.Distance != testMessage.Distance {
		t.Error("Distance no match")
	}
	if msgS.OneTime != testMessage.OneTime {
		t.Error("OneTime no match")
	}
	if msgS.Sync != testMessage.Sync {
		t.Error("Sync no match")
	}
	if msgS.Hidden != testMessage.Hidden {
		t.Error("Hidden no match")
	}
	now := time.Now().Unix()
	expires, err := db.SelectMessageExpire(now)
	if err != nil {
		t.Errorf("SelectMessageExpire: %s", err)
	}
	for _, mid := range expires {
		if mid.MessageID == testMessage.MessageID {
			t.Fatal("Found message at wrong expire!")
		}
	}
	err = db.SetMessageExpireByID(&testMessage.MessageID, now-1)
	if err != nil {
		t.Errorf("SetMessageExpireByID: %s", err)
	}
	expires, err = db.SelectMessageExpire(now)
	if err != nil {
		t.Errorf("SelectMessageExpire: %s", err)
	}
	found := false
	for _, mid := range expires {
		if mid.MessageID == testMessage.MessageID {
			found = true
			if mid.SignerPub != testMessage.SignerPub {
				t.Error("SignerPub on expired message does not match")
			}
		}
	}
	if !found {
		t.Error("Select Expire no result")
	}
	err = db.DeleteMessageByID(&testMessage.MessageID)
	if err != nil {
		t.Errorf("DeleteMessageByID: %s", err)
	}
	_, _, err = db.SelectMessageByID(&testMessage.MessageID)
	if err == nil {
		t.Error("Message was deleted but found anyways")
	}
	if err = db.ExpireMessageCounter(100); err != nil {
		t.Errorf("ExpireMessageCounter: %s", err)
	}
}
