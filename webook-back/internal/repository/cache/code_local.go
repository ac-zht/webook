package cache

import (
	"context"
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2"
	"sync"
	"time"
)

type LocalCodeCache struct {
	cache      *lru.Cache[string, CodeItem]
	lock       *sync.Mutex
	expiration time.Duration
}

type CodeItem struct {
	code     string
	cnt      int
	deadline time.Time
}

func (l *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	key := l.key(biz, phone)
	codeItem, ok := l.cache.Get(key)
	//一分钟内已发送过
	if ok && codeItem.deadline.Sub(time.Now()) > 4*time.Minute {
		return ErrCodeSendTooMany
	}
	l.cache.Add(key, CodeItem{
		code:     code,
		cnt:      3,
		deadline: time.Now().Add(l.expiration),
	})
	return nil
}

func (l *LocalCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	key := l.key(biz, phone)
	codeItem, ok := l.cache.Get(key)
	if !ok {
		return false, nil
	}
	if codeItem.cnt <= 0 {
		return false, ErrCodeVerifyTooManyTimes
	}
	if codeItem.code != inputCode {
		codeItem.cnt--
		l.cache.Add(key, codeItem)
		return false, nil
	}
	return true, nil
}

func (l *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

func NewLocalCodeCache(cache *lru.Cache[string, CodeItem], expiration time.Duration) CodeCache {
	return &LocalCodeCache{
		cache:      cache,
		lock:       &sync.Mutex{},
		expiration: expiration,
	}
}
