package quotes

import (
	"sync"
	"time"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"math/rand"
)

type Quotes struct {
	lock   sync.RWMutex
	invoke time.Time
	auto   time.Time
	queue  []string
}

func New() *Quotes {
	return &Quotes{
		auto:   time.Now(),
		invoke: time.Now(),
	}
}

func (q *Quotes) CanInvoke() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return time.Since(q.invoke) > time.Duration(10*time.Second)
}

func (q *Quotes) CanRespond() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return time.Since(q.auto) > time.Duration(12*time.Hour)
}

func (q *Quotes) Cooldown()  {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.auto = time.Now()
}

func (q *Quotes) NextInvoke() string {
	q.lock.Lock()
	defer q.lock.Unlock()
	quote := next(q)
	q.invoke = time.Now()
	return quote
}

func (q *Quotes) NextResponse() string {
	q.lock.Lock()
	defer q.lock.Unlock()
	quote := next(q)
	q.auto = time.Now()
	return quote
}

func next(q *Quotes) string {
	if len(q.queue) == 0 {
		fillQueue(q)
		// still empty ? :(
		if len(q.queue) == 0 {
			return ""
		}
	}
	quote := q.queue[0]
	q.queue = q.queue[1:]
	return quote
}

func fillQueue(q *Quotes) {
	d, e := ioutil.ReadFile("data.json")
	if e != nil {
		fmt.Println("read error:", e)
		return
	}

	e = json.Unmarshal(d, &q.queue)
	if e != nil {
		fmt.Println("unmarshal error:", e)
		return
	}

	// shuffle the queue
	for i := range q.queue {
		j := rand.Intn(i + 1)
		q.queue[i], q.queue[j] = q.queue[j], q.queue[i]
	}
}