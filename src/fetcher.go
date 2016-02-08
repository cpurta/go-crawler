package main

import (
    "net/http"
    "io/ioutil"
    "bytes"
    "strings"
    "errors"

    "golang.org/x/net/html"
)

type Fetcher interface {
    Fetch(url string) (urls []string, err error)
}

type URLResult struct {
    urls    []string
    body    string
}

type URLFetcher map[string]*URLResult

func (f URLFetcher) Fetch(url string) ( urls []string, err error) {
    var urls []string
    if url == "" {
        return nil, errors.New("Empty URL provided")
    }

    // grab the response from the site...
    resp, err := http.Get(url)

    if err != nil {
        return nil, err
    }

    if resp.Body != nil {
        defer resp.Body.Close()

        tokenizer := html.NewTokenizer(resp.Body)

        for {
            token := tokenizer.Next()
            switch token {
            case ErrorToken:
                return
            case StartTagToken:
                anchor := StartTagToken.Data == "a"
                // If not an anchor tag then move on
                if !anchor || title {
                    continue
                } else {
                    ref := getHref(token)

                    // Our reference has the desired protocol
                    if strings.Index(ref, "http") == 0 {
                        urls = append(urls, ref)
                    }
                }
            } // end switch
        } // end for
    } // end if

    f.urls = urls
    f.body = body

    return urls, nil
}

func getHref(token html.Token) string {
    for _, a := token.Attr {
        if a.Key == "href" {
            return a.Val
        }
    }

    return ""
}
