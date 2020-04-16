package service

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var localCache *cache.Cache

func init() {
	localCache = cache.New(10*time.Minute, 20*time.Minute)
}
