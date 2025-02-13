package shared

import (
	"sync"
)

type RateLimiter struct {
	Visits map[string]int
	Mutex  sync.Mutex
}

var Limiter = &RateLimiter{Visits: make(map[string]int)}
