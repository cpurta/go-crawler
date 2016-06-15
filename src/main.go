package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"

	"gopkg.in/redis.v3"
)

var (
	seedUrl string
	search  string
	depth   int

	validURL *regexp.Regexp

	// Environment variable from docker-compose
	redisHost string
	redisPort string
)

func main() {
	initFlags()
	if err := checkFlags(); err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	loadEnvironmentVariables()

	client := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "",
		DB:       0,
	})

	if pong, err := client.Ping().Result(); err != nil {
		log.Fatalf("%s. Cannot connect to redis client: %s", pong, err.Error())
	}

	if search != "" {
		validURL = regexp.MustCompile(search)
	}

	fetcher := URLFetcher{}
	go Crawl(seedUrl, depth, fetcher, client)
}

func Crawl(searchUrl string, depth int, fetcher Fetcher, client *redis.Client) {
	if depth <= 0 {
		return
	}

	host, err := url.Parse(searchUrl)

	// Send this to our redis queue for indexing
	if err != nil {
		client.LPush("unknown_url_crawler_queue", searchUrl)
	} else {
		client.LPush(host.Host+"_crawler_queue", searchUrl)
	}

	urls, err := fetcher.Fetch(searchUrl)
	if err != nil {
		fmt.Printf("Error fetching results from %s: %s", searchUrl, err.Error())
	}

	for _, u := range urls {
		if validURL.MatchString(u) {
			go Crawl(u, depth-1, fetcher, client)
		}
	}
}

func initFlags() {
	flag.IntVar(&depth, "depth", 0, "The depth of how far the crawler will search in the network graph. Must be greater than 0.")
	flag.StringVar(&seedUrl, "seed-url", "", "The root url from which the crawler will look for network links.")
	flag.StringVar(&search, "search", "", `Regex that will be used against the urls crawled. Only urls matching the regex will be crawled. e.g. ^http(s)?://cnn.com\?+([0-9a-zA-Z]=[0-9a-zA-Z])$`)
}

func checkFlags() error {
	flag.Parse()
	if seedUrl == "" {
		return errors.New("url flag cannot be empty")
	}
	if depth <= 0 {
		return errors.New("depth cannot be less than to equal to 0")
	}

	return nil
}

func loadEnvironmentVariables() {
	redisHost = os.Getenv("REDIS_PORT_6379_TCP_ADDR")
	redisPort = os.Getenv("REDIS_PORT_6379_TCP_PORT")

	if redisHost != "" && redisPort != "" {
		fmt.Printf("Redis found on %s:%s\n", redisHost, redisPort)
	}
}
