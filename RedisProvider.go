package go2cache

import (
	"github.com/garyburd/redigo/redis"
	"sync"
	"time"
	"log"
)

//基于redis缓存提供者
type RedisProvider struct {
	//redisCacheMap lock
	mapLock sync.Mutex
	//redis cache
	redisCacheMap map[string]*RedisCache
	// pool
	redisClient *redis.Pool
	//regions
	regions []Region
	//redis name space
	redisNameSpace string
}

//init redis cahce
func (p *RedisProvider) InitRedisCache(config *Go2CacheRedis) {
	p.mapLock.Lock()
	defer p.mapLock.Unlock()
	p.redisCacheMap = make(map[string]*RedisCache)
	p.redisClient = &redis.Pool{
		MaxActive:   config.MaxActive,
		MaxIdle:     config.MaxIdle,
		IdleTimeout: time.Duration(config.IdleTimeOut) * time.Second,
		//test on borrow
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			c, e := redis.Dial("tcp", config.ConnectInfo)
			if e != nil {
				log.Printf("connect redis error:%s", e)
				//循环尝试链接redis，直到成功
				var wait sync.WaitGroup
				wait.Add(1)
				var count = 0
				go func() {
					for {
						c, e = redis.Dial("tcp", config.ConnectInfo)
						if e != nil {
							//连接失败
							log.Printf("the %d times try connect address:%s but  failed :%s ", count, config.ConnectInfo, e)
							count++
							time.Sleep(3 * time.Second)
						} else {
							//连接成功
							log.Printf("try to connect address:%s successful.... ...", config.ConnectInfo)
							wait.Done()
							break
						}
					}
				}()
				wait.Wait() //等待redis 连接成功之后 在继续逻辑
			}
			if config.Password != "" {
				if _, err := c.Do("AUTH", config.Password); err != nil {
					c.Close()
					return nil, err
				}
			}
			if _, err := c.Do("SELECT", config.DbIndex); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		}}
}

// build cache
func (p *RedisProvider) BuildCache(region string) (interface{}, error) {
	p.mapLock.Lock()
	defer p.mapLock.Unlock()
	region = p.redisNameSpace + ":" + region
	cache := p.redisCacheMap[region]
	if cache == nil {
		cache = &RedisCache{
			redisClient: p.redisClient,
			region:      region}
		p.redisCacheMap[region] = cache
		p.regions = append(p.regions, Region{Name: region})
	}
	return cache, nil
}

// 缓存 等级
func (p *RedisProvider) Level() int {
	return LEVEL_2
}

//region name default  go2cache_redis
func (p *RedisProvider) Name() string {
	return "go2cache_redis_provider"
}

//获取region 列表
func (p *RedisProvider) GetRegions() []Region {
	return nil
}
