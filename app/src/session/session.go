package session

import (
	"go-auth/src/store"
	"sync"
	"time"
)

type operation int

var stream chan *sessionItem
var streamMutex sync.RWMutex
var flushStream []sessionItem

var sessionStore map[string]Session

const (
	read operation = iota
	write
	gc
)

//Session stuct for session data storing
type Session struct {
	emial    string
	expireAt int64
}

type sessionItem struct {
	session  *Session
	token    string
	op       operation
	feedback chan bool
}

func init() {
	stream = make(chan *sessionItem)
	flushStream = make([]sessionItem, 0)
	sessionStore = make(map[string]Session)

	go runtime()

	scheduler(60)
}

func runtime() {
	for {
		switch itm := <-stream; itm.op {
		case read:
			result, ok := sessionStore[itm.token]
			if ok {
				itm.session = &result
			}
			itm.feedback <- ok
		case write:
			sessionStore[itm.token] = *itm.session
			itm.feedback <- true
		case gc:
			garbageCollector(time.Now().Unix())
		}
	}
}

func scheduler(interval int) chan bool {
	notifier := make(chan bool)
	go func() {
		for range time.Tick(time.Duration(interval) * time.Second) {
			stream <- &sessionItem{
				op: gc,
			}

			flush()
			notifier <- true
		}
	}()
	return notifier
}

func garbageCollector(now int64) {
	for token, session := range sessionStore {
		if session.expireAt < now {
			delete(sessionStore, token)
		}
	}
}

func flush() {
	now := time.Now().Unix()
	for _, itm := range flushStream {
		if now < itm.session.expireAt {
			store.AddRenewToken(itm.token, itm.session.expireAt)
		}
	}
	flushStream = make([]sessionItem, 0)
}

//Create cunstructor for session struct
func Create(email string, expireAt int64) *Session {
	var s Session
	s.emial = email
	s.expireAt = expireAt
	return &s
}

//Add puts toket into session
func (s *Session) Add(token string) bool {
	itm := sessionItem{
		session:  s,
		token:    token,
		op:       write,
		feedback: make(chan bool, 1),
	}

	stream <- &itm
	if ok := <-itm.feedback; ok {
		flushStream = append(flushStream, itm)
	}

	return true
}

//Get get serssion form sessionStore
func Get(token string) (bool, *Session) {
	itm := sessionItem{
		session:  nil,
		token:    token,
		op:       read,
		feedback: make(chan bool, 1),
	}

	stream <- &itm
	if ok := <-itm.feedback; !ok {
		return false, nil
	}

	return true, itm.session
}
