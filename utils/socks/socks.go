package socks

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	urlPackage "net/url"
	"time"

	"github.com/repbin/repbin/utils"
	"golang.org/x/net/proxy"
)

// AcceptNoSocks allows workaround if no socks proxy is given
var AcceptNoSocks = false

// ErrNoProxy is returned if no proxy has been configured and AcceptNoSocks == false
var ErrNoProxy = errors.New("socks: No proxy configured")

// Timeout for socks connections. In seconds
var Timeout = int64(90)

// Proxy is a socks proxy
type Proxy string

// LimitGet fetches at most limit bytes from url. Throws error if more data is available
func (sprox Proxy) LimitGet(url string, limit int64) ([]byte, error) {
	resp, err := sprox.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := utils.MaxRead(limit, resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

// PostBytes makes a post of []byte
func (sprox Proxy) PostBytes(url string, bodyType string, body []byte) (*http.Response, error) {
	return sprox.Post(url, bodyType, bytes.NewBuffer(body))
}

// LimitPostBytes posts []byte and limits the return to limit. Throws error if more data is available
func (sprox Proxy) LimitPostBytes(url string, bodyType string, body []byte, limit int64) ([]byte, error) {
	resp, err := sprox.PostBytes(url, bodyType, body)
	if err != nil {
		return nil, err
	}
	retbody, err := utils.MaxRead(limit, resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return retbody, nil
}

// LimitPost posts body and reads a maximum of limit. Throws error if more data is available
func (sprox Proxy) LimitPost(url string, bodyType string, body io.Reader, limit int64) ([]byte, error) {
	resp, err := sprox.Post(url, bodyType, body)
	if err != nil {
		return nil, err
	}
	retbody, err := utils.MaxRead(limit, resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return retbody, nil
}

// Get execute a get call
func (sprox Proxy) Get(url string) (*http.Response, error) {
	client, err := sprox.Create()
	if err != nil {
		return nil, err
	}
	return client.Get(url)
}

// PostForm posts a form. data is url.Values
func (sprox Proxy) PostForm(url string, data urlPackage.Values) (resp *http.Response, err error) {
	client, err := sprox.Create()
	if err != nil {
		return nil, err
	}
	return client.PostForm(url, data)
}

// Head call
func (sprox Proxy) Head(url string) (resp *http.Response, err error) {
	client, err := sprox.Create()
	if err != nil {
		return nil, err
	}
	return client.Head(url)
}

// Do a request
func (sprox Proxy) Do(req *http.Request) (resp *http.Response, err error) {
	client, err := sprox.Create()
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

// Post execute a post call
func (sprox Proxy) Post(url string, bodyType string, body io.Reader) (*http.Response, error) {
	client, err := sprox.Create()
	if err != nil {
		return nil, err
	}
	return client.Post(url, bodyType, body)
}

// Create a new socks proxy
func (sprox Proxy) Create() (*http.Client, error) {
	if AcceptNoSocks && string(sprox) == "" {
		return &http.Client{}, nil
	} else if string(sprox) == "" {
		return nil, ErrNoProxy
	}
	socksURL, err := urlPackage.Parse(string(sprox))
	if err != nil {
		return nil, err
	}
	if socksURL.Scheme == "" {
		socksURL, err = urlPackage.Parse("socks5://" + string(sprox))
		if err != nil {
			return nil, err
		}
	}
	dialer, err := proxy.FromURL(socksURL, proxy.Direct)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		//Proxy: http.ProxyFromEnvironment,
		Dial: dialer.Dial,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * time.Duration(Timeout), // 45 second timeout is pretty nice
	}
	return client, nil
}
