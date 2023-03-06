package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Attempts int = iota
	Retry
)

// contains data about a server
type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

// SetAlive for this backend
func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

// Returns true when backend is alive
func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return
}

// storing all of the backend instances using a slice
type ServerPool struct {
	Backends []*Backend
	current  uint64
}

// store a backend instance inside the serverpool
func(s *ServerPool) addBackend(b *Backend){
	s.Backends = append(s.Backends, b);

}

// atomically increases the counter and returns an index
func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.Backends)))
}

//  changes the status of a backend, if down
func(s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool){
	for _, i:= range s.Backends {
		if i.URL.String() == backendUrl.String(){
			i.SetAlive(alive)
			break
		}
	}
}

// checks whether a backend is alive by establishing a TCP connection
func isBackendAlive(u *url.URL) bool{
	timeout:= 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil{
		log.Println("Site unreachable")
		return false
	}
	defer conn.Close()
	return true
}


func (s *ServerPool) GetNextPeer() *Backend {
	next := s.NextIndex()
	length := len(s.Backends) + next
	for i := next; i < length; i++ {
		idx := i % len(s.Backends)
		if s.Backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.Backends[idx]
		}
	}
	return nil
}


func loadbalancer(w http.ResponseWriter, r *http.Request, serverPool *ServerPool){
	peer := serverPool.GetNextPeer()
	if peer != nil{
		peer.ReverseProxy.ServeHTTP(w,r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
	
}

func main() {
	u, _ := url.Parse("http://localhost:8080")
	rp := httputil.NewSingleHostReverseProxy(u)

	handler := http.HandlerFunc(rp.ServeHTTP)

	if err := http.ListenAndServe(":8081", handler); err != nil {
		panic(err)
	}

}
