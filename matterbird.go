package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/gorilla/mux"
)

var conf config
var configFile string
var tomatter = make(chan string)
var Nicklist = make(map[string]string)
var api *anaconda.TwitterApi

type mattermessage struct {
	Text string `json:"text"`
}

func init() {
	flag.StringVar(&configFile, "config", "./config.json", "Sets the file location for config.json")
	flag.StringVar(&configFile, "c", "./config.json", "Sets the file location for config.json")
}

func main() {
	flag.Parse()
	conf.load(configFile)

	anaconda.SetConsumerKey(conf.ConsumerKey)
	anaconda.SetConsumerSecret(conf.ConsumerSecret)

	var cred *oauth.Credentials
	if conf.AccessToken == "" {
		u, tmpCred, err := anaconda.AuthorizationURL("oob")
		if err != nil {
			log.Fatal(err)
		}
		log.Println(u)

		var pin string
		fmt.Scanln(&pin)

		cred, _, err = anaconda.GetCredentials(tmpCred, pin)
		if err != nil {
			log.Fatal(err)
		}

		conf.AccessToken = cred.Token
		conf.AccessTokenSecret = cred.Secret
		conf.save(configFile)
	}

	api = anaconda.NewTwitterApi(conf.AccessToken, conf.AccessTokenSecret)
	api.SetLogger(anaconda.BasicLogger)

	r := mux.NewRouter()
	r = r.StrictSlash(true)
	r.HandleFunc("/twitter", twitterhandler)
	log.Println("Starting Server")
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", r))

}

func bodytomap(s string) map[string]string {
	res := make(map[string]string)
	split := strings.Split(s, "&")
	for _, sp := range split {
		spl := strings.Split(sp, "=")
		if len(spl) == 2 {
			res[spl[0]] = spl[1]
		}
	}
	return res
}

func getshortnick(nick string) string {
	res := ""
	s := Nicklist[nick]
	if s != "" {
		res = " ^" + s
	}
	return res
}

func twitterhandler(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	v := bodytomap(buf.String())
	token := v["token"]
	if token != conf.Mattermosttoken {
		log.Println("Mattermost Token wrong")
		log.Println(token)
		log.Println(conf.Mattermosttoken)

		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	nick := v["user_name"]
	trigger, err := url.QueryUnescape(v["trigger_word"])
	if err != nil {
		log.Println("Error Unescaping:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	text := v["text"]
	text, err = url.QueryUnescape(text)
	if err != nil {
		log.Println("Error Unescaping:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	text = strings.TrimPrefix(text, "!tweet ")
	text = strings.TrimPrefix(text, "!dm ")
	short := getshortnick(nick)
	tweet := text + short

	var m mattermessage
	if (utf8.RuneCountInString(tweet) > 140) && (trigger == "!tweet") {
		log.Println("Tweet too long", utf8.RuneCountInString(tweet))
		log.Println(tweet)
		m.Text = "Tweet longer than 140 letters"
		b, err := json.MarshalIndent(&m, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		return
	}

	switch trigger {
	case "!tweet":
		err = sendtweet(tweet)
	case "!dm":
		err = senddm(tweet)
	}

	if err != nil {
		m.Text = "Error: " + err.Error()
	}
	m.Text = "Tweet successfully send"
	b, err := json.MarshalIndent(&m, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	return

}

func sendtweet(tweet string) error {
	log.Println("Tweet", tweet)
	_, err := api.PostTweet(tweet, nil)
	return err
}

func senddm(tweet string) error {
	return nil
}
