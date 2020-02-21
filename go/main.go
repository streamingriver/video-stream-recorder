package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/grafov/m3u8"
)

var (
	flagURL    = flag.String("url", "", "url to fetch")
	flagTail   = flag.Int("tail", 24, "how much hours to keep")
	flagHost   = flag.String("host", "http://localhos:8080", "add host to m3u8")
	flagBindTo = flag.String("bind-to", ":8080", "bind to ip:port")

	flagDebug = flag.Bool("debug", false, "")
)

func main() {
	flag.Parse()

	if *flagURL == "" {
		log.Printf("Set url to fetch: ./app --url [url-here]")
		return
	}

	err := os.MkdirAll("./files", 0755)
	if err != nil {
		log.Fatalf("mkdir fail: %v", err)
	}

	database_init()

	if !*flagDebug {
		go fetcher()
	}

	go database_worker()

	router := mux.NewRouter()

	router.HandleFunc("/start/{start}/{limit}/stream.m3u8", func(w http.ResponseWriter, r *http.Request) {

		varz := mux.Vars(r)

		out := "#EXTM3U\n"
		out += "#EXT-X-PLAYLIST-TYPE:VOD\n"
		out += "#EXT-X-TARGETDURATION:20\n"
		out += "#EXT-X-VERSION:4\n"
		out += "#EXT-X-MEDIA-SEQUENCE:0\n"

		items := database_get(varz["start"], varz["limit"])

		for _, item := range items {
			out += fmt.Sprintf("#EXTINF:%f\n", item.Len)
			out += fmt.Sprintf("%s\n", item.Name)
		}

		out += "#EXT-X-ENDLIST\n"

		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Header().Set("Content-Lenght", fmt.Sprintf("%d", len(out)))
		w.Write([]byte(out))
	})

	router.HandleFunc("/start/{start}/{limit}/{ts:.+}", func(w http.ResponseWriter, r *http.Request) {
		varz := mux.Vars(r)
		w.Header().Set("Content-Type", "text/vnd.trolltech.linguist")
		b, err := ioutil.ReadFile(fmt.Sprintf("./files/%s", varz["ts"]))
		if err != nil {
			log.Printf("error %v", err)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))
		w.Write(b)

	})

	log.Printf("Starting server on %v", *flagBindTo)
	log.Fatal(http.ListenAndServe(*flagBindTo, router))
}

func fetcher() {
	mainurl, _ := url.Parse(*flagURL)
	for {
	start_at:
		b := fetch(mainurl.String())
		buf := bytes.NewBuffer(b)
		pl, pt, err := m3u8.Decode(*buf, true)
		if err != nil {
			log.Printf("fetcher error: %v %v", mainurl.String(), err)
			time.Sleep(1 * time.Second)
			continue
		}
		if pt == m3u8.MASTER {
			masterpl := pl.(*m3u8.MasterPlaylist)
			for _, variant := range masterpl.Variants {
				mainurl, _ = mainurl.Parse(variant.URI)
				log.Printf("%v", mainurl.String())
				goto start_at
			}

		} else if pt == m3u8.MEDIA {
			mediapl := pl.(*m3u8.MediaPlaylist)
			for _, segment := range mediapl.Segments {
				if segment == nil {
					continue
				}
				fetchurl, _ := mainurl.Parse(segment.URI)
				fetchurl.RawQuery = mainurl.RawQuery
				if cache_set(fetchurl.String()) {
					log.Printf("%v", fetchurl.String())
					currenttime := time.Now().UnixNano()
					item := &DatabaseItem{
						Name: fmt.Sprintf("%v.ts", currenttime),
						Len:  segment.Duration,
						T:    currenttime,
					}
					database_store(item)

					b := fetch(fetchurl.String())
					if b != nil {
						err := ioutil.WriteFile("./files/"+item.Name, b, 0755)
						if err != nil {
							log.Printf("error on write file to fs %v", err)
							continue
						}
					}
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func fetch(url string) []byte {
	hc := http.Client{Timeout: 10 * time.Second}

	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("User-Agent", "iptv/1.0")

	response, err := hc.Do(request)
	if err != nil {
		log.Printf("fetch error %v %v", url, err)
		return nil
	}
	defer response.Body.Close()
	if response.StatusCode/100 != 2 {
		log.Printf("Invalid response code %v %v", url, response.StatusCode)
		return nil
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil
	}
	return b
}
