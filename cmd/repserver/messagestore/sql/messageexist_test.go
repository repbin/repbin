package sql

import (
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

var learnMessageID = sliceToMessageID([]byte(
	strconv.Itoa(
		int(
			time.Now().Unix(),
		),
	) + "Message",
))

func newMySQLForTest(dir string, shards int) (*MessageDB, error) {
	var url string
	// MySQL in Travis CI doesn't have a password
	if os.Getenv("TRAVIS") == "true" {
		url = "root@/repbin"
	} else {
		url = "root:root@/repbin"
	}
	return New("mysql", url, dir, shards)
}

func TestLearnMysql(t *testing.T) {
	if !testing.Short() {
		dir := path.Join(os.TempDir(), "repbinmsg")
		db, err := newMySQLForTest(dir, 100)
		if err != nil {
			t.Fatalf("New Mysql: %s", err)
		}
		defer db.Close()
		if db.MessageKnown(learnMessageID) {
			t.Error("Non-existent message found")
		}
		err = db.LearnMessage(learnMessageID)
		if err != nil {
			t.Fatalf("LearnMessage: %s", err)
		}
		if !db.MessageKnown(learnMessageID) {
			t.Error("Existent message not found")
		}
		err = db.ForgetMessages(time.Now().Unix())
		if err != nil {
			t.Fatalf("ForgetMessages: %s", err)
		}
		if db.MessageKnown(learnMessageID) {
			t.Error("Expired message found")
		}
	}
}

func TestLearnSQLite(t *testing.T) {
	dir := path.Join(os.TempDir(), "repbinmsg")
	dbFile := path.Join(os.TempDir(), "db.test-learn")
	db, err := New("sqlite3", dbFile, dir, 100)
	if err != nil {
		t.Fatalf("New sqlite3: %s", err)
	}
	defer os.Remove(dbFile)
	defer db.Close()
	if db.MessageKnown(learnMessageID) {
		t.Error("Non-existent message found")
	}
	err = db.LearnMessage(learnMessageID)
	if err != nil {
		t.Fatalf("LearnMessage: %s", err)
	}
	if !db.MessageKnown(learnMessageID) {
		t.Error("Existent message not found")
	}
	err = db.ForgetMessages(time.Now().Unix())
	if err != nil {
		t.Fatalf("ForgetMessages: %s", err)
	}
	if db.MessageKnown(learnMessageID) {
		t.Error("Expired message found")
	}
}
