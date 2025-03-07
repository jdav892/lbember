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

//implements constants incrementally, each containing a unique value
const (
  Attmepts int = iota 
  Retry
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
      if i != next {
        atomic.StoreUint64(&s.current, uint64(idx)) //to mark current one
      }
      return s.backends[idx]
    }
  }
  return nil
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

//returns the attempts for request
func GetRetryFromContext(r *http.Request) int {
  if retry, ok := r.Context().Value(Retry).(int); ok {
    return retry
  }
  return 0
}

//isAlive checks whether a backend is alive by establishing a tcp connection
func isBackendAlive(u *url.URL) bool {
  timeout := 2 * time.Second
  conn, err := net.DialTimeout("tcp", u.Host, timeout)
  if err != nil {
    log.Println("Site unreachable, error: ", err)
    return false
  }
  _ = conn.Close()//close it as we don't need to maitnain the connection
    return true
  )
}

func (s *ServerPool) HealthCheck() {
  for _, b := range s.backends {
    status := "up"
    alive := isBackendAlive(b.URL)
    b.SetAlive(alive)
    if !alive {
      status = "down"
    }
    log.Printf("%s [%s]\n", b.URL, status)
  }
}


func healthCheck() {
  t := time.NewTicker(time.Second * 20)
  for {
    select {
    case <-t.C:
      log.Println("Starting health check...")
      serverPool.HealthCheck
      log.pringln("Health check completed")
    }
  }
}

go healthCheck()
//TODO Refactor to a more efficient and working solution


func main() {
  //relays requests through ReverseProxy
  u, _ := url.Parse("http://localhost:8080")
  rp := httputil.NewSingleHostReverseProxy(u)
  
  proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
  retries := GetRetryFromContext(request)
  log.Printf("[%s] %s\n", serverUrl.Host, e.Error())
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

  //initializes server and adds handler
  http.HandlerFunc(rp.ServeHTTP)
  fmt.Println("Server running")
  
  //passes method to HandlerFunc
  server := http.Server{
    Address: fmt.Sprintf(":d%", port),
    Handler: http.HandlerFunc(lb),
  }
}
