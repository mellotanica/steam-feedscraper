package games_cache

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"log"
)

type CacheError struct {
	What string
}

func (e CacheError) Error() string {
	return fmt.Sprintf("CacheError: %s", e.What)
}

type Game struct {
	Name string `json:"name"`
	Gid string `json:"gid"`
	Link string `json:"link"`
	Genre string `json:"genre"`
}

func (g Game) String() string {
	return fmt.Sprintf("Title: %s (%s)\n\t%s\n", g.Name, g.Genre, g.Link)
}

func (g Game) Equals(other Game) (bool) {
	return g.Gid == other.Gid && g.Name == other.Name
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
			var games map[string]Game
			err = json.Unmarshal(data, &games)
			if err == nil {
				cache = Cache{filename: fname, games: games, lastUpdate: time.Now()}
			} else {
				cache = Cache{filename: fname, games: make(map[string]Game), lastUpdate: time.Now()}
				log.Print(err)
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

func (c *Cache) GetContent() (list []Game) {
	c.RLock()
	list = make([]Game, len(c.games))
	i := 0
	for _, v := range c.games {
		list[i] = v
		i ++
	}
	c.RUnlock()
	return
}

func (c *Cache) GetFirst() (game Game) {
	c.RLock()
	for _, g := range c.games {
		game = g
		break
	}
	c.RUnlock()
	return
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

func (c *Cache) GameInList(game Game) (alreadyPresent bool) {
	c.RLock()
	g, alreadyPresent := c.games[game.Gid]
	if alreadyPresent {
		alreadyPresent = g.Equals(game)
	}
	c.RUnlock()
	return
}

func (c *Cache) LastUpdate() (t time.Time) {
	c.RLock()
	t = c.lastUpdate
	c.RUnlock()
	return
}

func (c *Cache) Lenght() (size int) {
	c.RLock()
	size = len(c.games)
	c.RUnlock()
	return
}

func (c *Cache) Migrage(target *Cache, gid string, name string) error {
	c.Lock()
	g, ok := c.games[gid]
	if ok {
		if g.Name == name {
			delete(c.games, gid)
		} else {
			ok = false
		}
	}
	c.Unlock()
	if !ok {
		return CacheError{fmt.Sprintf("Cache does not contain game %s", gid)}
	}
	target.AppendElements(g)
	return nil
}