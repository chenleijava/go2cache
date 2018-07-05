package go2cache

import (
	"io/ioutil"
	"log"
	"gopkg.in/yaml.v2"
	"strings"
	"strconv"
	"sync"
	"github.com/golang/protobuf/proto"
	"os/exec"
	"os"
)

//cache channel ref
var cacheChannel *CacheChannel
var synOnce sync.Once

//cache channel single object
func GetCacheChannel() *CacheChannel {
	if cacheChannel == nil {
		synOnce.Do(func() {
			//init cache channel
			cacheChannel = &CacheChannel{}
			cacheChannel.initCacheChannel()
		})
	}
	return cacheChannel
}

//缓存操作入口
type CacheChannel struct {
	//mm provider
	mmp *MemoryProvider
	//redis provider
	rdp *RedisProvider
	//region collections
	ragionArray []Region
	//redis get mutex
	readDataMutex sync.Mutex
	//redis pub and sub
	psb *PubSub
}

// redis cache config
type Go2CacheRedis struct {
	ConnectInfo    string `yaml:"connect_info"`
	DbIndex        string `yaml:"db_index"`
	Password       string `yaml:"password"`
	MaxIdle        int    `yaml:"max_idle"`
	MaxActive      int    `yaml:"max_active"`
	IdleTimeOut    int    `yaml:"idle_time_out"`
	Channel        string `yaml:"channel"`
	RedisNameSpace string `yaml:"redis_name_space"`
}

//region config
type Go2CacheConfig struct {
	Go2CacheRegions []string      `yaml:"go2cache_regions"`
	Go2CacheRedis   Go2CacheRedis `yaml:"go2cache_redis"`
}

func exePath() string {
	execPath, _ := exec.LookPath(os.Args[0])
	log.Printf("path:%s", execPath)
	return execPath[0:strings.LastIndex(execPath, "/")]
}

//init go2cache
func (c *CacheChannel) initCacheChannel() error {
	path := exePath()
	//loading config
	bytes, err := ioutil.ReadFile(path + "/config/go2cache.yaml")
	if err != nil {
		bb, e := ioutil.ReadFile(path + "/resources/config/go2cache.yaml")
		if e != nil {
			log.Fatalf("go2cache loading config error:%s ,go2cache.yaml file must in exe path : '/config' or '/resources/config' ", err.Error())
		}
		bytes = bb
	}
	var config Go2CacheConfig
	e := yaml.Unmarshal(bytes, &config)
	if e != nil {
		log.Fatal("go2cache yaml Unmarshal error:", e)
	}
	//get go2cache config
	regions := config.Go2CacheRegions
	//cache regions
	for _, region := range regions {
		region_src := strings.Split(region, ",")
		//first is name ,second is size
		if len(region_src) != 2 {
			continue
		}
		size, err := strconv.Atoi(region_src[1])
		if err != nil {
			log.Fatal("go2cache region size conv error:", err)
		}
		c.ragionArray = append(c.ragionArray, Region{Name: region_src[0], Size: size})
	}
	//init memory cache
	if len(c.ragionArray) != 0 {
		//init MemoryProvider and RedisProvider
		c.mmp = new(MemoryProvider)
		c.rdp = new(RedisProvider)
		//init redis cache
		c.rdp.InitRedisCache(&config.Go2CacheRedis)
		//init psc with redis
		c.psb = &PubSub{Client: c.rdp.redisClient, Channel: config.Go2CacheRedis.Channel, CacheChannel: c}
		c.psb.Subscribe()
		//redis name space
		c.rdp.redisNameSpace = config.Go2CacheRedis.RedisNameSpace
		for _, region := range c.ragionArray {
			c.mmp.BuildCache(region.Name)
			log.Println("build memory cahce region :", region.String())
			c.rdp.BuildCache(region.Name)
			log.Println("build redis cahce region :", region.String())
			//done
		}
	}
	return nil
}

//base protobuf struck ,read from level1 cache
func (c *CacheChannel) GetLevel1(region, key string) interface{} {
	memoryCache, _ := c.mmp.BuildCache(region)
	cacheObject := memoryCache.(*MemoryCache).Get(key)
	if cacheObject == nil {
		return nil
	}
	return cacheObject.Value
}

//read from level2 cache
func (c *CacheChannel) GetBytesLevel2(region, key string) (reply interface{}, err error) {
	redisCache, _ := c.rdp.BuildCache(region)
	return redisCache.(*RedisCache).GetBytes(key)
}

//Get redis cache by region
func (c *CacheChannel) GetRedisCache(region string) *RedisCache {
	redisCache, _ := c.rdp.BuildCache(region)
	return redisCache.(*RedisCache)
}

//base protobuf struck ,read from level2 cache
//may be return nil
func (c *CacheChannel) GetProtoBufLevel2(region, key string, message proto.Message) interface{} {
	//must be mutex ,in case of more i/o in redis cache
	c.readDataMutex.Lock()
	defer c.readDataMutex.Unlock()

	memoryCache, _ := c.mmp.BuildCache(region)
	cacheObject := memoryCache.(*MemoryCache).Get(key)
	if cacheObject != nil {
		return cacheObject.Value
	}
	//unhit level1 ,try to read form level2 cache
	redisCache, _ := c.rdp.BuildCache(region)
	bytes, err := redisCache.(*RedisCache).GetBytes(key)
	if err != nil {
		log.Printf("GetProtoBufLevel2 error:%s  key:%s", err, key)
		return nil
	} else if bytes != nil {
		proto.Unmarshal(bytes.([]byte), message)
		memoryCache.(*MemoryCache).Put(key, message)
		return message
	} else {
		return nil
	}
}

//base protobuf .set into cache
//pusub notify other node (x2cache) evict level1 cache
func (c *CacheChannel) SetProtoBuf(region, key string, message proto.Message) {
	bytes, marshalError := proto.Marshal(message)
	if marshalError != nil {
		log.Printf("proto Marshal error:%s", marshalError)
		return
	}
	memoryCache, _ := c.mmp.BuildCache(region)
	memoryCache.(*MemoryCache).Put(key, message) //pb struck
	redisCache, _ := c.rdp.BuildCache(region)
	err := redisCache.(*RedisCache).Put(key, bytes)
	if err != nil {
		log.Printf("put cahe error:%s", err)
		return
	}
	//clear leve1 cache
	c.sendEvictCmd(region, key)
}

//set into cache
//pusub notify other node (x2cache) evict level1 cache
func (c *CacheChannel) Set(region, key string, value interface{}) {
	memoryCache, _ := c.mmp.BuildCache(region)
	memoryCache.(*MemoryCache).Put(key, value)
	redisCache, _ := c.rdp.BuildCache(region)
	err := redisCache.(*RedisCache).Put(key, value)
	if err != nil {
		log.Printf("put cahe error:%s", err)
		return
	}
	//clear leve1 cache
	c.sendEvictCmd(region, key)
}

//send evict cmd to nodes
func (c *CacheChannel) sendEvictCmd(region, key string) {
	log.Println("发送清理指令:", region+"@"+key)
	c.psb.SendEvictCmd(region, key)
}

//evict level1 cache
func (c *CacheChannel) Evict(region string, keys []string) {
	cache, _ := c.mmp.BuildCache(region)
	for _, key := range keys {
		cache.(*MemoryCache).Delete(key)
	}
}

//get all regions
func (c *CacheChannel) GetRegions() []Region {
	return c.ragionArray
}

//mm provider
func (c *CacheChannel) GetMemoryProvider() *MemoryProvider {
	return c.mmp
}

//redis provider
func (c *CacheChannel) GetRedisProvider() *RedisProvider {
	return c.rdp
}
