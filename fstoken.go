package feishu

import (
	"sync"
	"time"
)

type fsToken struct {
	val        string
	expiration time.Time
	min        time.Duration

	sync.Mutex
}

func (t *fsToken) get() string {
	return t.val
}

func (t *fsToken) set(val string, lifetime time.Duration, min time.Duration) {
	t.val = val
	t.expiration = time.Now().Add(lifetime)
	t.min = min
}

func (t *fsToken) isEmpty() bool {
	return t.val == ""
}

func (t *fsToken) notExpired() bool {
	return t.expiration.Sub(time.Now()) > t.min
}
