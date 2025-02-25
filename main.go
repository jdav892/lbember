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

//Set alive for this backend
func (b *Backend) SetAlive(alive bool){
  b.mux.Lock()
  b.Alive = alive
  b.mux.Unlock()
}

//IsAlive returns true when backend is alive
func (b *Backend) isAlive() (alive bool) {
  b.mux.RLock()
  alive = b.Alive
  b.mux.RUnlock()
  return
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
fmt.Println("Server running")

//increments index atomically 
func (s *ServerPool) NextIndex() int {
  return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

//GetNextPeer returns the next active peer to take a connection
func (s *ServerPool) GetNextPeer() *Backend {
  //loop entire backends to find out an Alive backend
  next := s.NextIndex()
  l := len(s.backends) + next //start from next and move a full cycle
  for i := next; i < l; i++ {
    idx := i % len(s.backends) //take an index by modding with length
    //if backend is alive, use it and store if its not the original
    if s.backends[idx].IsAlive() {
      if != next {
        atomic.StoreUint64(&s.current, uint64(idx)) //to mark current one
      }
      return s.backends[idx]
    }
  }
  return nil
}
