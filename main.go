package main 

import (
  "fmt"
  "net/http"
)

type Backend struct {
  URL   *url.URL
  Alive   bool
  mux   sync.RWMutex
  ReverseProxy    *httputil.ReverseProxy
}


