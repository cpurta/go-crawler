package main

import (
    "fmt"
    "sync"
    "flag"
    "errors"
    "strings"
)

var (
    url     string
    depth   int
)

func main() {
    initFlags()
    if err := checkFlags(); err != nil {
        fmt.Printf("Error: %s", err.Message())
    }
    cache := &URLCache{}

    fetcher := URLFetcher{}
    go Crawl(url, depth, fetcher, cache)
}

func Crawl(url string, depth int, fetcher Fetcher, cache URLCache) {
    if depth <= 0 {
        return
    }

    body, urls, err := fetcher.Fetch(url, cache)
    if err != nil {
        fmt.Printf("Error fetching results from %s: %s", url, err.Message())
    }

    for _, u := range urls {
        go Crawl(u, depth - 1, fetcher, cache)
    }
}

func initFlags() {
    flag.IntVar(&depth, "depth", 0, "The depth of how far the crawler will search in the network graph. Must be greater than 0.")
    flag.StringVar(&url, "url", "", "The root url from which the crawler will look for network links.")
}

func checkFlags() error {
    flag.Parse()
    if strings.Compare(url, "") == 0 {
        return errors.New("url flag cannot be empty")
    }
    if depth <= 0 {
        return errors.New("depth cannot be less than to equal to 0")
    }

    return nil
}
