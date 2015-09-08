package sql

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path"
	"sync"

	"github.com/agl/ed25519"
	"github.com/repbin/repbin/hashcash"
	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils/keyproof"
)

// Version of this release
const Version = "0.0.1 very alpha"

var shardRand = make([]byte, 16)

func init() {
	io.ReadFull(rand.Reader, shardRand)
}

var (
	// ErrNoModify is returned if a row was not modified
	ErrNoModify = errors.New("storage: Row not modified")
)

// MessageDB implements a message database
type MessageDB struct {
	db                     *sql.DB
	NumShards              uint64
	shardMutexes           []*sync.Mutex
	dir                    string
	driver                 string
	queries                map[string]string
	signerInsertQ          *sql.Stmt
	signerSelectPublicKeyQ *sql.Stmt
	signerSelectIDQ        *sql.Stmt
	signerUpdateQ          *sql.Stmt
	signerUpdateInsertQ    *sql.Stmt
	signerAddMessageQ      *sql.Stmt
	signerDelMessageQ      *sql.Stmt
	signerPrepareExpireQ   *sql.Stmt
	signerExpireQ          *sql.Stmt
	signerSetExpireQ       *sql.Stmt
	peerInsertQ            *sql.Stmt
	peerUpdateStatQ        *sql.Stmt
	peerUpdateTokenQ       *sql.Stmt
	peerUpdateNotifyQ      *sql.Stmt
	peerSelectQ            *sql.Stmt
	nextMessageCounterQ    *sql.Stmt
	incrMessageCounterQ    *sql.Stmt
	expireMessageCounterQ  *sql.Stmt
	insertMessageCounterQ  *sql.Stmt
	insertMessageQ         *sql.Stmt
	selectMessageQ         *sql.Stmt
	deleteMessageQ         *sql.Stmt
	updateExpireMessageQ   *sql.Stmt
	selectExpireMessageQ   *sql.Stmt
	globalIndexAddQ        *sql.Stmt
	getKeyIndexQ           *sql.Stmt
	getGlobalIndexQ        *sql.Stmt
	messageBlobInsertQ     *sql.Stmt
	messageBlobSelectQ     *sql.Stmt
	messageBlobDeleteQ     *sql.Stmt
}

// New returns a new message database. driver is the database driver to use,
// url the database url. dir is the optional directory in which to store the
// raw message blobs. If dir is empty blobs will be stored in the database (which
// may not be a good idea at all). Shards is the number of lock shards to use
// for sequence generation (memory/lock-probability tradeoff)
func New(driver, url, dir string, shards int) (*MessageDB, error) {
	var db *sql.DB
	var err error
	if driver == "sqlite3" {
		if err := os.MkdirAll(path.Dir(url), 0700); err != nil {
			return nil, err
		}
	}
	db, err = sql.Open(driver, url)
	if err != nil {
		return nil, err
	}
	mdb := &MessageDB{
		NumShards: uint64(shards),
		queries:   queries[driver],
		db:        db,
		dir:       dir,
		driver:    driver,
	}

	if _, err := mdb.db.Exec(mdb.queries["SignerCreate"]); err != nil {
		return nil, err
	}
	if _, err := mdb.db.Exec(mdb.queries["PeerCreate"]); err != nil {
		return nil, err
	}
	if _, err := mdb.db.Exec(mdb.queries["MessageCreate"]); err != nil {
		return nil, err
	}
	if _, err := mdb.db.Exec(mdb.queries["GlobalIndexCreate"]); err != nil {
		return nil, err
	}
	if _, err := mdb.db.Exec(mdb.queries["messageBlobCreate"]); err != nil {
		return nil, err
	}
	if _, err := mdb.db.Exec(mdb.queries["MessageCounterCreate"]); err != nil {
		return nil, err
	}
	if mdb.signerInsertQ, err = mdb.db.Prepare(mdb.queries["SignerInsert"]); err != nil {
		return nil, err
	}
	if mdb.signerSelectPublicKeyQ, err = mdb.db.Prepare(mdb.queries["SelectSignerPublicKey"]); err != nil {
		return nil, err
	}
	if mdb.signerSelectIDQ, err = mdb.db.Prepare(mdb.queries["SelectSignerID"]); err != nil {
		return nil, err
	}
	if mdb.signerUpdateQ, err = mdb.db.Prepare(mdb.queries["UpdateSigner"]); err != nil {
		return nil, err
	}
	if mdb.signerAddMessageQ, err = mdb.db.Prepare(mdb.queries["AddMessageSigner"]); err != nil {
		return nil, err
	}
	if mdb.signerDelMessageQ, err = mdb.db.Prepare(mdb.queries["DelMessageSigner"]); err != nil {
		return nil, err
	}
	if mdb.signerPrepareExpireQ, err = mdb.db.Prepare(mdb.queries["PrepareExpireSigner"]); err != nil {
		return nil, err
	}
	if mdb.signerExpireQ, err = mdb.db.Prepare(mdb.queries["DeleteExpireSigner"]); err != nil {
		return nil, err
	}

	if mdb.peerInsertQ, err = mdb.db.Prepare(mdb.queries["InsertPeer"]); err != nil {
		return nil, err
	}
	if mdb.peerUpdateStatQ, err = mdb.db.Prepare(mdb.queries["UpdateStatPeer"]); err != nil {
		return nil, err
	}
	if mdb.peerUpdateNotifyQ, err = mdb.db.Prepare(mdb.queries["UpdateNotifyPeer"]); err != nil {
		return nil, err
	}
	if mdb.peerUpdateTokenQ, err = mdb.db.Prepare(mdb.queries["UpdateTokenPeer"]); err != nil {
		return nil, err
	}
	if mdb.peerSelectQ, err = mdb.db.Prepare(mdb.queries["SelectPeer"]); err != nil {
		return nil, err
	}

	if mdb.nextMessageCounterQ, err = mdb.db.Prepare(mdb.queries["NextMessageCounter"]); err != nil {
		return nil, err
	}
	if mdb.incrMessageCounterQ, err = mdb.db.Prepare(mdb.queries["IncreaseMessageCounter"]); err != nil {
		return nil, err
	}
	if mdb.insertMessageCounterQ, err = mdb.db.Prepare(mdb.queries["InsertMessageCounter"]); err != nil {
		return nil, err
	}
	if mdb.expireMessageCounterQ, err = mdb.db.Prepare(mdb.queries["ExpireMessageCounter"]); err != nil {
		return nil, err
	}

	if mdb.insertMessageQ, err = mdb.db.Prepare(mdb.queries["InsertMessage"]); err != nil {
		return nil, err
	}
	if mdb.selectMessageQ, err = mdb.db.Prepare(mdb.queries["SelectMessage"]); err != nil {
		return nil, err
	}
	if mdb.deleteMessageQ, err = mdb.db.Prepare(mdb.queries["DeleteMessage"]); err != nil {
		return nil, err
	}
	if mdb.updateExpireMessageQ, err = mdb.db.Prepare(mdb.queries["UpdateExpireMessage"]); err != nil {
		return nil, err
	}
	if mdb.selectExpireMessageQ, err = mdb.db.Prepare(mdb.queries["SelectExpireMessage"]); err != nil {
		return nil, err
	}

	if mdb.globalIndexAddQ, err = mdb.db.Prepare(mdb.queries["globalIndexAdd"]); err != nil {
		return nil, err
	}
	if mdb.getKeyIndexQ, err = mdb.db.Prepare(mdb.queries["getKeyIndex"]); err != nil {
		return nil, err
	}
	if mdb.getGlobalIndexQ, err = mdb.db.Prepare(mdb.queries["getGlobalIndex"]); err != nil {
		return nil, err
	}

	if mdb.messageBlobInsertQ, err = mdb.db.Prepare(mdb.queries["messageBlobInsert"]); err != nil {
		return nil, err
	}
	if mdb.messageBlobSelectQ, err = mdb.db.Prepare(mdb.queries["messageBlobSelect"]); err != nil {
		return nil, err
	}
	if mdb.messageBlobDeleteQ, err = mdb.db.Prepare(mdb.queries["messageBlobDelete"]); err != nil {
		return nil, err
	}

	if driver == "mysql" {
		if mdb.signerUpdateInsertQ, err = mdb.db.Prepare(mdb.queries["UpdateOrInsertSigner"]); err != nil {
			return nil, err
		}
	}
	if dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	mdb.setMutexes()
	return mdb, nil
}

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

// LockShard locks shard s
func (db *MessageDB) LockShard(s []byte) {
	db.shardMutexes[db.calcShard(s)].Lock()
}

// UnlockShard unlocks shard s, if locked. Runtime error otherwise
func (db *MessageDB) UnlockShard(s []byte) {
	db.shardMutexes[db.calcShard(s)].Unlock()
}

func (db *MessageDB) setMutexes() {
	db.shardMutexes = make([]*sync.Mutex, db.NumShards)
	for i := range db.shardMutexes {
		db.shardMutexes[i] = new(sync.Mutex)
	}
}

// Close the database
func (db *MessageDB) Close() error {
	return db.db.Close()
}

func toHex(d []byte) string {
	return hex.EncodeToString(d)
}

func fromHex(s string) []byte {
	d, _ := hex.DecodeString(s)
	return d
}

func sliceToSignerPubKey(d []byte) *[message.SignerPubKeySize]byte {
	r := new([message.SignerPubKeySize]byte)
	copy(r[:], d)
	return r
}

func sliceToNonce(d []byte) *[hashcash.NonceSize]byte {
	r := new([hashcash.NonceSize]byte)
	copy(r[:], d)
	return r
}

func sliceToProofTokenSigned(d []byte) *[keyproof.ProofTokenSignedSize]byte {
	r := new([keyproof.ProofTokenSignedSize]byte)
	copy(r[:], d)
	return r
}

func sliceToEDPublicKey(d []byte) *[ed25519.PublicKeySize]byte {
	r := new([ed25519.PublicKeySize]byte)
	copy(r[:], d)
	return r
}

func sliceToCurve25519Key(d []byte) *message.Curve25519Key {
	r := new(message.Curve25519Key)
	copy(r[:], d)
	return r
}

func sliceToMessageID(d []byte) *[message.MessageIDSize]byte {
	r := new([message.MessageIDSize]byte)
	copy(r[:], d)
	return r
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	if i > 0 {
		return true
	}
	return false
}

func updateConvertNilError(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n < 1 {
		return ErrNoModify
	}
	return nil
}

func (db *MessageDB) calcShard(d []byte) uint64 {
	h := sha256.Sum256(append(shardRand, d...))
	return binary.BigEndian.Uint64(h[:]) % db.NumShards
}
