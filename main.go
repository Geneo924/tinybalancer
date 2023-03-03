package main

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	URL				*url.URL
	Alive			bool
	mux				sync.RWMutex
	ReverseProxy	*httputil.ReverseProxy
}

type ServerPool struct {
	Backends 	 []*Backend
	current uint64
}

