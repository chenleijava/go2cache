package go2cache

const (
	// memory cache
	LEVEL_1 = iota
	//redis  cache
	LEVEL_2
)

//cache provider interface
type CacheProvider interface {
	//cacheProvider name
	Name() string
	//cache levek
	Level() int
	//build cache
	BuildCache(region string) (interface{}, error)
	// region list
	GetRegions() []Region
}
