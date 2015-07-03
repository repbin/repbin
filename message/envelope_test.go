package message

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestEncodeBase64(t *testing.T) {
	msg := []byte("Some message to encode")
	msgenc := EncodeBase64(msg)
	msgdec, err := base64.StdEncoding.DecodeString(string(msgenc))
	if err != nil {
		t.Errorf("Decode failure: %s", err)
	}
	if !bytes.Equal(msg, msgdec) {
		t.Error("Encode failed")
	}
	msgdec2, err := msgenc.Decode()
	if err != nil {
		t.Errorf("Decode2 failure: %s", err)
	}
	if !bytes.Equal(msg, msgdec2) {
		t.Error("Decode failed")
	}
	msgenc.GetSignHeader()
}
