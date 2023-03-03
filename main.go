package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

type ServerPool struct {
	Backends []*Backend
	current  uint64
}


func main() {
	u, _ := url.Parse("http://localhost:8080")
	rp := httputil.NewSingleHostReverseProxy(u)

	handler := http.HandlerFunc(rp.ServeHTTP)

	if err := http.ListenAndServe(":8081",handler); err != nil {
		panic(err)
	}

}
