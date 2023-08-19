package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

type Session struct {
	Webhook string
	last    struct {
		time time.Time
		mu   sync.RWMutex
	}
}

func (session *Session) Message(content string) (err error) {
	session.last.mu.RLock()
	last := session.last.time
	session.last.mu.RUnlock()

	time.Sleep(time.Until(last.Add(2 * time.Second)))

	buf, err := json.Marshal(map[string]string{"content": content})
	if err != nil {
		return
	}

	resp, err := http.Post(session.Webhook, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	session.last.mu.Lock()
	session.last.time = time.Now()
	session.last.mu.Unlock()

	if resp.StatusCode != 204 {
		err = errors.New(resp.Status)
	}

	return
}
