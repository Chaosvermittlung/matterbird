package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
)

type config struct {
	Mattermosttoken   string
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
	Rate              int
	LastCheck         time.Time
	Nicklist          string
}

func (c *config) load(f string) {

	file, err := ioutil.ReadFile(f)
	if err != nil {
		log.Fatal("Could not load: ", err)
	}
	err = json.Unmarshal(file, c)
	if err != nil {
		log.Fatal("Could not parse: ", err)
	}

	file, err = ioutil.ReadFile(conf.Nicklist)
	if err != nil {
		log.Fatal("Could not load: ", err)
	}
	err = json.Unmarshal(file, &Nicklist)
	if err != nil {
		log.Fatal("Could not parse: ", err)
	}
}

func (c *config) save(f string) {
	b, err := json.MarshalIndent(c, "", "	")
	if err != nil {
		log.Fatal("Could not marshal:", err)
	}
	err = ioutil.WriteFile(f, b, 664)
	if err != nil {
		log.Fatal("Could not save:", err)
	}
}
