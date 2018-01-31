#!/usr/bin/env python3

import steam_deallist.steam_deallist as steam
from steam_deallist.userdata import Game
import sys
import urllib.error
import urllib.request
from bs4 import BeautifulSoup

# bisogna impostare (e ottenere) una chiave per isthereanydeal nel .bashrc
# itad_api = os.environ["ISTHEREANYDEAL_API_KEY"]

def get_checked_list():
    return []

def save_checked_list(l):
    return None

def get_pending_items():
    return []

def save_pending_items(l):
    return None

def game_from_link_and_name(link, name):
    gid = None

    if link is not None:
        gid = steam.get_id_from_store_url(link)
    else:
        # se non c'e` faccio una ricerca su steam in base al nome
        if name is not None:
            gid, link = steam.query_steam_for_game(name)

    if link is not None and gid is not None:
        return Game(gid, 0, 0, 0, link, name, None)

    return None


def parseskidrowcrack(item):
    link = None
    name = None

    # per il titolo devo scandire i <category>, li dentro devo cercare il match "minore" con il titolo, cioe` il subset piu` piccolo del contenuto di <title>
    title = item.find("title").text.lower()
    for c in item.findAll("category"):
        t = c.text
        if title.startswith(t.lower()) and (name is None or len(t) < len(name)):
            name = t

    if name is None:
        return None

    # cerco dentro al <pre> se c'e` un link allo store
    cont = BeautifulSoup(item.find("content:encoded").text, "lxml")
    pre = cont.find("pre")
    if pre is not None:
        toks = pre.text.split()
        for t in toks:
            if "store.steampowered.com" in t:
                link = t
                break

    return game_from_link_and_name(link, name)

def parseskidrowreloaded(item):
    link = None
    name = None

    # prima di tutto devo decodificare il contenuto encoded
    cont = BeautifulSoup(item.find("content:encoded").text, "lxml")

    # cerco il <p> che (trimmato) inizia con "Title: ", tengo come titolo quello che c'e` fino alla mandata a capo (o <br />)
    for p in cont.findAll("p"):
        if p.text.strip().startswith("Title:"):
            name = p.text.splitlines()[0].lstrip("Title:").strip()
            break

    # poi cerco fra tutti i tag <a> se ce n'e` uno con href che inizia per http(s)://store.steampowered
    for a in cont.findAll("a"):
        if "store.steampowered.com" in a['href']:
            link = a['href']
            break

    return game_from_link_and_name(link, name)

def parsefitgirl(item):
    return None

sources = {
        'http://feeds.feedburner.com/SkidrowReloadedGames': parseskidrowreloaded,
        'https://feeds.feedburner.com/skidrowgamesfeed': parseskidrowcrack,
        'http://fitgirl-repacks.com/feed/': parsefitgirl,
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
        print("{}:\n\t{}".format(g.name, g.link))
