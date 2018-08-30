package go2cache

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"testing"
	"time"
)

func TestGetCacheChannel(t *testing.T) {
	cacheChannel := GetCacheChannel()
	cache := cacheChannel.GetRedisCache("user_region")
	var key = time.Now().Format("2006-01-02 15:04:05")
	var filed = key
	v := cache.HincrBy(key, filed, 1)
	log.Printf("hincyBy v:%d", v)
	hget, _ := redis.Int(cache.Hget(key, filed), nil)
	log.Printf("Hget v:%d", hget)
	intMap := cache.HgetAllIntMap(key)
	log.Printf("HgetAllIntMap:%d", intMap[key])
	for k := range intMap {
		delete(intMap, k)
	}
	cache.Hset(key, "1", "this is test")
	vv := cache.Hget(key, "1") // Get bytes array
	log.Printf("hset value:%s", string(vv.([]byte)))
}
