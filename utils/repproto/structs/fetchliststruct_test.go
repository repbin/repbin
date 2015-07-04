package structs

import (
	"testing"

	"github.com/repbin/repbin/message"
)

func TestFetchListStructEncode(t *testing.T) {
	td := FetchListStruct{
		MessageID:   [message.MessageIDSize]byte{0xff, 0x00, 0x00, 0x00, 0x01},
		TimeEntered: 20910390213,
	}
	enc := td.Encode().Fill()
	if len(enc) != FetchListStructSize {
		t.Error("Fill failed")
	}
	dec := FetchListStructDecode(enc)
	if td.MessageID != dec.MessageID {
		t.Errorf("Decode MessageID failed: %x", dec.MessageID)
	}
	if td.TimeEntered != dec.TimeEntered {
		t.Errorf("Decode TimeEntered failed: %d", dec.TimeEntered)
	}
}
