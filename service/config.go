package service

import (
	"bushyang/tomatoslackbot/util"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type ConfigService struct {
	conf Config
}

type Config struct {
	IncommingHookUrl string `json:"incomming_hook_url"`
}

func InitConfigService(incommingHookUrl string) *ConfigService {
	service := &ConfigService{}

	// Save and load config
	if err := service.read(); err != nil {
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

func (service *ConfigService) GetInHoolUrl() string {
	return service.conf.IncommingHookUrl
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
