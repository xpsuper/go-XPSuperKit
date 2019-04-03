package XPSuperKit

import (
	"sync"
	"time"
	"errors"
	"container/list"
)

type XPMemoryCacheImpl struct {
	capacity   int             //最大缓存项
	bucket     *CacheContainer //缓存容器
	objectPool *sync.Pool      //对象池
	rwMutex    *sync.RWMutex   //过期列表读写锁
}

type CacheContainer struct {
	bucket  map[string]*list.Element //缓存容器
	lruList *list.List //节点链表结果
}

type CacheEntry struct {
	Key   string
	Value interface{}
	Expiration int64
}

func NewMemoryCache(count int, cap int) *XPMemoryCacheImpl {
	return &XPMemoryCacheImpl{
		capacity   : cap,
		bucket     : &CacheContainer{ bucket: make(map[string]*list.Element, count), lruList: list.New()},
		rwMutex    : new(sync.RWMutex),
		objectPool : &sync.Pool{
			New : func() interface {} {
				return &CacheEntry{}
			},
		},
	}
}

//判断指定的元素是否过期，如果过期则添加到过期列表中
func (memoryCache *XPMemoryCacheImpl) pushExpiredCacheEntry(el *list.Element,forcibly bool) bool {
	item, ok := el.Value.(*CacheEntry)
	//如果缓存已过期则转移到过期列表中暂存

	if time.Now().Unix() > item.Expiration || forcibly {
		memoryCache.bucket.lruList.Remove(el)
		if ok {
			delete(memoryCache.bucket.bucket, XPString().MD5(item.Key))
			memoryCache.objectPool.Put(&item)
		}
		return true
	}

	memoryCache.bucket.lruList.MoveToBack(el)
	return false
}

func (memoryCache *XPMemoryCacheImpl) Count() int {
	if memoryCache.bucket == nil {
		return 0;
	}

	return memoryCache.bucket.lruList.Len()
}

func (memoryCache *XPMemoryCacheImpl) Contains(key string) bool {
	if memoryCache.bucket == nil {
		return false
	}

	memoryCache.rwMutex.RLock()

	if el,ok := memoryCache.bucket.bucket[XPString().MD5(key)]; ok {
		memoryCache.rwMutex.RUnlock()
		memoryCache.rwMutex.Lock()
		defer memoryCache.rwMutex.Unlock()

		if timeout := memoryCache.pushExpiredCacheEntry(el,false); timeout {
			return false
		}

		return ok
	}

	memoryCache.rwMutex.RUnlock()

	return false

}

func (memoryCache *XPMemoryCacheImpl) Get(key string) (interface{} ,bool) {
	if memoryCache.bucket == nil {
		return nil, false
	}

	memoryCache.rwMutex.Lock()
	defer memoryCache.rwMutex.Unlock()

	if el,ok := memoryCache.bucket.bucket[XPString().MD5(key)]; ok {
		if ok = memoryCache.pushExpiredCacheEntry(el,false); ok {
			return nil, false
		}

		return el.Value.(*CacheEntry).Value, true
	}

	return nil, false
}

func (memoryCache *XPMemoryCacheImpl) GetCacheEntry(key string) (*CacheEntry, bool){
	if memoryCache.bucket == nil {
		return &CacheEntry{}, false
	}

	memoryCache.rwMutex.Lock()
	defer memoryCache.rwMutex.Unlock()

	if el,ok := memoryCache.bucket.bucket[XPString().MD5(key)]; ok {
		if ok = memoryCache.pushExpiredCacheEntry(el,false); ok {
			return &CacheEntry{},false
		}

		return el.Value.(*CacheEntry),true
	}

	return &CacheEntry{}, false
}

func (memoryCache *XPMemoryCacheImpl) Set(key string, value interface{}, duration time.Duration) error {
	if memoryCache.bucket == nil {
		return errors.New("缓存容器没有初始化")
	}

	memoryCache.rwMutex.Lock()
	defer memoryCache.rwMutex.Unlock()

	if el,ok := memoryCache.bucket.bucket[XPString().MD5(key)]; ok {
		if item,ok := el.Value.(*CacheEntry); ok {
			item.Value      = value
			item.Expiration = time.Now().Add(duration).Unix()
			return nil
		}
	}

	var cacheItem *CacheEntry
	var ok bool

	cacheElement := memoryCache.objectPool.Get()


	if cacheItem,ok = cacheElement.(*CacheEntry); ok == false{
		cacheItem = &CacheEntry{}
	}

	cacheItem.Key        = key
	cacheItem.Value      = value
	cacheItem.Expiration = time.Now().Add(duration).Unix()


	el := memoryCache.bucket.lruList.PushFront(cacheItem)
	memoryCache.bucket.bucket[XPString().MD5(key)] = el

	if memoryCache.bucket.lruList.Len() > memoryCache.capacity {
		temp := memoryCache.bucket.lruList.Back()
		memoryCache.pushExpiredCacheEntry(temp,true)
	}

	return nil
}

func (memoryCache *XPMemoryCacheImpl) Remove(key string) (interface{},bool) {
	if memoryCache.bucket == nil {
		return nil,false
	}

	memoryCache.rwMutex.RLock()

	if el,ok := memoryCache.bucket.bucket[XPString().MD5(key)]; ok {
		memoryCache.rwMutex.RUnlock()
		memoryCache.rwMutex.Lock()
		defer memoryCache.rwMutex.Unlock()

		delete(memoryCache.bucket.bucket, XPString().MD5(key))
		memoryCache.bucket.lruList.Remove(el)

		if item,ok := el.Value.(*CacheEntry); ok {
			memoryCache.objectPool.Put(&item)
		}

		return el.Value.(*CacheEntry).Value, true
	}

	return nil, false
}

func (memoryCache *XPMemoryCacheImpl) Clear() {
	memoryCache.rwMutex.Lock()
	defer memoryCache.rwMutex.Unlock()

	memoryCache.bucket.lruList = list.New()
	memoryCache.bucket.bucket = make(map[string]*list.Element, memoryCache.capacity)
}