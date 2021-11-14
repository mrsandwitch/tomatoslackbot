package main

import (
	"bushyang/tomatoslackbot/service"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

var ip = flag.String("ip", "0.0.0.0", "server ip")
var rootDir = flag.String("root_dir", "/root/workspace", "root directory")

var port = flag.Int("port", 8000, "server port")

//go:embed dist/*
var webDist embed.FS

func main() {
	flag.Parse()

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	dbService, err := service.InitDbService()
	if err != nil {
		log.Fatal(err)
	}
	defer dbService.Close()

	webDistRootFs, err := fs.Sub(webDist, "dist")
	if err != nil {
		log.Fatal(err)
	}

	confService := service.InitConfigService()
	senderService := service.InitSenderService(confService)
	clockService := service.InitClockService(senderService, dbService, confService)
	webService := service.InitWebviewService(senderService, dbService)

	corsHandler := cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	r.Use(corsHandler)
	r.Group(func(r chi.Router) {
		{
			r.Post("/tomato", clockService.TomatoClockStart)
			r.Post("/weburl", webService.WebUrlGet)
			r.Post("/setting", confService.Setting)
			r.Handle("/*", http.FileServer(http.FS(webDistRootFs)))
			r.Get("/api/records", webService.RecordGet)
			r.Get("/api/clocks", clockService.RunningClockGet)
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
