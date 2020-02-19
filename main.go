package main

import (
	"golang.org/x/net/html"
	"net/http"
	"log"
	"strings"
	"os"
)

func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	return
}

func crawl(url string, ch chan string, chFinished chan bool) {
	resp, err := http.Get(url)

	defer func() {
		chFinished <- true
	}()

	if err != nil {
		log.Println("Cannot crawl \"" + url + "\" ")
		return
	}

	b := resp.Body
	defer b.Close()

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			t := z.Token()

			isAnchor := t.Data == "a"
			if(!isAnchor) {
				continue
			}

			ok, url := getHref(t)
			if !ok {
				continue
			}

			hasProto := strings.Index(url, "http") == 0 
			if hasProto {
				ch <- url
			}
		}
	}
}

var (
	foundedUrls = make(map[string]bool)
	chkUrls = make(chan string)
	chFinished = make(chan bool)
	seedUrls = os.Args[1:]
)

func run() {
	for _, url := range seedUrls {
		go crawl(url, chkUrls, chFinished)
	}

	for c := 0; c < len(seedUrls); {
		select {
		case url := <- chkUrls:
			foundedUrls[url] = true
		case <- chFinished:
			c++
		}
	}

	log.Println("Found", len(foundedUrls), "external links:")
	for url := range foundedUrls {
		log.Println(" - " + url)
	}
	close(chkUrls)
}

func main() {
	run()
}

