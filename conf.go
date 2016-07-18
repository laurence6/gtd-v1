package main

import (
	"encoding/json"
	"os"

	"gopkg.in/pg.v4"
	"gopkg.in/redis.v4"
)

type Conf struct {
	RedisOptions redis.Options `json:"redis_options"`

	PgOptions pg.Options `json:"pg_options"`

	WebListenAddr string `json:"web_listen_addr"`
}

func parseConfFile(path string) (*Conf, error) {
	confFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(confFile)
	conf := &Conf{}

	err = dec.Decode(conf)
	if err != nil {
		return nil, err
	}

	confFile.Close()

	return conf, nil
}
