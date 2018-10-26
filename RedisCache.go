package go2cache

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"log"
)

//const (
//	PING, SET, GET, QUIT, EXISTS, DEL, TYPE, FLUSHDB, KEYS, RANDOMKEY,
// RENAME, RENAMENX, RENAMEX, DBSIZE, EXPIRE, EXPIREAT, TTL, SELECT, MOVE,
// FLUSHALL, GETSET, MGET, SETNX, SETEX, MSET, MSETNX, DECRBY, DECR, INCRBY,
// INCR, APPEND, SUBSTR, HSET, HGET, HSETNX, HMSET, HMGET, HINCRBY, HEXISTS,
// HDEL, HLEN, HKEYS, HVALS, HGETALL, RPUSH, LPUSH, LLEN, LRANGE, LTRIM,
// LINDEX, LSET, LREM, LPOP, RPOP, RPOPLPUSH, SADD, SMEMBERS, SREM, SPOP,
// SMOVE, SCARD, SISMEMBER, SINTER, SINTERSTORE, SUNION, SUNIONSTORE, SDIFF,
// SDIFFSTORE, SRANDMEMBER, ZADD, ZRANGE, ZREM, ZINCRBY, ZRANK, ZREVRANK,
// ZREVRANGE, ZCARD, ZSCORE, MULTI, DISCARD, EXEC, WATCH, UNWATCH, SORT,
// BLPOP, BRPOP, AUTH, SUBSCRIBE, PUBLISH, UNSUBSCRIBE, PSUBSCRIBE, PUNSUBSCRIBE,
// PUBSUB, ZCOUNT, ZRANGEBYSCORE, ZREVRANGEBYSCORE, ZREMRANGEBYRANK,
// ZREMRANGEBYSCORE, ZUNIONSTORE, ZINTERSTORE, ZLEXCOUNT, ZRANGEBYLEX,
// ZREVRANGEBYLEX, ZREMRANGEBYLEX, SAVE, BGSAVE, BGREWRITEAOF,
// LASTSAVE, SHUTDOWN, INFO, MONITOR, SLAVEOF, CONFIG, STRLEN,
// SYNC, LPUSHX, PERSIST, RPUSHX, ECHO, LINSERT, DEBUG, BRPOPLPUSH, SETBIT,
// GETBIT, BITPOS, SETRANGE, GETRANGE, EVAL, EVALSHA, SCRIPT, SLOWLOG, OBJECT,
// BITCOUNT, BITOP, SENTINEL, DUMP, RESTORE, PEXPIRE, PEXPIREAT, PTTL, INCRBYFLOAT,
// PSETEX, CLIENT, TIME, MIGRATE, HINCRBYFLOAT, SCAN, HSCAN, SSCAN, ZSCAN, WAIT, CLUSTER,
// ASKING, PFADD, PFCOUNT, PFMERGE, READONLY, GEOADD, GEODIST, GEOHASH, GEOPOS,
// GEORADIUS, GEORADIUSBYMEMBER, BITFIELD
//)

//redigo doc
//https://godoc.org/github.com/garyburd/redigo/redis

type RedisCache struct {
	redisClient *redis.Pool
	region      string // region   -->  redis_name_space+":"+region
}


//direct Do command
func (cache *RedisCache) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	con := cache.redisClient.Get()
	defer con.Close()
	return con.Do(commandName, args...)
}


//send msg to redis
func (cache *RedisCache) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	args[0] = cache.region + ":" + args[0].(string) //[0]上数据是key ，这里进行key的拼接形成最终的key为   region:key ,同 j2cache保持一致
	con := cache.redisClient.Get()
	defer con.Close()
	return con.Do(commandName, args...)
}

//获取 cacheObject
func (cache *RedisCache) GetCacheObject(key string) *CacheObject {
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
func (cache *RedisCache) Del(key string) error {
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

//获取 key 对应的值
func (cache *RedisCache) Get(key string) interface{} {
	reply, _ := cache.do("GET", key)
	return reply
}

//hincryBy 基于hash 计数
func (cache *RedisCache) HincrBy(key, filed string, value int) int64 {
	v, _ := redis.Int64(cache.do("HINCRBY", key, filed, value))
	return v
}

//HSET
func (cache *RedisCache) Hset(key, filed string, value interface{}) int64 {
	v, _ := redis.Int64(cache.do("HSET", key, filed, value))
	return v
}

//HGET
func (cache *RedisCache) Hget(key, filed string) interface{} {
	reply, _ := cache.do("HGET", key, filed)
	return reply
}

//HDEL
func (cache *RedisCache) Hdel(key, filed string) interface{} {
	reply, _ := cache.do("HDEL", key, filed)
	return reply
}


//https://blog.csdn.net/weixin_37696997/article/details/78634043
//SADD
func (cache *RedisCache) SAdd(key, member interface{}) interface{} {
	reply, _ := cache.do("SADD", key, member)
	return reply
}

//SREM
func (cache *RedisCache) SRem(key, member interface{}) interface{} {
	reply, _ := cache.do("SREM", key, member)
	return reply
}

//smembers
//[]interface{} is data array , elm is []byte
//for _,v:=range []insterface{}
// {
//   v.([]byte)
//   do something
// }
func (cache *RedisCache) Smembers(key interface{}) []interface{} {
	//  Reply type      Result
	//  array           reply, nil
	//  nil             nil, ErrNil
	//  other           nil, error
	value, err := redis.Values(cache.do("SMEMBERS", key))
	if err != nil {
		return nil
	}
	return value
}

// string set ---
func (cache *RedisCache) SmembersString(key interface{}) *[]string {
	//  Reply type      Result
	//  array           reply, nil
	//  nil             nil, ErrNil
	//  other           nil, error
	value, err := redis.Values(cache.do("SMEMBERS", key))
	if err != nil {
		return nil
	}
	var dd = make([]string, len(value))
	for index, v := range value {
		dd[index] = string(v.([]byte))
	}
	return &dd
}

//sismember
func (cache *RedisCache) Sismember(key, member interface{}) bool {
	reply, _ := cache.do("sismember", key, member)
	return reply.(int64) == 1
}

//HGETALL
func (cache *RedisCache) HgetAllStringMap(key string) map[string]string {
	mp, err := redis.StringMap(cache.do("HGETALL", key))
	if err != nil {
		log.Printf("Hgetall error:%s", err.Error())
	}
	return mp
}

//HGETALL
func (cache *RedisCache) HgetAllIntMap(key string) map[string]int {
	mp, err := redis.IntMap(cache.do("HGETALL", key))
	if err != nil {
		log.Printf("Hgetall error:%s", err.Error())
	}
	return mp
}

//HGETALL
func (cache *RedisCache) HgetAllBytesMap(key string) map[string][]byte {
	mp, err := bytesMap(cache.do("HGETALL", key))
	if err != nil {
		log.Printf("Hgetall error:%s", err.Error())
	}
	return mp
}

//HLen
func (cache *RedisCache) Hlen(key string) int {
	l, _ := redis.Int(cache.do("HLEN", key))
	return l
}

//return base data array ?
func bytesMap(result interface{}, err error) (map[string][]byte, error) {
	values, err := redis.Values(result, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redigo: StringMap expects even number of values result")
	}
	m := make(map[string][]byte, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, okKey := values[i].([]byte)
		value, okValue := values[i+1].([]byte)
		if !okKey || !okValue {
			return nil, errors.New("redigo: StringMap key not a bulk string value")
		}
		m[string(key)] = value
	}
	return m, nil
}
