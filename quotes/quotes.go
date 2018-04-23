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

const API = "https://api.tumblr.com/v2/blog/withyourface.tumblr.com/posts/quote?api_key=%s"
const APIPAGE = "https://api.tumblr.com/v2/blog/withyourface.tumblr.com/posts/quote?api_key=%s&before=%s"

type Quotes struct {
	lock   sync.RWMutex
	APIKey string
	Last   time.Time
	Queue  []string
}

type Response struct {
	Meta     Meta  `json:"meta"`
	Response *Data `json:"response"`
}

type Meta struct {
	Status int `json:"status"`
}

type Data struct {
	Posts []Post `json:"posts"`
}

type Post struct {
	ID        int    `json:"id"`
	Timestamp int    `json:"timestamp"`
	Text      string `json:"text"`
}

func New(key string) *Quotes {
	return &Quotes{
		APIKey: key,
		Last:   time.Now(),
	}
}

func (q *Quotes) ShouldRespond() bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	return time.Since(q.Last) > time.Duration(30 * time.Minute)
}

func (q *Quotes) NextQuote() string {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.Queue) == 0 {
		update(q)
		if len(q.Queue) == 0 {
			return ""
		}
	}

	quote := q.Queue[0]
	q.Queue = q.Queue[1:]
	q.Last = time.Now()

	return quote
}

func update(q *Quotes) {
	posts := make(map[int]Post)
	before := ""

	for {
		page := getPage(q.APIKey, before)
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

	q.Queue = make([]string, len(posts))

	i := 0
	for _, p := range posts {
		q.Queue[i] = lower(p.Text)
		i++
	}

	for i := range q.Queue {
		j := rand.Intn(i + 1)
		q.Queue[i], q.Queue[j] = q.Queue[j], q.Queue[i]
	}
}

func getPage(apiKey, before string) []Post {
	var url string
	var resp Response

	if before == "" {
		url = fmt.Sprintf(API, apiKey)
	} else {
		url = fmt.Sprintf(APIPAGE, apiKey, before)
	}

	r, e := http.Get(url)
	if e != nil {
		fmt.Println("http get error:", e)
		return []Post{}
	}

	e = json.NewDecoder(r.Body).Decode(&resp)
	if e != nil {
		fmt.Println("json decode error:", e)
		return []Post{}
	}

	if resp.Meta.Status != 200 || resp.Response == nil {
		fmt.Println("invalid response:", resp)
		return []Post{}
	}

	return resp.Response.Posts
}

func lower(s string) string {
	s = strings.Trim(s, " ")
	return html.UnescapeString(strings.ToLower(string(s[0])) + string(s[1:]))
}
