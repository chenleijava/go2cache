package go2cache

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

//redigo doc
//https://godoc.org/github.com/garyburd/redigo/redis

type RedisCache struct {
	redisClient *redis.Pool
	region      string // region   -->  redis_name_space+":"+region
}

//send msg to redis
func (cache *RedisCache) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	args[0] = cache.region + ":" + args[0].(string) //[0]上数据是key ，这里进行key的拼接形成最终的key为   region:key ,同 j2cache保持一致
	con := cache.redisClient.Get()
	defer con.Close()
	return con.Do(commandName, args...)
}

//获取 cacheObject
func (cache *RedisCache) Get(key string) *CacheObject {
	reply, err := cache.do("GET", key)
	if err != nil {
		log.Printf("get bytes with key:%s error:%s ", key, err)
		return nil
	}
	return &CacheObject{Value: reply}
}

//redis 缓存中获取byte[]
func (cache *RedisCache) GetBytes(key string) (reply interface{}, err error) {
	return cache.do("GET", key)
}

//存储数据到当前cache中
//timeout 对象有效期 0 永不过期
func (cache *RedisCache) Put(key string, value interface{}) error {
	_, err := cache.do("SET", key, value)
	return err
}

//删除缓存数据
func (cache *RedisCache) Delete(key string) error {
	_, err := cache.do("DEL", key)
	return err
}

//检查当前key 是否存在
func (cache *RedisCache) IsExist(key string) bool {
	result, err := redis.Bool(cache.do("EXISTS", key))
	if err != nil {
		return false
	}
	return result
}

//计数 +1
func (cache *RedisCache) Incr(key string) error {
	_, err := redis.Bool(cache.do("INCRBY", key, 1))
	return err
}

//todo 根据业务需求，新增 redis的 通用方法
