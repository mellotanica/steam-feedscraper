#!/usr/bin/env python3

import steam_deallist.steam_deallist as steam
import os
import urllib.request
from bd4 import import BeautifulSoup

def get_checked_list():
    return []

def save_checked_list(l):
    return None

def get_pending_items():
    return []

def save_pending_items(l):
    return None

def parseskidrowcrack(item):
    # prima di tutto cerco dentro al <pre> se c'e` un link allo store 

    # se non lo trovo devo scandire i <category>, tirando fuori il testo (e` un commento, tipo <![CDATA[...]]>)
    # li dentro devo cercare il match "minore" con il titolo, cioe` il subset piu` piccolo del contenuto di <title>
    # poi faccio la ricerca su steam con il titolo

    return None

def parseskidrowreloaded(item):
    # prima di tutto cerco fra tutti i tag <a> se ce n'e` uno con href che inizia per http(s)://store.steampowered

    # se non lo trovo cerco il <p> che (trimmato) inizia con "Title: ", tengo come titolo quello che c'e` fino alla mandata a capo (o <br />)
    # poi faccio la ricerca su steam con il titolo

    return None

def parsefitgirl(item):
    return None

sources = {
        'http://feeds.feedburner.com/SkidrowReloadedGames': parseskidrowreloaded,
        'https://feeds.feedburner.com/skidrowgamesfeed': parseskidrowcrack,
        'http://fitgirl-repacks.com/feed/': parsefitgirl,
        'http://feeds.feedburner.com/skidrowcrack': parseskidrowcrack
}

def update_all():
    # bisogna impostare (e ottenere) una chiave per isthereanydeal nel .bashrc
    itad_api = os.environ[ISTHEREANYDEAL_API_KEY]

    checked_list = get_checked_list()
    items = get_pending_items()
    
    for s in sources.keys():
        soup = BeautifulSoup(urllib.request.urlopen(s), 'lxml')
        items.extend([g for g in [sources[s](i) for i in soup.findAll("item")] if g.gid not in checked_list])

    save_pending_items(items)

    return items

if __name__ == "__main__":
    steam.print_game_list(update_all())
