package games_cache

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Game struct {
	Name string `json:"name"`
	Gid string `json:"gid"`
	Link string `json:"link"`
	Genre string `json:"genre"`
}

func (g Game) String() string {
	return fmt.Sprintf("Title: %s (%s)\n\t%s\n", g.Name, g.Genre, g.Link)
}

type Cache struct {
	filename string
	games []Game
	lastUpdate time.Time
	mutex sync.Mutex
}

const MediumCacheSize = 30
const GamesCachePendingFile = "feeds_cache_pending"
const GamesCacheDubiousFile = "feeds_cache_dubious"
const GamesCacheCheckedFile = "feeds_cache_checked"

type cacheHandler map[string]Cache

var handler = cacheHandler(nil)

func LoadCache(fname string) *Cache {
	if handler == nil {
		handler = make(map[string]Cache)
	}

	cache, ok := handler[fname]

	if !ok {
		data, err := ioutil.ReadFile(fname)
		if err != nil {
			cache = Cache{fname, make([]Game, 0, MediumCacheSize), time.Now(), sync.Mutex{}}
		} else {
			var list []Game
			err = json.Unmarshal(data, &list)
			if err != nil {
				cache = Cache{fname, make([]Game, 0, MediumCacheSize), time.Now(), sync.Mutex{}}
			} else {
				cache = Cache{fname, list, time.Now(), sync.Mutex{}}
			}
		}
		handler[fname] = cache
	}
	return &cache
}

func (c *Cache) Store() error {
	c.mutex.Lock()
	data, err := json.Marshal(c.games)
	if err != nil {
		return err
	}
	c.mutex.Unlock()

	return ioutil.WriteFile(c.filename, data, 0600)
}

func (c *Cache) GetContent() []Game {
	c.mutex.Lock()
	res := make([]Game, len(c.games))
	copy(res, c.games)
	c.mutex.Unlock()
	return res
}

func (c *Cache) ClearContent() {
	c.mutex.Lock()
	c.games = c.games[:0]
	c.lastUpdate = time.Now()
	c.mutex.Unlock()
}

func (c *Cache) AppendElements(games ...Game) {
	c.mutex.Lock()
	c.games = append(c.games, games...)
	c.lastUpdate = time.Now()
	c.mutex.Unlock()
}

func (c *Cache) LastUpdate() time.Time {
	c.mutex.Lock()
	t := c.lastUpdate
	c.mutex.Unlock()
	return t
}
