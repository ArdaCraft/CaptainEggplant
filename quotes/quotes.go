package quotes

import (
	"fmt"
	"net/http"
	"encoding/json"
	"math/rand"
	"strings"
	"html"
	"time"
	"sync"
)

type Quotes struct {
	lock   sync.RWMutex
	apiKey string
	last   time.Time
	queue  []string
}

func New(key string) *Quotes {
	return &Quotes{
		apiKey: key,
		last:   time.Now(),
	}
}

func (q *Quotes) ShouldRespond() bool {
	// read only
	q.lock.RLock()
	defer q.lock.RUnlock()

	// check if 30 mins has passed since last message
	return time.Since(q.last) > time.Duration(30 * time.Minute)
}

func (q *Quotes) NextQuote() string {
	// full lock since we may be writing to the queue and/or timestamp
	q.lock.Lock()
	defer q.lock.Unlock()

	// fill queue if empty
	if len(q.queue) == 0 {
		fillQueue(q)
		// still empty ? :(
		if len(q.queue) == 0 {
			return ""
		}
	}

	// drain the first quote from the queue
	quote := q.queue[0]
	q.queue = q.queue[1:]

	// stamp the last usage time
	q.last = time.Now()

	return quote
}

func fillQueue(q *Quotes) {
	// map post id against the post to filter duplicates posts returned by the api pagination
	posts := make(map[int]Post)
	before := ""

	// api responds with 20 items per page
	// use the timestamp of the 20th item as the 'before' value in the next query
	for {
		page := getPage(q.apiKey, before)
		before = ""

		for _, p := range page {
			if _, ok := posts[p.ID]; !ok {
				posts[p.ID] = p
				before = fmt.Sprint(p.ID)
			}
		}

		if before == "" {
			break
		}
	}

	// transfer the unique quotes to the quotes queue
	q.queue = make([]string, len(posts))
	for i, p := range posts {
		q.queue[i] = tidy(p.Text)
	}

	// shuffle the queue
	for i := range q.queue {
		j := rand.Intn(i + 1)
		q.queue[i], q.queue[j] = q.queue[j], q.queue[i]
	}
}

// removes whitespace prefix/suffix and ensures first word is un-capitalized
func tidy(s string) string {
	s = strings.Trim(s, " ")
	return html.UnescapeString(strings.ToLower(string(s[0])) + string(s[1:]))
}
