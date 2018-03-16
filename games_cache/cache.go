package games_cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

type CacheError struct {
	What string
}

func (e CacheError) Error() string {
	return fmt.Sprintf("CacheError: %s", e.What)
}

type Game struct {
	Name  string `json:"name"`
	Gid   string `json:"gid"`
	Link  string `json:"link"`
	Genre string `json:"genre"`
}

func (g Game) String() string {
	return fmt.Sprintf("Title: %s (%s)\n\t%s\n", g.Name, g.Genre, g.Link)
}

func (g Game) Equals(other Game) bool {
	return g.Gid == other.Gid && g.Name == other.Name
}

type Cache struct {
	sync.RWMutex
	filename   string
	byIid      map[string]Game
	byName	   map[string]Game
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
		cache = Cache{filename: fname, byIid: make(map[string]Game), byName: make(map[string]Game), lastUpdate: time.Now()}
		data, err := ioutil.ReadFile(fname)
		if err == nil {
			var games []Game
			err = json.Unmarshal(data, &games)
			if err == nil {
				cache.AppendElements(games...)
			} else {
				log.Print(err)
			}
		}
		handler[fname] = cache
	}
	return &cache
}

func (c *Cache) Store() error {
	cont := c.GetContent()
	c.Lock()
	data, err := json.Marshal(cont)
	c.Unlock()

	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.filename, data, 0600)
}

func (c *Cache) GetContent() (list []Game) {
	c.RLock()
	list = make([]Game, len(c.byIid))
	i := 0
	for _, v := range c.byIid {
		list[i] = v
		i++
	}
	c.RUnlock()
	return
}

func (c *Cache) getElement(key string, cache map[string]Game) (*Game, bool) {
	c.RLock()
	g, ok := cache[key]
	c.RUnlock()
	if ok {
		return &g, true
	}
	return nil, false

}

func (c *Cache) GetElementById(gid string) (*Game, bool) {
	return c.getElement(gid, c.byIid)
}

func (c *Cache) GetElementByName(name string) (*Game, bool) {
	return c.getElement(name, c.byName)
}

func (c *Cache) GetElementByNameAndId(name, gid string) (g *Game, ok bool) {
	g, ok = c.GetElementById(gid)
	if ok {
		ok = g.Name == name
	}

	if !ok {
		g = nil
	}
	return
}

func (c *Cache) GetElementByNameOrId(name, gid string) (g *Game, ok bool) {
	ok = false
	g = nil
	switch {
		case len(gid) > 0:
			g, ok = c.GetElementById(gid)
		case len(name) > 0:
			g, ok = c.GetElementByName(name)
	}
	return
}

func (c *Cache) GetFirst() (game Game) {
	c.RLock()
	for _, g := range c.byIid {
		game = g
		break
	}
	c.RUnlock()
	return
}

func (c *Cache) ClearContent() {
	c.Lock()
	c.byName = make(map[string]Game)
	c.byIid = make(map[string]Game)
	c.lastUpdate = time.Now()
	c.Unlock()
}

func (c *Cache) AppendElements(games ...Game) {
	c.Lock()
	for _, g := range games {
		_, id_ok := c.byIid[g.Gid]
		_, name_ok := c.byName[g.Name]
		if !id_ok && !name_ok{
			c.byIid[g.Gid] = g
			c.byName[g.Name] = g
		}
	}
	c.lastUpdate = time.Now()
	c.Unlock()
}

func (c *Cache) GameInList(game Game) (alreadyPresent bool) {
	c.RLock()
	_, alreadyPresent = c.byIid[game.Gid]
	_, ok := c.byName[game.Name]
	alreadyPresent = alreadyPresent || ok

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
	if len(c.byIid) > len(c.byName) {
		size = len(c.byIid)
	} else {
		size = len(c.byName)
	}
	c.RUnlock()
	return
}

func (c *Cache) Migrate(target *Cache, gid string, name string) error {
	c.Lock()
	g, ok := c.byIid[gid]
	done := ok
	if ok {
		if g.Name == name {
			delete(c.byIid, gid)
		} else {
			done = false
		}
	}
	if done {
		g, ok = c.byName[name]
		done = ok
		if ok {
			if g.Gid == gid {
				delete(c.byName, name)
			} else {
				done = false
			}
		}
	}
	c.Unlock()
	if !done {
		return CacheError{fmt.Sprintf("Cache does not contain game %s", gid)}
	}
	target.AppendElements(g)
	return nil
}
