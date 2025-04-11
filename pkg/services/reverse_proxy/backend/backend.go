package backend

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/letronghoangminh/reproxy/pkg/interfaces"
)

type backend struct {
	url          *url.URL
	alive        bool
	mux          sync.RWMutex
	connections  int
	reverseProxy *httputil.ReverseProxy
	cookies      []*http.Cookie
}

func (b *backend) GetActiveConnections() int {
	b.mux.RLock()
	connections := b.connections
	b.mux.RUnlock()
	return connections
}

func (b *backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.alive = alive
	b.mux.Unlock()
}

func (b *backend) IsAlive() bool {
	b.mux.RLock()
	alive := b.alive
	defer b.mux.RUnlock()
	return alive
}

func (b *backend) GetURL() *url.URL {
	return b.url
}

func (b *backend) Serve(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		b.mux.Lock()
		b.connections--
		b.mux.Unlock()
	}()

	b.mux.Lock()
	b.connections++
	b.mux.Unlock()

	for _, cookie := range b.cookies {
		http.SetCookie(rw, cookie)
	}

	b.reverseProxy.ServeHTTP(rw, req)
}

func (b *backend) AddCookie(cookie *http.Cookie) {
	b.mux.Lock()
	b.cookies = append(b.cookies, cookie)
	b.mux.Unlock()
}

func NewBackend(u *url.URL, rp *httputil.ReverseProxy) interfaces.Backend {
	return &backend{
		url:          u,
		alive:        true,
		reverseProxy: rp,
	}
}
