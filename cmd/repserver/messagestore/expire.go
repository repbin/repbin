package messagestore

import (
	"time"

	log "github.com/repbin/repbin/deferconsole"
	"github.com/repbin/repbin/fileback"
	"github.com/repbin/repbin/utils"
	"github.com/repbin/repbin/utils/repproto/structs"
)

// Expire....
// 1.) Handle expire index
// 2.) Delete from keyindex, signers what hasnt been touched within the last X seconds

// ExpireFromFS expires data based on filesystem last change
func (store Store) ExpireFromFS() {
	store.expireSigners()
	store.expireKeyIndices()
}

func (store Store) expireSigners() {
	//s.signers = fileback.NewRoundRobin(dir+signerDir, structs.SignerStructSize, 10, ' ', []byte("\n"), workers)       // We keep signer history, last 10 entries
	now := time.Now().Unix()
	indices := store.signers.Indices()
	if indices == nil {
		log.Debugs("ExpireFS: No signers\n")
		return
	}
	for _, signer := range indices {
		last := store.signers.Index(signer).LastChange(0)
		if last != 0 && last < now-int64(FSExpire) {
			store.signers.Index(signer).Truncate() // We don't delete for real, otherwise the signer would become reusable.
			log.Debugf("ExpireFS: Signer truncated %x\n", signer)
		} else {
			log.Debugf("ExpireFS: Signer remains %x\n", signer)
		}
	}
}

func (store Store) expireKeyIndices() {
	//s.keyindex = fileback.NewRolling(dir+keyindexDir, structs.MessageStructSize, 2048, ' ', []byte("\n"), workers)    // Maximum 2048 entries per file
	now := time.Now().Unix()
	indices := store.keyindex.Indices()
	if indices == nil {
		log.Debugs("ExpireFS: No Key Indices\n")
		return
	}
	for _, keylist := range indices {
		entry := store.signers.Index(keylist)
		last := entry.LastChange(0)
		if last != 0 && last < now-int64(FSExpire) {
			// go forward
			lastEntry := entry.Entries() - 1
			if lastEntry >= 0 {
				last := entry.LastChange(lastEntry)
				if last < now-int64(FSExpire) {
					log.Debugf("ExpireFS: Key Index complete delete %x\n", keylist)
					entry.Delete() // Key lists are deleted
				} else {
					maxEntries := entry.MaxEntries()
					deleteBefore := int64(-1)
				LastLoop:
					for pos := lastEntry; pos > 0; pos = pos - maxEntries { // Search the first entry that has expired
						last := entry.LastChange(pos)
						if last != 0 && last < now-int64(FSExpire) {
							break LastLoop
						}
						deleteBefore = pos - 1 // This will actually skip the last expired object
					}
					if deleteBefore >= 0 {
						log.Debugf("ExpireFS: Key Index delete before %d %x\n", deleteBefore, keylist)
						entry.DeleteBefore(deleteBefore)
					} else {
						log.Debugf("ExpireFS: Key Index no delete match %x\n", keylist)
					}
				}
			}
		} else {
			log.Debugf("ExpireFS: Key Index remains %x\n", keylist)
		}
	}
}

// ExpireFromIndex reads the expire index and expires messages as they are recorded
func (store Store) ExpireFromIndex(cycles int) {
	// run from now backwards N cycles
	now := uint64(time.Now().Unix())
	expireTime := (now / uint64(ExpireRun))
	for i := cycles; i > 0; i-- {
		expireTime -= uint64(ExpireRun)
		log.Debugf("Expire index try: %d\n", expireTime)
		if expireTime > 0 {
			if store.expireindex.Index(utils.EncodeUInt64(expireTime)).Exists() {
				j := int64(0)
				log.Debugf("Expire index found: %d\n", expireTime)
			ExpireEntries:
				for {
					entry, err := store.expireindex.Index(utils.EncodeUInt64(expireTime)).ReadEntry(j)
					if err != nil { // no more entries
						log.Debugf("Expire index empty: %d\n", expireTime)
						break ExpireEntries
					}
					j++
					if entry == nil {
						log.Debugf("Expire index bad entry: %d %d\n", j, expireTime)
						continue // bad entry
					}
					entryStruct := structs.ExpireStructDecode(entry)
					if entryStruct == nil {
						log.Debugf("Expire index bad entry decode: %d %d\n", j, expireTime)
						continue // bad entry
					}
					if entryStruct.ExpireTime > now {
						log.Debugf("Expire index bad entry not expired: %d %d\n", j, expireTime)
						continue // bad entry
					}
					store.messages.Index(entryStruct.MessageID[:]).Delete() // Delete message
					store.signers.Index(entryStruct.SignerPubKey[:]).Update(func(tx fileback.Tx) error {
						data := tx.GetLast()
						if data == nil { // we don't care about errors here
							return nil
						}
						signer := structs.SignerStructDecode(data)
						if signer == nil {
							return nil
						}
						signer.MessagesRetained--
						tx.Append(signer.Encode().Fill())
						return nil
					})
				}
				// Remove expire list
				store.expireindex.Index(utils.EncodeUInt64(expireTime)).Delete()
				log.Debugf("Expire index deleted: %d\n", expireTime)
			} else {
				log.Debugf("Expire index does not exist: %d\n", expireTime)
			}
		}
	}
}
