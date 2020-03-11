package service

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Clock struct {
	sender *Sender
	db     *DbService
}

func InitClockService(inHookUrl string, db *DbService) *Clock {
	sender := &Sender{
		IncommingHookUrl: inHookUrl,
	}

	return &Clock{
		sender: sender,
		db:     db,
	}
}

func (clock *Clock) Destroy() {
	clock.db.Close()
}

func (clock *Clock) Start(record *ClockRecord) error {
	duration := 25 * time.Minute
	t := time.Now()
	timeFormat := "2006-01-02 15:04:05"
	text := fmt.Sprintf("Tomato clock start on [%s]", t.Format(timeFormat))

	record.Start = t
	record.Duration = duration

	_, err := clock.sender.SendMsg(text)
	if err != nil {
		log.Println(err)
		return err
	}

	go func() {
		time.Sleep(duration)
		t := time.Now()

		text := fmt.Sprintf("Tomato clock finished on [%s]. Elapse[%s]", t.Format(timeFormat), duration.String())
		if _, err := clock.sender.SendMsg(text); err != nil {
			log.Fatal(err)
		}

		// Record into database
		err := clock.db.ClockRecordAdd(*record)
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(2 * time.Second)

		if _, err := clock.sender.SendMsg("Please take a rest"); err != nil {
			log.Fatal(err)
		}
	}()

	return nil
}

func (clock *Clock) TomatoClockStart(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		log.Fatal(err)
	}

	text := req.PostForm.Get("text")
	splits := strings.Split(text, " ")
	var tag, desc string

	if len(splits) > 0 {
		if splits[0] == "w" || splits[0] == "work" {
			tag = "work"
		} else if splits[0] == "s" || splits[0] == "spare" {
			tag = "spare"
		}
		desc = strings.Join(splits[1:], " ")
	}

	record := &ClockRecord{
		Tag:  tag,
		Desc: desc,
	}

	err := clock.Start(record)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
