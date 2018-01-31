#!/usr/bin/env python3

import steam_deallist.steam_deallist as steam
import sys
import urllib.error
import urllib.request
import os
import json
from bs4 import BeautifulSoup

genre_blacklist = [
    'tower defense',
    'racing',
    'race',
]
cache_prefix = "feeds_cache"

class Game:
    def __init__(self, name, gid, link, genre):
        self.name = name
        self.gid = gid
        self.link = link
        self.genre = genre

    def __eq__(self, other):
        return (self.name == other.name and self.gid == other.gid)

    def __str__(self):
        return "{} ({}):\n{}".format(self.name, self.genre, self.link)

    def to_dict(self):
        return {
            "name": self.name,
            "gid": self.gid,
            "link": self.link,
            "genre": self.genre
        }

    @staticmethod
    def from_dict(d):
        return Game(d['name'], d['gid'], d['link'], d['genre'])

class Cache:
    PENDING_SUFF = "_pending"
    PENDING = 0
    CHECKED_SUFF = "_checked"
    CHECKED = 1

    def __init__(self, prefix):
        self.prefix = prefix
        if os.path.isfile(prefix+Cache.PENDING_SUFF):
            self.load(Cache.PENDING)
        else:
            self.pending = []

        if os.path.isfile(prefix+Cache.CHECKED_SUFF):
            self.load(Cache.CHECKED)
        else:
            self.checked = []

    def load(self, list):
        if list is Cache.PENDING:
            self.pending = Cache.load_json(self.prefix+Cache.PENDING_SUFF)
        elif list is Cache.CHECKED:
            self.checked = Cache.load_json(self.prefix+Cache.CHECKED_SUFF)

    @staticmethod
    def load_json(fname):
        f = open(fname, 'r')
        try:
            d = [Game.from_dict(g) for g in json.load(f)]
        except:
            d = None
        f.close()
        return d

    def save(self):
        Cache.save_json(self.prefix+Cache.PENDING_SUFF, self.pending)
        Cache.save_json(self.prefix+Cache.CHECKED_SUFF, self.checked)

    @staticmethod
    def save_json(fname, c_list):
        if os.path.isfile(fname):
            mode = "w"
        else:
            mode = "x"

        f = open(fname, mode)
        json.dump([g.to_dict() for g in c_list], f)
        f.close()


def game_from_link_and_name(link, name, genre):
    gid = None

    if link is not None:
        gid = steam.get_id_from_store_url(link)
    else:
        # se non c'e` faccio una ricerca su steam in base al nome
        if name is not None:
            gid, link = steam.query_steam_for_game(name)

    if link is not None and gid is not None:
        return Game(name, gid, link, genre)

    return None


def parseskidrowcrack(item):
    link = None
    name = None
    genre = None

    # per il titolo devo scandire i <category>, li dentro devo cercare il match "minore" con il titolo, cioe` il subset piu` piccolo del contenuto di <title>
    title = item.find("title").text.lower()
    for c in item.findAll("category"):
        t = c.text
        if title.startswith(t.lower()) and (name is None or len(t) < len(name)):
            name = t

    if name is None:
        return None

    # cerco dentro al <pre> se c'e` un link allo store e il genere
    cont = BeautifulSoup(item.find("content:encoded").text, "lxml")
    pre = cont.find("pre")
    if pre is not None:
        lines = pre.text.splitlines()
        for l in lines:
            if "store.steampowered.com" in l:
                for t in l.split():
                    if "store.steampowered.com" in t:
                        link = t
            if 'Genre:' in l:
                genre = l[l.index("Genre:")+len("Genre:"):].strip()
            if genre is not None and link is not None:
                break
    else:
    # altrimenti il genere dovrebbe essere in un <p> generico
        for p in cont.findAll("p"):
            if "Genre:" in p.text:
                for l in p.text.splitlines():
                    if "Genre:" in l:
                        genre = l[l.index("Genre:")+len("Genre:"):].strip()
                        break
            if genre is not None:
                break

    return game_from_link_and_name(link, name, genre)

def parseskidrowreloaded(item):
    link = None
    name = None
    genre = None

    # prima di tutto devo decodificare il contenuto encoded
    cont = BeautifulSoup(item.find("content:encoded").text, "lxml")

    # cerco il <p> che (trimmato) inizia con "Title: ", tengo come titolo quello che c'e` fino alla mandata a capo (o <br />)
    for p in cont.findAll("p"):
        if p.text.strip().startswith("Title:"):
            name = p.text.splitlines()[0].lstrip("Title:").strip()
        if "Genre:" in p.text:
            for l in p.text.splitlines():
                if "Genre:" in l:
                    genre = l[l.index("Genre:")+len("Genre:"):].strip()
        if genre is not None and name is not None:
            break

    # poi cerco fra tutti i tag <a> se ce n'e` uno con href che inizia per http(s)://store.steampowered
    for a in cont.findAll("a"):
        if "store.steampowered.com" in a['href']:
            link = a['href']
            break

    return game_from_link_and_name(link, name, genre)

def parsefitgirl(item):
    return None

sources = {
        'http://feeds.feedburner.com/SkidrowReloadedGames': parseskidrowreloaded,
        'https://feeds.feedburner.com/skidrowgamesfeed': parseskidrowreloaded,
        'http://fitgirl-repacks.com/feed/': parsefitgirl,
        'https://feeds.feedburner.com/skidrowgames': parseskidrowcrack,
        'http://feeds.feedburner.com/skidrowcrack': parseskidrowcrack
}

def update_all(cache):
    global genre_blacklist

    items = []
    items.extend(cache.pending)

    for s in sources.keys():
        try:
            soup = BeautifulSoup(urllib.request.urlopen(s), 'lxml-xml')
            for i in soup.findAll("item"):
                game = sources[s](i)
                if game is None or game in items or game in cache.checked:
                    continue

                genre_ok = True
                if game.genre is not None:
                    for g in genre_blacklist:
                        if g.lower() in game.genre.lower():
                            genre_ok = False
                            break

                if genre_ok:
                    items.append(game)

        except urllib.error.HTTPError as e:
            print("feed source {} is unreachable: {}".format(s, e), file=sys.stderr)

    cache.pending = items
    cache.save()

    return items

if __name__ == "__main__":
    for g in update_all(Cache(cache_prefix)):
        print(g)
