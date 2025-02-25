package main 

import (
  "fmt"
  "net/http"
  "sync"
  "sync/atomic"
)

//structs to hold backends
type Backend struct {
  URL   *url.URL
  Alive   bool
  mux   sync.RWMutex
  ReverseProxy    *httputil.ReverseProxy
}

//struct to track all backends in balancer
type ServerPool struct {
  backends    []*Backend
  current   uint64
}

//relays requests through ReverseProxy
u, _ := url.Parse("http://localhost:8080")
rp := httputil.NewSingleHostReverseProxy(u)


//initializes server and adds handler
http.HandlerFunc(rp.ServeHTTP)

//increments index atomically 
func (s *ServerPool) NextIndex() int {
  return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}




