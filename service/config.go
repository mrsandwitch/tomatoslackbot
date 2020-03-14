package service

import (
	"bushyang/tomatoslackbot/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ConfigService struct {
	conf Config
}

type Config struct {
	IncommingHookUrl string `json:"incomming_hook_url"`
	Duration         string `json:"duration"`
}

func InitConfigService(incommingHookUrl string) *ConfigService {
	service := &ConfigService{}

	// Save and load config
	if err := service.read(); err != nil {
		log.Println("Please set incomming hook url")
		log.Fatal(err)
	}

	if incommingHookUrl != "" {
		service.conf.IncommingHookUrl = incommingHookUrl
		if err := service.save(); err != nil {
			log.Fatal(err)
		}
	}

	return service
}

func (service *ConfigService) InHoolUrlGet() string {
	return service.conf.IncommingHookUrl
}

func (service *ConfigService) DurationGet() time.Duration {
	dur, err := time.ParseDuration(service.conf.Duration)
	if err != nil {
		dur = 25 * time.Minute
	}
	return dur
}

func (service *ConfigService) save() error {
	dir, err := filepath.Abs(util.GetDataDir())
	if err != nil {
		log.Println(err)
		return err
	}

	os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(service.conf, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(util.GetConfigPath(), data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (service *ConfigService) read() error {
	js, err := ioutil.ReadFile(util.GetConfigPath())
	if err != nil {
		return err
	}

	err = json.Unmarshal(js, &service.conf)
	if err != nil {
		return err
	}

	return nil
}

func (service *ConfigService) Setting(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		log.Fatal(err)
	}

	text := req.PostForm.Get("text")
	log.Println(text)
	splits := strings.Split(text, " ")
	var attr, dur string

	if len(splits) >= 2 {
		attr = splits[0]
		dur = splits[1]
		if attr == "dur" || attr == "duration" {
			_, err := time.ParseDuration(dur)
			if err != nil {
				log.Println(err)
				return
			}
			service.conf.Duration = dur

			if err := service.save(); err != nil {
				log.Println(err)
				return
			}
		}
	}

	w.Write([]byte(fmt.Sprintf("Duration: %s\n", service.conf.Duration)))

	w.WriteHeader(http.StatusOK)
}
