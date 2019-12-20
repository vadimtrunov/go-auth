package session

import (
	"fmt"
	"go-auth/src/store"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	if err := store.OpenDatabase("../data/teststore.db"); err != nil {
		panic("cant open database")
	}
}

func TestAddTokenToSession(t *testing.T) {
	sessionStore = make(map[string]Session)
	flushStream = flushStream[:0]

	test := func(n int) {
		token := fmt.Sprintf("some_token_%v", n)
		expireAt := time.Now().Add(2 * time.Second).Unix()

		session := Create("test@test.com", expireAt)
		added := session.Add(token)

		assert.True(t, added)

		exists, _ := Get(token)
		assert.True(t, exists)
	}

	var wg sync.WaitGroup
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func(n int) {
			test(n)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestGetTokensFromSession(t *testing.T) {
	sessionStore = make(map[string]Session)
	flushStream = flushStream[:0]

	test := func(n int) {
		token := fmt.Sprintf("some_token_%v", n)

		ok, s := Get(token)

		assert.True(t, ok)
		assert.Equal(t, "test@test.com", s.emial)
	}

	for i := 0; i < 1000; i++ {
		token := fmt.Sprintf("some_token_%v", i)

		session := Create("test@test.com", time.Now().Add(2*time.Second).Unix())
		session.Add(token)
	}

	var wg sync.WaitGroup
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func(n int) {
			test(n)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestGetInvalidTokenFromSession(t *testing.T) {
	sessionStore = make(map[string]Session)
	flushStream = flushStream[:0]

	for i := 0; i < 5; i++ {
		token := fmt.Sprintf("some_token_%v", i)

		session := Create("test@test.com", time.Now().Add(2*time.Second).Unix())
		session.Add(token)
	}

	token := "some_invalid_token"

	ok, s := Get(token)

	assert.False(t, ok)
	assert.Nil(t, s)
}

func TestGC(t *testing.T) {
	sessionStore = make(map[string]Session)
	flushStream = flushStream[:0]

	for i := 0; i < 3; i++ {
		token := fmt.Sprintf("some_token_%v", i)

		session := Create("test@test.com", time.Now().Add(2*time.Minute).Unix())
		session.Add(token)
	}
	for i := 3; i < 5; i++ {
		token := fmt.Sprintf("some_token_%v", i)

		session := Create("test@test.com", 0)
		session.Add(token)
	}
	for i := 5; i < 7; i++ {
		token := fmt.Sprintf("some_token_%v", i)

		session := Create("test@test.com", time.Now().Add(2*time.Minute).Unix())
		session.Add(token)
	}

	garbageCollector(time.Now().Unix())

	assert.Contains(t, sessionStore, "some_token_0")
	assert.Contains(t, sessionStore, "some_token_1")
	assert.Contains(t, sessionStore, "some_token_2")

	assert.NotContains(t, sessionStore, "some_token_3")
	assert.NotContains(t, sessionStore, "some_token_4")

	assert.Contains(t, sessionStore, "some_token_5")
	assert.Contains(t, sessionStore, "some_token_6")
}

func TestScheduler(t *testing.T) {
	sessionStore = make(map[string]Session)
	flushStream = flushStream[:0]

	prepareBolt(t)

	for i := 0; i < 3; i++ {
		token := fmt.Sprintf("some_token_%v", i)

		session := Create("test@test.com", time.Now().Add(2*time.Minute).Unix())
		session.Add(token)
	}
	for i := 3; i < 5; i++ {
		token := fmt.Sprintf("some_token_%v", i)

		session := Create("test@test.com", 0)
		session.Add(token)
	}
	for i := 5; i < 7; i++ {
		token := fmt.Sprintf("some_token_%v", i)

		session := Create("test@test.com", time.Now().Add(2*time.Minute).Unix())
		session.Add(token)
	}

	notif := scheduler(1)

	assert.True(t, <-notif)

	assert.Contains(t, sessionStore, "some_token_0")
	assert.Contains(t, sessionStore, "some_token_1")
	assert.Contains(t, sessionStore, "some_token_2")

	assert.NotContains(t, sessionStore, "some_token_3")
	assert.NotContains(t, sessionStore, "some_token_4")

	assert.Contains(t, sessionStore, "some_token_5")
	assert.Contains(t, sessionStore, "some_token_6")

	tokens := store.GetAllRenewTokens()

	assert.Equal(t, 5, len(tokens))

}

func TestFlushToPersistentStorage(t *testing.T) {
	sessionStore = make(map[string]Session)
	flushStream = make([]sessionItem, 0)

	prepareBolt(t)

	for i := 0; i < 5; i++ {
		token := fmt.Sprintf("some_token_%v", i)

		session := Create("test@test.com", time.Now().Add(2*time.Minute).Unix())
		session.Add(token)
	}

	flush()

	tokens := store.GetAllRenewTokens()

	assert.Equal(t, 5, len(tokens))
}

func prepareBolt(t *testing.T) {
	if err := store.DropDatabase(); err != nil {
		t.FailNow()
	}
	if err := store.CreateDefaultBacket(); err != nil {
		t.FailNow()
	}
}
