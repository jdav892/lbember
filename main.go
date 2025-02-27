package main 

import (
  "context"
  "flag"
  "fmt"
  "log"
  "net"
  "net/http/httputil"
  "net/http"
  "net/url"
  "strings"
  "sync"
  "sync/atomic"
  "time"
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


proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
  log.Printf("[%s] %s\n", serverUrl.Host, e.Error())
  retries := GetRetryFromContext(request)
  if retries < 3 {
    select {
    case <-time.After(10 * time.Millisecond):
      ctx := context.WithValue(request.Context(), Retry, retries+1)
      proxy.ServeHttp(writer, request.WithContext(ctx))
    }
    return 
  }

  //after 3 retries, make this backend as down
  serverPool.MarkBackendStatus(serverUrl, false)
  //if the same request routing for few attempts with different backends, increase the count
  attempts := GetAttemptsFromContext
  log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
  ctx := context.WithValue(request.Context(), Attempts, attempts+1)
  lb(writer, requet.WithContext(ctx))
}

//lb balances the incoming request
func lb(w http.ResponseWriter, r *http.Request) {
  attempts := GetAttemptsFromContext(r)
  if attempts > 3 {
    log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
    http.Error(w, "Service not available", http.StatusServiceUnavailable)
    return
  }


  peer := serverPool.GetNextPeer()
  if peer != nil {
    peer.ReverseProxy.ServeHttp(w, r)
    return
  }
  http.Error(w, "Service not available", http.StatusServiceUnavailable)
}


func main() {

  //passes method to HandlerFunc
  server := http.Server{
    Address: fmt.Sprintf(":d%", port),
    Handler: http.HandlerFunc(lb),
  }
}
