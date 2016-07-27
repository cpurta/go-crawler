package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"gopkg.in/redis.v4"
)

var (
	seedUrl string
	search  string
	depth   int

	validURL *regexp.Regexp

	// Environment variable from docker-compose
	redisHost  string
	redisPort  string
	influxHost string
	influxPort string

	crawlError = errors.New("already crawled")

	urlTest = regexp.MustCompile(`^((http[s]?):\/)?\/?([^:\/\s]+)((\/\w+)*\/)([\w\-\.]+[^#?\s]+)(.*)?(#[\w\-]+)?$`)
)

var urlcache = struct {
	m    map[string]error
	lock sync.Mutex
}{m: make(map[string]error)}

func main() {
	initFlags()
	if err := checkFlags(); err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	loadEnvironmentVariables()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "",
		DB:       0,
	})

	if pong, err := redisClient.Ping().Result(); err != nil {
		log.Fatalf("%s. Cannot connect to redis client: %s", pong, err.Error())
	}

	influxClient, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%s", influxHost, influxPort),
		Username: "root",
		Password: "root",
	})

	if err != nil {
		log.Fatalf("Error occured when connecting to InfluxDB", err.Error())
	}

	if search != "" {
		validURL = regexp.MustCompile(search)
	}

	fetcher := URLFetcher{}
	Crawl(seedUrl, depth, fetcher, redisClient, influxClient)
}

func Crawl(searchUrl string, depth int, fetcher Fetcher, redisClient *redis.Client, influxClient influx.Client) {
	if depth <= 0 {
		return
	}

	fmt.Printf("Depth: %d Crawling: %s\n", depth, searchUrl)

	bp, _ := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  "crawler",
		Precision: "s",
	})

	host, err := url.Parse(searchUrl)

	// Send this to our redis queue for indexing
	if err != nil {
		redisClient.LPush("unknown_url_crawler_queue", searchUrl)
	} else {
		redisClient.LPush(host.Host+"_crawler_queue", searchUrl)
	}

	urlcache.lock.Lock()
	urlcache.m[searchUrl] = crawlError
	urlcache.lock.Unlock()

	// let's determine how long it is taking to fetch all urls on a page
	startFetch := time.Now()
	urls, err := fetcher.Fetch(searchUrl)
	crawlTime := time.Since(startFetch)

	if err != nil {
		fmt.Printf("Error fetching results from %s: %s\n", searchUrl, err.Error())
	}

	tags := map[string]string{
		"domain": host.String(),
	}

	fields := map[string]interface{}{
		"urls_found":         len(urls),
		"crawl_time":         crawlTime.Nanoseconds(),
		"total_urls_crawled": len(urlcache.m),
		"urls_by_page":       len(urls),
	}

	for _, u := range urls {
		// check our cache to make sure that we are not about to crawl
		// a page we have already visted
		if !urlTest.MatchString(u) {
			u = "http://" + host.Host + u
		}

		urlcache.lock.Lock()
		_, crawled := urlcache.m[u]
		urlcache.lock.Unlock()

		if validURL.MatchString(u) && urlTest.MatchString(u) && !crawled {
			Crawl(u, depth-1, fetcher, redisClient, influxClient)
		}
	}

	point, _ := influx.NewPoint(
		"crawl_usage",
		tags,
		fields,
		time.Now(),
	)

	// add data point to influx
	bp.AddPoint(point)

	if err := influxClient.Write(bp); err != nil {
		log.Printf("Unable to write batch point to influxdb: %s\n", err.Error())
	}
}

func initFlags() {
	flag.IntVar(&depth, "depth", 0, "The depth of how far the crawler will search in the network graph. Must be greater than 0.")
	flag.StringVar(&seedUrl, "seed-url", "", "The root url from which the crawler will look for network links.")
	flag.StringVar(&search, "search", "^.*$", `Regex that will be used against the urls crawled. Only urls matching the regex will be crawled. e.g. ^http(s)?://cnn.com\?+([0-9a-zA-Z]=[0-9a-zA-Z])$`)
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

	influxHost = os.Getenv("INFLUXDB_PORT_8086_TCP_ADDR")
	influxPort = os.Getenv("INFLUXDB_PORT_8086_TCP_PORT")

	if influxHost != "" && influxPort != "" {
		fmt.Printf("InfluxDB found on %s:%s\n", influxHost, influxPort)
	} else if influxHost == "" && influxPort == "" {
		log.Fatalln("Unable to load InfluxDB environment variables!")
	}
}
