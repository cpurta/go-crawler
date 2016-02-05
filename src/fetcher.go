package main

import (
    "fmt"
    "net/http"
    "strings"
    "errors"
)

type Fetcher interface {
    Fetch(url string) (body string, urls []string, err error)
}

type NetResult struct {
    urls    []string
    body    string
}

type NetFetcher map[string]*NetResult

func (n NetFetcher) Fetch(url string) (body string, urls []string, err error) {
    if strings.Compare("", url) == 0 {
        return nil, nil, errors.New("Empty URL provided")
    }

    
}
