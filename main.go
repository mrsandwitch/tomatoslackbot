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

//go:embed web_app/templates/*
var templates embed.FS

//go:embed web_app/assets/*
var assets embed.FS

//go:embed web_app/src/*
var webSrc embed.FS

//go:embed dist/*
var webDist embed.FS

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

	assetRootFs, err := fs.Sub(assets, "web_app")
	if err != nil {
		log.Fatal(err)
	}
	templateRootFs, err := fs.Sub(templates, "web_app/templates")
	if err != nil {
		log.Fatal(err)
	}
	webSrcRootFs, err := fs.Sub(webSrc, "web_app")
	if err != nil {
		log.Fatal(err)
	}
	webDistRootFs, err := fs.Sub(webDist, "dist")
	if err != nil {
		log.Fatal(err)
	}

	confService := service.InitConfigService()
	senderService := service.InitSenderService(confService)
	clockService := service.InitClockService(senderService, dbService, confService)
	webService := service.InitWebviewService(senderService, dbService, templateRootFs)

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
			//r.Get("/record", webService.RecordPage)
			//r.Get("/clock", webService.ClockPage)
			//r.Get("/test", webService.TestPage)
			r.Handle("/assets/*", http.FileServer(http.FS(assetRootFs)))
			r.Handle("/src/*", http.FileServer(http.FS(webSrcRootFs)))
			//r.Handle("/", http.FileServer(http.FS(templateRootFs)))
			r.Handle("/*", http.FileServer(http.FS(webDistRootFs)))
			r.Get("/api/records", webService.RecordGet)
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
