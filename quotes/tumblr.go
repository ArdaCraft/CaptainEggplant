package quotes

import (
	"fmt"
	"net/http"
	"encoding/json"
)

const first = "https://api.tumblr.com/v2/blog/withyourface.tumblr.com/posts/quote?api_key=%s"
const page = "https://api.tumblr.com/v2/blog/withyourface.tumblr.com/posts/quote?api_key=%s&before=%s"

type APIResponse struct {
	Meta     Meta      `json:"meta"`
	Response *Response `json:"response"`
}

type Meta struct {
	Status int `json:"status"`
}

type Response struct {
	Posts []Post `json:"posts"`
}

type Post struct {
	ID        int    `json:"id"`
	Timestamp int    `json:"timestamp"`
	Text      string `json:"text"`
}

func getPage(apiKey, before string) []Post {
	var url string

	// get most recent 20 posts, or those before the provided 'before' timestamp
	if before == "" {
		url = fmt.Sprintf(first, apiKey)
	} else {
		url = fmt.Sprintf(page, apiKey, before)
	}

	// http get
	r, e := http.Get(url)
	if e != nil {
		fmt.Println("http get error:", e)
		return []Post{}
	}

	// unmarshal the response
	var resp APIResponse
	e = json.NewDecoder(r.Body).Decode(&resp)
	if e != nil {
		fmt.Println("json decode error:", e)
		return []Post{}
	}

	// check response is valid
	if resp.Meta.Status != 200 || resp.Response == nil {
		fmt.Println("invalid response:", resp)
		return []Post{}
	}

	// return posts
	return resp.Response.Posts
}