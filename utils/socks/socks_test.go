package socks

import (
	"github.com/repbin/repbin/utils"
	"fmt"
	"testing"
)

func TestProxy_Get(t *testing.T) {
	url := "http://google.com/"
	socksProxy := Proxy("socks5://127.0.0.1:9050")
	resp, err := socksProxy.Get(url)
	if err != nil {
		t.Fatalf("Could not call: %s", err)
	}
	body, err := utils.MaxRead(250000, resp.Body)
	fmt.Println(string(body))
}
