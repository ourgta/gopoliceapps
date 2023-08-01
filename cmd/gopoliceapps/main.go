package main

import (
	"errors"
	"gopoliceapps/internal/discord"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/exp/slices"
)

func env(key string) (value string) {
	value = os.Getenv(key)
	if value == "" {
		log.Fatalf("%v no provided", strings.ToLower(key))
	}

	return
}

func main() {
	t := env("TIMEOUT")
	timeout, err := strconv.Atoi(t)
	if err != nil {
		log.Fatal(err)
	}

	webhook := env("WEBHOOK")
	session := discord.Session{
		Webhook: webhook,
	}

	var apps []string
	var last time.Time
	init := false

	for {
		time.Sleep(time.Until(last.Add(time.Duration(timeout) * time.Minute)))

		doc, err := func() (doc *goquery.Document, err error) {
			resp, err := http.Get("https://grandtheftarma.com/forum/64-applications/")
			if err != nil {
				return
			}

			last = time.Now()

			defer func(resp *http.Response) {
				err := resp.Body.Close()
				if err != nil {
					log.Println(err)
				}
			}(resp)

			if resp.StatusCode != 200 {
				err = errors.New(resp.Status)
				return
			}

			doc, err = goquery.NewDocumentFromReader(resp.Body)
			return
		}()

		if err != nil {
			log.Println(err)
			continue
		}

		var newApps []string
		selector := "ol li div h4 span a"
		doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
			link, exists := s.Attr("href")
			if !exists {
				return
			}

			newApps = append(newApps, link)
			if !init {
				return
			}

			if slices.Contains(apps, link) {
				return
			}

			err := session.Message(link)
			if err != nil {
				log.Println(err)
			}
		})

		apps = newApps
		init = true
	}
}
