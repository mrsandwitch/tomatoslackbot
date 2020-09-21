package main

import (
	"bushyang/tomatoslackbot/service"
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
)

var inHookUrl = flag.String("in_hook_url", "", "Incomming hool url")
var ip = flag.String("ip", "0.0.0.0", "server ip")
var rootDir = flag.String("root_dir", "/root/workspace", "root directory")

var port = flag.Int("port", 8000, "server port")

func main() {
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	dbService, err := service.InitDbService()
	if err != nil {
		log.Fatal(err)
	}
	defer dbService.Close()

	confService := service.InitConfigService(*inHookUrl)
	senderService := service.InitSenderService(confService)
	clockService := service.InitClockService(senderService, dbService, confService)
	webService := service.InitWebviewService(senderService, dbService)

	//-- Test function
	//senderService.SendMsg("hello2")
	//clockService.Start()

	r.Group(func(r chi.Router) {
		{
			r.Post("/tomato", clockService.TomatoClockStart)
			r.Post("/weburl", webService.WebUrlGet)
			r.Post("/setting", confService.Setting)
			r.Get("/record", webService.RecordPage)
			r.Get("/clock", webService.ClockPage)
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
