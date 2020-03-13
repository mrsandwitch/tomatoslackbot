package service

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type SenderService struct {
	IncommingHookUrl string
}

type message struct {
	Text string `json:"text"`
}

func InitSenderService(incommingHookUrl string) *SenderService {
	service := &SenderService{
		IncommingHookUrl: incommingHookUrl,
	}
	return service
}

func (s *SenderService) SendMsg(text string) (resp *http.Response, err error) {
	url := s.IncommingHookUrl

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
