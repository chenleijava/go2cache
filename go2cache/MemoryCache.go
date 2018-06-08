package go2cache

import (
	"sync"
)



type MemoryCache struct {
	//存储读写锁
	lock sync.RWMutex
	//存储元数据
	cacheObjectMap map[string]* CacheObject
}

//初始化 MemoryCache
//当前容器不限制大小
func BuildMemoryCache() *MemoryCache {
	return &MemoryCache{cacheObjectMap: make(map[string]*CacheObject)}
}

////周期检查cacheObject对象是否过期，限制大小，清理过期对象
//func (cache *MemoryCache) check() {
//	go func() {
//		for true {
//			<-time.After(DefaultInternal * time.Second)
//			//check
//			log.Println("begin to clear expired cache object")
//			cache.clearExpiredCacheObject()
//		}
//	}()
//}
//
////清理过期对象
//func (cache *MemoryCache) clearExpiredCacheObject() {
//	lock := cache.lock
//	lock.Lock()
//	defer lock.Unlock()
//	for key, cacheObject := range cache.cacheObjectMap {
//		if cacheObject.IsExpired() {
//			delete(cache.cacheObjectMap, key)
//		}
//	}
//}

//内存中获取数据
func (cache *MemoryCache) Get(key string) *CacheObject {
	lock := cache.lock
	lock.RLock()
	defer lock.RUnlock()
	cacheObj := cache.cacheObjectMap[key]
	if cacheObj == nil {
		return nil
	}
	return cacheObj
}

//存储数据到当前cache中
//timeout 对象有效期 0 永不过期
func (cache *MemoryCache) Put(key string, value interface{}) error {
	lock := cache.lock
	lock.Lock()
	defer lock.Unlock()
	cache.cacheObjectMap[key] =
		&CacheObject{
			Value:        value} //设置过期时间
	return nil
}

//删除缓存数据
func (cache *MemoryCache) Delete(key string) error {
	lock := cache.lock
	lock.Lock()
	defer lock.Unlock()
	delete(cache.cacheObjectMap, key)
	return nil
}

//检查当前key 是否存在
func (cache *MemoryCache) IsExist(key string) bool {
	lock := cache.lock
	lock.RLock()
	defer lock.RUnlock()
	return cache.cacheObjectMap[key] != nil
}
