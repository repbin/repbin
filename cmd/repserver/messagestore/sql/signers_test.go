package sql

import (
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/repbin/repbin/hashcash"               // must
	"github.com/repbin/repbin/utils/repproto/structs" // must
)

var testSigner = &structs.SignerStruct{
	PublicKey: *sliceToSignerPubKey(
		[]byte(
			strconv.Itoa(
				int(
					time.Now().Unix(),
				),
			),
		),
	),
	Nonce:               [hashcash.NonceSize]byte{0xff, 0xac, 0x00, 0x01},
	Bits:                8,
	MaxMessagesPosted:   1,
	MaxMessagesRetained: 19,
	ExpireTarget:        291928122,
}

func insertSignerQuery(t *testing.T, db *MessageDB, data *structs.SignerStruct) int64 {
	id, err := db.InsertSigner(data)
	if err != nil {
		t.Fatalf("InsertSigner: %s", err)
	}
	if id < 1 {
		t.Fatalf("InsertSigner InsertID: %d<1", id)
	}
	return id
}

func selectSignerQuery(t *testing.T, db *MessageDB, data *structs.SignerStruct) (int64, *structs.SignerStruct) {
	id, str, err := db.SelectSigner(&data.PublicKey)
	if err != nil {
		t.Fatalf("SelectSigner: %s", err)
	}
	if id < 1 {
		t.Fatalf("SelectSigner InsertID: %d<1", id)
	}
	return id, str
}

func selectSignerIDQuery(t *testing.T, db *MessageDB, tid int64) (int64, *structs.SignerStruct) {
	id, str, err := db.SelectSignerByID(tid)
	if err != nil {
		t.Fatalf("SelectSignerByID: %s", err)
	}
	if id < 1 {
		t.Fatalf("SelectSignerByID InsertID: %d<1", id)
	}
	return id, str
}

func updateSigner(t *testing.T, db *MessageDB, data *structs.SignerStruct) {
	err := db.UpdateSigner(data)
	if err != nil {
		t.Fatalf("UpdateSigner: %s", err)
	}
}

func insertOrUpdateSigner(t *testing.T, db *MessageDB, data *structs.SignerStruct) {
	err := db.InsertOrUpdateSigner(data)
	if err != nil {
		t.Fatalf("InsertOrUpdateSigner: %s", err)
	}
}

func signerCompare(t *testing.T, id1, id2 int64, baseData, data *structs.SignerStruct) {
	if id1 != id2 {
		t.Fatalf("signerCompare ID No match: %d != %d", id1, id2)
	}
	if data.PublicKey != baseData.PublicKey {
		t.Fatal("signerCompare PublicKey")
	}
	if data.Nonce != baseData.Nonce {
		t.Fatal("signerCompare Nonce")
	}
	if data.Bits != baseData.Bits {
		t.Fatal("signerCompare Bits")
	}
	if data.MaxMessagesPosted != baseData.MaxMessagesPosted {
		t.Fatal("signerCompare MaxMessagesPosted")
	}
	if data.MaxMessagesRetained != baseData.MaxMessagesRetained {
		t.Fatalf("signerCompare MaxMessagesRetained: %d!=%d", data.MaxMessagesRetained, baseData.MaxMessagesRetained)
	}
	if data.ExpireTarget != baseData.ExpireTarget {
		t.Fatal("signerCompare ExpireTarget")
	}
}

func TestSignerMysql(t *testing.T) {
	dir := path.Join(os.TempDir(), "repbinmsg")
	db, err := New("mysql", "root:root@/repbin", dir)
	if err != nil {
		t.Fatalf("New Mysql: %s", err)
	}
	defer db.Close()
	// ===== SIGNER TESTS =====
	sigID := insertSignerQuery(t, db, testSigner)
	sigID2, sigData2 := selectSignerQuery(t, db, testSigner)
	signerCompare(t, sigID, sigID2, testSigner, sigData2)
	sigID3, sigData3 := selectSignerIDQuery(t, db, sigID)
	signerCompare(t, sigID, sigID3, testSigner, sigData3)
	sigData3.MaxMessagesRetained = 13
	updateSigner(t, db, sigData3)
	sigID4, sigData4 := selectSignerIDQuery(t, db, sigID)
	signerCompare(t, sigID, sigID4, sigData3, sigData4)
	sigData3.MaxMessagesRetained = 10
	insertOrUpdateSigner(t, db, sigData3)
	sigID5, sigData5 := selectSignerIDQuery(t, db, sigID)
	signerCompare(t, sigID, sigID5, sigData3, sigData5)
	sigData3.PublicKey = *sliceToSignerPubKey(
		[]byte(
			strconv.Itoa(
				int(
					time.Now().Unix(),
				),
			) + "-new",
		),
	)
	insertOrUpdateSigner(t, db, sigData3)
	_, sigData5 = selectSignerQuery(t, db, sigData3)
	signerCompare(t, 0, 0, sigData3, sigData5)
	err = db.AddMessage(&sigData5.PublicKey)
	if err != nil {
		t.Errorf("AddMessage: %s", err)
	}
	_, sigData6, _ := db.SelectSigner(&sigData5.PublicKey)
	if sigData5.MessagesRetained != sigData6.MessagesRetained-1 {
		t.Error("AddMessage MessagesRetained")
	}
	if sigData5.MessagesPosted != sigData6.MessagesPosted-1 {
		t.Error("AddMessage MessagesPosted")
	}
	err = db.DelMessage(&sigData5.PublicKey)
	if err != nil {
		t.Errorf("DelMessage: %s", err)
	}
	_, sigData6, _ = db.SelectSigner(&sigData5.PublicKey)
	if sigData5.MessagesRetained != sigData6.MessagesRetained {
		t.Error("DelMessage MessagesRetained")
	}
	if sigData5.MessagesPosted != sigData6.MessagesPosted-1 {
		t.Error("DelMessage MessagesPosted")
	}
	_, _, err = db.ExpireSigners(10)
	if err != nil {
		t.Errorf("ExpireSigners: %s", err)
	}
	// ===== SIGNER TESTS =====
}

func TestSignersSQLite(t *testing.T) {
	dir := path.Join(os.TempDir(), "repbinmsg")
	dbFile := path.Join(os.TempDir(), "db.test-signers")
	db, err := New("sqlite3", dbFile, dir)
	if err != nil {
		t.Fatalf("New sqlite3: %s", err)
	}
	defer os.Remove(dbFile)
	defer db.Close()
	// ===== SIGNER TESTS =====
	sigID := insertSignerQuery(t, db, testSigner)
	sigID2, sigData2 := selectSignerQuery(t, db, testSigner)
	signerCompare(t, sigID, sigID2, testSigner, sigData2)
	sigID3, sigData3 := selectSignerIDQuery(t, db, sigID)
	signerCompare(t, sigID, sigID3, testSigner, sigData3)
	sigData3.MaxMessagesRetained = 13
	updateSigner(t, db, sigData3)
	sigID4, sigData4 := selectSignerIDQuery(t, db, sigID)
	signerCompare(t, sigID, sigID4, sigData3, sigData4)
	sigData3.MaxMessagesRetained = 10
	insertOrUpdateSigner(t, db, sigData3)
	sigID5, sigData5 := selectSignerIDQuery(t, db, sigID)
	signerCompare(t, sigID, sigID5, sigData3, sigData5)
	sigData3.PublicKey = *sliceToSignerPubKey(
		[]byte(
			strconv.Itoa(
				int(
					time.Now().Unix(),
				),
			) + "-new",
		),
	)
	insertOrUpdateSigner(t, db, sigData3)
	_, sigData5 = selectSignerQuery(t, db, sigData3)
	signerCompare(t, 0, 0, sigData3, sigData5)
	err = db.AddMessage(&sigData5.PublicKey)
	if err != nil {
		t.Errorf("AddMessage: %s", err)
	}
	_, sigData6, _ := db.SelectSigner(&sigData5.PublicKey)
	if sigData5.MessagesRetained != sigData6.MessagesRetained-1 {
		t.Error("AddMessage MessagesRetained")
	}
	if sigData5.MessagesPosted != sigData6.MessagesPosted-1 {
		t.Error("AddMessage MessagesPosted")
	}
	err = db.DelMessage(&sigData5.PublicKey)
	if err != nil {
		t.Errorf("DelMessage: %s", err)
	}
	_, sigData6, _ = db.SelectSigner(&sigData5.PublicKey)
	if sigData5.MessagesRetained != sigData6.MessagesRetained {
		t.Error("DelMessage MessagesRetained")
	}
	if sigData5.MessagesPosted != sigData6.MessagesPosted-1 {
		t.Error("DelMessage MessagesPosted")
	}
	_, _, err = db.ExpireSigners(10)
	if err != nil {
		t.Errorf("ExpireSigners: %s", err)
	}
	// ===== SIGNER TESTS =====
}
