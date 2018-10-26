package go2cache

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"strconv"
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
	cache.Hset(key, "2", "this is test")
	vv := cache.Hget(key, "2") // Get bytes array
	obj := cache.Hdel(key, "2")
	vv = cache.Hget(key, "2") // Get bytes array

	log.Printf("%s", obj)

	cache.Hset(key, "2", "this is test")

	vv = cache.Hget(key, "2") // Get bytes array
	log.Printf("hset value:%s", string(vv.([]byte)))
	cache.HgetAllBytesMap(key)

	l := cache.Hlen(key)

	log.Printf("len:%d", l)

	key = "key2"
	cache.SAdd(key, 1)
	cache.SAdd(key, 2)

	setLen, e := redis.Int64(cache.Do("SCARD", "go2cache:user_region:key2"))
	if e != nil {
		log.Fatal(e.Error())
	}
	log.Printf("scard len :%d", setLen)

	bb := cache.Sismember(key, 1)
	log.Printf("Sismember:%t", bb)
	smembers := cache.Smembers(key)
	for _, dd := range smembers {
		x, _ := strconv.Atoi(string(dd.([]byte)))
		log.Printf("data:%d", x)
	}

	sa := cache.SmembersString(key)
	for _, v := range *sa {
		log.Printf("value %s", v)
	}
	sa = nil
	cache.Del(key)

}
