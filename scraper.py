#!/usr/bin/env python3

import steam_deallist.steam_deallist as steam
import sys
import urllib.error
import urllib.request
import json
from bs4 import BeautifulSoup

genre_blacklist = [
    'tower defense',
    'racing',
    'race'
]
cache_file = "feeds_cache"

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

def get_checked_list():
    global cache_file

    return []

def save_checked_list(l):
    global cache_file

    return None

def get_pending_items():
    global cache_file

    return []

def save_pending_items(l):
    global cache_file

    return None

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

def update_all():
    checked_list = get_checked_list()
    items = get_pending_items()
    
    for s in sources.keys():
        try:
            soup = BeautifulSoup(urllib.request.urlopen(s), 'lxml-xml')
            items.extend([g for g in [sources[s](i) for i in soup.findAll("item")] if g is not None and g.gid not in checked_list])
        except urllib.error.HTTPError as e:
            print("feed source {} is unreachable: {}".format(s, e), file=sys.stderr)

    save_pending_items(items)

    return items

if __name__ == "__main__":
    for g in update_all():
        print(g)
