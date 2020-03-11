package main

import (
	"bushyang/tomatoslackbot/service"
	"bushyang/tomatoslackbot/util"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var inHookUrl = flag.String("in_hook_url", "", "Incomming hool url")
var ip = flag.String("ip", "0.0.0.0", "server ip")
var rootDir = flag.String("root_dir", "/root/workspace", "root directory")

var port = flag.Int("port", 8000, "server port")

type config struct {
	IncommingHookUrl string `json:"incomming_hook_url"`
}

func (conf *config) save() error {
	dir, err := filepath.Abs(util.GetDataDir())
	if err != nil {
		log.Println(err)
		return err
	}

	os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(util.GetConfigPath(), data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (conf *config) read() error {
	js, err := ioutil.ReadFile(util.GetConfigPath())
	if err != nil {
		return err
	}

	err = json.Unmarshal(js, conf)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	// Save and load config
	conf := config{}
	if *inHookUrl == "" {
		if err := conf.read(); err != nil {
			log.Fatal(err)
		}
		*inHookUrl = conf.IncommingHookUrl
	} else {
		conf.IncommingHookUrl = *inHookUrl
		if err := conf.save(); err != nil {
			log.Fatal(err)
		}
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	db, err := service.InitDbService()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	clockService := service.InitClockService(*inHookUrl, db)
	webService := service.InitWebviewService(db)

	//-- Test function
	//sender := service.Sender{IncommingHookUrl: *inHookUrl}
	//if _, err := sender.SendMsg("hello2"); err != nil {
	//	log.Fatal(err)
	//}
	//clockService.Start()

	r.Group(func(r chi.Router) {
		{
			r.Post("/tomato", clockService.TomatoClockStart)
			r.Get("/view", webService.WebShow)
		}
	})

	log.Println("Server start running")

	url := fmt.Sprintf("%s:%d", *ip, *port)
	server := &http.Server{Addr: url, Handler: r}

	err = server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
