package listparse

import (
	"bytes"
	"testing"

	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils/repproto/structs"
)

func TestWriteMessageList(t *testing.T) {
	var testData []*structs.MessageStruct
	testData = append(testData, &structs.MessageStruct{
		PostTime:      1,
		ExpireTime:    3012,
		MessageID:     [message.MessageIDSize]byte{0x00, 0x01, 0x01, 0x03},
		SignerPub:     [message.SignerPubKeySize]byte{0x00, 0x02, 0x02, 0x03},
		Distance:      3,
		ExpireRequest: 100,
		OneTime:       true,
		Sync:          false,
		Hidden:        false,
		Counter:       200,
	})
	testData = append(testData, &structs.MessageStruct{
		PostTime:      2,
		ExpireTime:    3312,
		MessageID:     [message.MessageIDSize]byte{0x00, 0x01, 0x01, 0x03},
		SignerPub:     [message.SignerPubKeySize]byte{0x00, 0x02, 0x02, 0x03},
		Distance:      3,
		ExpireRequest: 100,
		OneTime:       true,
		Sync:          false,
		Hidden:        false,
		Counter:       200,
	})

	buf := new(bytes.Buffer)
	WriteMessageList(testData, buf)
	entries, lastline, err := ReadMessageList(buf.Bytes())
	if err != nil {
		t.Errorf("Parsing failed: %s", err)
	}
	if string(lastline) != "IDX: 200 2 3312 100 1tVtzky3tqDFrq7u7XXbappG4gXdi6831ufD4smVPu 11111111111111111111111111111111 12mzf6NRkKJFting3NgV3oos1t8qHX9KrPKcgo1xkLF 3 true false false" {
		t.Errorf("Bad last line: %s", string(lastline))
	}
	if len(testData) != len(entries) {
		t.Fatal("Skipped lines")
	}
	for i, s := range testData {
		if s.PostTime != entries[i].PostTime {
			t.Errorf("Parse failed: PostTime: %d != %d", s.PostTime, entries[i].PostTime)
		}
		if s.ExpireTime != entries[i].ExpireTime {
			t.Errorf("Parse failed: ExpireTime: %d != %d", s.ExpireTime, entries[i].ExpireTime)
		}
		if s.MessageID != entries[i].MessageID {
			t.Errorf("Parse failed: MessageID: %x != %x", s.MessageID, entries[i].MessageID)
		}
		if s.SignerPub != entries[i].SignerPub {
			t.Errorf("Parse failed: SignerPub: %x != %x", s.SignerPub, entries[i].SignerPub)
		}
		if s.Distance != entries[i].Distance {
			t.Errorf("Parse failed: Distance: %d != %d", s.Distance, entries[i].Distance)
		}
		if s.OneTime != entries[i].OneTime {
			t.Errorf("Parse failed: OneTime: %t != %t", s.OneTime, entries[i].OneTime)
		}
		if s.Sync != entries[i].Sync {
			t.Errorf("Parse failed: Sync: %t != %t", s.Sync, entries[i].Sync)
		}
	}
}
