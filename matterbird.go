package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/ChimeraCoder/anaconda"
	"github.com/gorilla/mux"
)

var conf config
var configFile string
var tomatter = make(chan string)
var Nicklist = make(map[string]string)

type mattermessage struct {
	Text string `json:"name"`
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
	u, tmpCred, err := anaconda.AuthorizationURL("oob")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(u)

	var pin string
	fmt.Scanln(&pin)

	cred, _, err := anaconda.GetCredentials(tmpCred, pin)
	if err != nil {
		log.Fatal(err)
	}

	api := anaconda.NewTwitterApi(cred.Token, cred.Secret)
	api.SetLogger(anaconda.BasicLogger)

	res, err := api.GetTweet(634839419317456896, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res)

	r := mux.NewRouter()
	r = r.StrictSlash(true)
	r.HandleFunc("/twitter", twitterhandler)

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
	if (utf8.RuneCountInString(tweet) > 140) || (trigger == "!tweet") {

		return
	}

	switch trigger {
	case "!tweet":
		err = sendtweet(tweet)
	case "!dm":
		err = senddm(tweet)
	}

}

func sendtweet(tweet string) error {
	return nil
}

func senddm(tweet string) error {
	return nil
}
