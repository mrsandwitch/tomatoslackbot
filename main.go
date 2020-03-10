package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
	"time"
)

var ip = flag.String("ip", "0.0.0.0", "server ip")
var rootDir = flag.String("root_dir", "/root/workspace", "root directory")

var port = flag.Int("port", 8000, "server port")

type message struct {
	Text string `json:"text"`
}

func TomatoClockStart(w http.ResponseWriter, req *http.Request) {
	t := time.Now()
	timeFormat := "2006-01-02 15:04:05"

	text := fmt.Sprintf("Tomato clock start on [%s]", t.Format(timeFormat))
	_, err := SendMsg(text)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		t := time.Now()
		timeFormat := "2006-01-02 15:04:05"

		//time.Sleep(5 * time.Second)
		time.Sleep(25 * time.Minute)

		text := fmt.Sprintf("Tomato clock finished on [%s]", t.Format(timeFormat))
		if _, err := SendMsg(text); err != nil {
			log.Fatal(err)
		}

		time.Sleep(2 * time.Second)

		if _, err := SendMsg("Please take a rest"); err != nil {
			log.Fatal(err)
		}
	}()

	w.WriteHeader(http.StatusOK)
}

func SendMsg(text string) (resp *http.Response, err error) {
	url := "https://hooks.slack.com/services/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"

	msg := message{
		Text: text,
	}

	js, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	body := bytes.NewBuffer(js)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	return resp, err
}

func main() {
	flag.Parse()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Group(func(r chi.Router) {
		{
			r.Post("/tomato", TomatoClockStart)
		}
	})

	log.Println("Server start running")

	url := fmt.Sprintf("%s:%d", *ip, *port)
	server := &http.Server{Addr: url, Handler: r}

	err := server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
