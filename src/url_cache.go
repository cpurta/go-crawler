package main

import (
    "fmt"
    "sync"
)

type URLCache struct {
    urls    map[string]bool
    mux     sync.Mutex
}

func (c *URLCache) CacheKey(url string) {
    c.mux.Lock()
    if _, ok := c.urls[url]; !ok {
        c.urls[url] = true
    }
    c.mux.Unlock()
}

func (c *URLCache) Exists(url string) bool {
    res := false
    c.mux.Lock()
    if _, ok := c.urls[url]; ok {
        res := true
    }
    c.mux.Unlock()
    return res
}
