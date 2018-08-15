package go2cache

import (
	"testing"
	"time"
	"log"
	"github.com/garyburd/redigo/redis"
)

func TestGetCacheChannel(t *testing.T) {
	SetConfigPath("/config/go2cache.yaml")
	cacheChannel := GetCacheChannel()
	cache := cacheChannel.GetRedisCache("user_region")
	var key = time.Now().Format("2006-01-02 15:04:05")
	var filed = key
	v := cache.HincrBy(key, filed, 1)
	log.Printf("hincyBy v:%d", v)
	hget, _:= redis.Int(cache.Hget(key, filed), nil)
	log.Printf("Hget v:%d", hget)
	intMap := cache.HgetAllIntMap(key)
	log.Printf("HgetAllIntMap:%d", intMap[key])
	for k := range intMap {
		delete(intMap, k)
	}
}
