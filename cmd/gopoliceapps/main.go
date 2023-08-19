package main

import (
	"encoding/json"
	"errors"
	"gopoliceapps/internal/discord"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type config struct {
	Forum   string `json:"forum"`
	Webhook string `json:"webhook"`
}

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)

	configs := []config{}
	if contents, err := os.ReadFile("config.json"); err != nil {
		log.Fatal(err)
	} else {
		if err := json.Unmarshal(contents, &configs); err != nil {
			log.Fatal(err)
		}
	}

	forums := map[string][]string{}
	for _, config := range configs {
		forums[config.Forum] = append(forums[config.Forum], config.Webhook)
	}

	timeout, err := strconv.Atoi(os.Getenv("TIMEOUT"))
	if err != nil {
		log.Fatal(err)
	}

	var (
		apps    = map[string][]string{}
		session = discord.Session{}
		init    bool
	)

	ticker := time.NewTicker(time.Duration(timeout) * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		var (
			newApps  = map[string][]string{}
			messages = map[string]string{}
		)

		for forum, webhooks := range forums {
			doc, err := func() (doc *goquery.Document, err error) {
				resp, err := http.Get(forum)
				if err != nil {
					return
				}
				defer func() {
					if err := resp.Body.Close(); err != nil {
						log.Println(err)
					}
				}()

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

			doc.Find("ol li div h4 span a").Each(func(_ int, s *goquery.Selection) {
				link, exists := s.Attr("href")
				if !exists {
					return
				}

				newApps[forum] = append(newApps[forum], link)
				if !init {
					return
				}

				if slices.Contains(apps[forum], link) {
					return
				}

				for _, webhook := range webhooks {
					if _, ok := messages[webhook]; !ok {
						messages[webhook] = link
						continue
					}

					messages[webhook] += "\n" + link
				}
			})
		}

		for webhook, message := range messages {
			session.Webhook = webhook
			err = session.Message(message)
			if err != nil {
				log.Println(err)
			}
		}

		apps = newApps
		init = true
	}
}
