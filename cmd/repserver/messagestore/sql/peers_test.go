package sql

import (
	"os"
	"path"
	"strconv"
	"testing"
)

var testPeerPubKey = sliceToEDPublicKey(
	[]byte(
		strconv.Itoa(
			int(
				CurrentTime(),
			),
		) + "Peer",
	),
)
var testSignedToken = sliceToProofTokenSigned(
	[]byte(
		strconv.Itoa(
			int(
				CurrentTime(),
			),
		) + "Token",
	),
)

func TestPeersMysql(t *testing.T) {
	if !testing.Short() {
		dir := path.Join(os.TempDir(), "repbinmsg")
		db, err := newMySQLForTest(dir, 100)
		if err != nil {
			t.Fatalf("New Mysql: %s", err)
		}
		defer db.Close()
		// ===== PEERS TESTS =====
		err = db.InsertPeer(testPeerPubKey)
		if err != nil {
			t.Fatalf("InsertPeer: %s", err)
		}
		err = db.TouchPeer(testPeerPubKey)
		if err != nil {
			t.Fatalf("TouchPeer: %s", err)
		}
		err = db.UpdatePeerStats(testPeerPubKey, 1, 2, 3)
		if err != nil {
			t.Fatalf("UpdatePeerStats: %s", err)
		}
		err = db.UpdatePeerNotification(testPeerPubKey, true)
		if err != nil {
			t.Fatalf("UpdatePeerNotification: %s", err)
		}
		err = db.UpdatePeerToken(testPeerPubKey, testSignedToken)
		if err != nil {
			t.Fatalf("UpdatePeerToken: %s", err)
		}
		peerData, err := db.SelectPeer(testPeerPubKey)
		if err != nil {
			t.Fatalf("SelectPeer: %s", err)
		}
		if peerData.AuthToken != *testSignedToken {
			t.Error("Token no match")
		}
		if peerData.ErrorCount != 4 {
			t.Errorf("Errorcount bad: %d != %d", 4, peerData.ErrorCount)
		}
		if peerData.LastFetch != 1 {
			t.Errorf("LastFetch bad: %d != %d", 1, peerData.LastFetch)
		}
		if peerData.LastPosition != 2 {
			t.Errorf("LastPosition bad: %d != %d", 2, peerData.LastPosition)
		}
		now := uint64(CurrentTime())
		if now != peerData.LastNotifyFrom && now+1 != peerData.LastNotifyFrom {
			t.Errorf("LastNotifyFrom: %d != %d", now, peerData.LastNotifyFrom)
		}
		if now != peerData.LastNotifySend && now+1 != peerData.LastNotifySend {
			t.Errorf("LastNotifySend: %d != %d", now, peerData.LastNotifySend)
		}
	}
}

func TestPeersSQLite(t *testing.T) {
	dir := path.Join(os.TempDir(), "repbinmsg")
	dbFile := path.Join(os.TempDir(), "db.test-peers")
	db, err := New("sqlite3", dbFile, dir, 100)
	if err != nil {
		t.Fatalf("New sqlite3: %s", err)
	}
	defer os.Remove(dbFile)
	defer db.Close()
	// ===== PEERS TESTS =====
	err = db.InsertPeer(testPeerPubKey)
	if err != nil {
		t.Fatalf("InsertPeer: %s", err)
	}
	err = db.TouchPeer(testPeerPubKey)
	if err != nil {
		t.Fatalf("TouchPeer: %s", err)
	}
	err = db.UpdatePeerStats(testPeerPubKey, 1, 2, 3)
	if err != nil {
		t.Fatalf("UpdatePeerStats: %s", err)
	}
	err = db.UpdatePeerNotification(testPeerPubKey, true)
	if err != nil {
		t.Fatalf("UpdatePeerNotification: %s", err)
	}
	err = db.UpdatePeerToken(testPeerPubKey, testSignedToken)
	if err != nil {
		t.Fatalf("UpdatePeerToken: %s", err)
	}
	peerData, err := db.SelectPeer(testPeerPubKey)
	if err != nil {
		t.Fatalf("SelectPeer: %s", err)
	}
	if peerData.AuthToken != *testSignedToken {
		t.Error("Token no match")
	}
	if peerData.ErrorCount != 4 {
		t.Errorf("Errorcount bad: %d != %d", 4, peerData.ErrorCount)
	}
	if peerData.LastFetch != 1 {
		t.Errorf("LastFetch bad: %d != %d", 1, peerData.LastFetch)
	}
	if peerData.LastPosition != 2 {
		t.Errorf("LastPosition bad: %d != %d", 2, peerData.LastPosition)
	}
	now := uint64(CurrentTime())
	if now != peerData.LastNotifyFrom && now+1 != peerData.LastNotifyFrom {
		t.Errorf("LastNotifyFrom: %d != %d", now, peerData.LastNotifyFrom)
	}
	if now != peerData.LastNotifySend && now+1 != peerData.LastNotifySend {
		t.Errorf("LastNotifySend: %d != %d", now, peerData.LastNotifySend)
	}
}
