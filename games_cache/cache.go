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
	sync.RWMutex
	filename string
	games map[string]Game
	lastUpdate time.Time
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
			cache = Cache{filename: fname, games: make(map[string]Game), lastUpdate: time.Now()}
		} else {
			var list []Game
			err = json.Unmarshal(data, &list)
			cache = Cache{filename: fname, games: make(map[string]Game), lastUpdate: time.Now()}
			if err == nil {
				cache.AppendElements(list...)
			}
		}
		handler[fname] = cache
	}
	return &cache
}

func (c *Cache) Store() error {
	c.Lock()
	data, err := json.Marshal(c.games)
	if err != nil {
		return err
	}
	c.Unlock()

	return ioutil.WriteFile(c.filename, data, 0600)
}

func (c *Cache) GetContent() []Game {
	c.RLock()
	res := make([]Game, len(c.games))
	i := 0
	for _, v := range c.games {
		res[i] = v
		i ++
	}
	c.RUnlock()
	return res
}

func (c *Cache) ClearContent() {
	c.Lock()
	c.games = make(map[string]Game)
	c.lastUpdate = time.Now()
	c.Unlock()
}

func (c *Cache) AppendElements(games ...Game) {
	c.Lock()
	for _, g := range games {
		_, ok := c.games[g.Gid]
		if !ok {
			c.games[g.Gid] = g
		}
	}
	c.lastUpdate = time.Now()
	c.Unlock()
}

func (c *Cache) LastUpdate() time.Time {
	c.RLock()
	t := c.lastUpdate
	c.RUnlock()
	return t
}
