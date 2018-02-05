package games_cache

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
)

// TODO: handle list concurrency

type Game struct {
	Name string `json:"name"`
	Gid string `json:"gid"`
	Link string `json:"link"`
	Genre string `json:"genre"`
}

func (g Game) String() string {
	return fmt.Sprintf("Title: %s (%s)\n\t%s\n", g.Name, g.Genre, g.Link)
}

func GetList(fname string) (*[]Game, error) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	var list []Game
	err = json.Unmarshal(data, &list)
	if err != nil {
		return nil, err
	}

	return &list, nil
}

func StoreList(fname string, list *[]Game) error {
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fname, data, 0600)
}

