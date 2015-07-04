package repproto

import (
	"testing"

	"github.com/repbin/repbin/message"
	"github.com/repbin/repbin/utils"
)

func TestNew(t *testing.T) {
	proto := New("socks5://127.0.0.1:9050/", "http://127.0.0.1:8080")
	server, err := proto.selectServer([]byte("testID"))
	if err != nil {
		t.Fatalf("Select: %s", err)
	}
	server, err = proto.selectServer([]byte("testID"))
	if err == nil {
		t.Error("End of server not detected")
	}
	proto = New("socks5://127.0.0.1:9050/", "http://127.0.0.1:8080", "http://127.0.0.2:8080")
	server, err = proto.selectServer([]byte("testID"))
	if err != nil {
		t.Fatalf("Select: %s", err)
	}
	server1, err := proto.selectServer([]byte("testID"))
	if err != nil {
		t.Fatalf("Select: %s", err)
	}
	if server == server1 {
		t.Error("Server permutation failed")
	}
	_, err = proto.selectServer([]byte("testID"))
	if err == nil {
		t.Error("End of servers undetected")
	}

	_ = server
}

func TestConstructURL(t *testing.T) {
	if "http://www.google.com" != constructURL("http://www.google.com") {
		t.Error("Construct error 1")
	}
	if "http://www.google.com/local" != constructURL("http://www.google.com", "local") {
		t.Error("Construct error 2")
	}
	if "http://www.google.com/local" != constructURL("http://www.google.com/", "local") {
		t.Error("Construct error 3")
	}
	if "http://www.google.com/local" != constructURL("http://www.google.com/", "/local") {
		t.Error("Construct error 4")
	}
	if "http://www.google.com/local/" != constructURL("http://www.google.com/", "/local/") {
		t.Error("Construct error 5")
	}
	if "http://www.google.com/local/?" != constructURL("http://www.google.com/", "/local/?") {
		t.Errorf("Construct error 6: %s", constructURL("http://www.google.com/", "/local/?"))
	}
	if "http://www.google.com/local/?a=" != constructURL("http://www.google.com/", "/local/?a=") {
		t.Error("Construct error 7")
	}
	if "http://www.google.com/local/?a=" != constructURL("http://www.google.com/", "/local/", "?a=") {
		t.Error("Construct error 8")
	}
	if "http://www.google.com/local?a=" != constructURL("http://www.google.com/", "/local", "?a=") {
		t.Error("Construct error 9")
	}
	if "http://www.google.com/local?a=b" != constructURL("http://www.google.com/", "/local", "?a=", "b") {
		t.Error("Construct error 10")
	}
	if "http://www.google.com/local?a=b" != constructURL("http://www.google.com/", "/local", "?a=", "=b") {
		t.Error("Construct error 11")
	}
	if "http://www.google.com/local?a=b&c=d" != constructURL("http://www.google.com/", "/local", "?a=", "=b", "c=", "d") {
		t.Error("Construct error 12")
	}
}

func TestIDSpecific(t *testing.T) {
	var privKey, pubKey message.Curve25519Key
	var pk []byte
	//pk = utils.B58decode("i2BNvH9h85NWDxDZj2VpF8nuvUpSQy57ud3hrzo3gxM")
	pk = utils.B58decode("ATWZibbLy5CFE5atp98hxFRAvFcjN7kNn16WtP43tnJ6")
	pk = utils.B58decode("FpYAGsxrpmgh8CUJkFEnz1CCY9ZUhbVxtTekfkFyWxdQ")
	copy(privKey[:], pk[:])
	pubKey = *message.GenPubKey(&privKey)
	t.Skip() // TODO: enable test with local server
	proto := New("", "http://127.0.0.1:8080")
	info, err := proto.Auth("http://127.0.0.1:8080", privKey[:])
	if err != nil {
		t.Fatalf("ID: %s", err)
	}
	_ = info
	msgs, more, err := proto.ListSpecific("http://127.0.0.1:8080", pubKey[:], privKey[:], 0, 10)
	if err != nil {
		t.Fatalf("ListSpecific: %s", err)
	}
	// fmt.Printf("%+v\n", msgs)
	_, _ = msgs, more
}
