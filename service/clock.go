package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type clockReq struct {
	CtlStr string `json:"ctlStr"`
}

type runningClock struct {
	Id        uint64
	StartTime time.Time
	Duration  time.Duration
	Tag       string
	Desc      string
}

type ClockService struct {
	sender        *SenderService
	db            *DbService
	conf          *ConfigService
	runningClocks sync.Map
	isChecking    uint32
	clockId       uint64
}

const timeFormat = "2006-01-02 15:04:05"

func InitClockService(sender *SenderService, db *DbService, conf *ConfigService) (service *ClockService) {
	service = &ClockService{
		sender: sender,
		db:     db,
		conf:   conf,
	}
	return
}

func (service *ClockService) checkExpire() {
	swap := atomic.CompareAndSwapUint32(&service.isChecking, 0, 1)
	if !swap {
		return
	}

	minDur := time.Duration(0)
	currTime := time.Now()
	var remain int

	service.runningClocks.Range(func(key interface{}, val interface{}) bool {
		clock, ok := val.(runningClock)
		if !ok {
			return true
		}
		if clock.StartTime.Add(clock.Duration).Before(currTime) {
			service.runningClocks.Delete(clock.Id)

			err := service.finishNotification(clock)
			if err != nil {
				log.Println(err)
				return true
			}
		} else {
			remain += 1
			expireTime := clock.StartTime.Add(clock.Duration).Sub(currTime)
			if expireTime > minDur {
				minDur = expireTime
			}
		}

		return true
	})

	atomic.StoreUint32(&service.isChecking, 0)
	if remain > 0 {
		time.AfterFunc(minDur, func() {
			service.checkExpire()
		})
	}
}

func (service *ClockService) Destroy() {
	service.db.Close()
}

func (service *ClockService) finishNotification(clock runningClock) (err error) {
	t := time.Now()

	text := fmt.Sprintf("Tomato clock finished on [%s]. Elapse[%s]", t.Format(timeFormat), clock.Duration.String())
	log.Println(text)
	if _, err := service.sender.SendMsg(text); err != nil {
		log.Println(err)
	}

	record := ClockRecord{
		Start:    clock.StartTime,
		Duration: clock.Duration,
		Tag:      clock.Tag,
		Desc:     clock.Desc,
	}

	// Record into database
	err = service.db.ClockRecordAdd(record)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(3 * time.Second)
	if _, err := service.sender.SendMsg("Please take a rest"); err != nil {
		log.Println(err)
	}

	time.Sleep(5 * time.Minute)
	if _, err := service.sender.SendMsg("Resting is finished"); err != nil {
		log.Println(err)
	}
	return
}

func (service *ClockService) AddClock(record *ClockRecord, duration time.Duration) (err error) {
	t := time.Now()
	timeFormat := "2006-01-02 15:04:05"
	text := fmt.Sprintf("Tomato clock start on [%s]. Duration:[%s]", t.Format(timeFormat), duration.String())
	log.Println(text)

	record.Start = t
	record.Duration = duration

	_, err = service.sender.SendMsg(text)
	if err != nil {
		log.Println(err)
	}

	atomic.AddUint64(&service.clockId, 1)
	clock := runningClock{Id: service.clockId, StartTime: t, Duration: duration}
	service.runningClocks.Store(service.clockId, clock)
	service.checkExpire()

	return
}

func (service *ClockService) TomatoClockStart(w http.ResponseWriter, req *http.Request) {
	clockReq := clockReq{}
	err := json.NewDecoder(req.Body).Decode(&clockReq)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	splits := strings.Split(clockReq.CtlStr, " ")
	var tag, desc string

	if len(splits) > 0 {
		if splits[0] == "w" || splits[0] == "work" {
			tag = "work"
		} else if splits[0] == "s" || splits[0] == "spare" {
			tag = "spare"
		}
		desc = strings.Join(splits[1:], " ")
	}

	// Check if custom duration is included in the command text
	provideDur := false
	var duration time.Duration
	for _, split := range splits {
		dur, err := time.ParseDuration(split)
		if err != nil {
			continue
		}
		provideDur = true
		duration = dur
	}

	if !provideDur {
		duration = service.conf.DurationGet()
	}

	record := &ClockRecord{
		Tag:  tag,
		Desc: desc,
	}

	err = service.AddClock(record, duration)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (service *ClockService) RunningClockGet(w http.ResponseWriter, req *http.Request) {
	var clocks []runningClock

	service.runningClocks.Range(func(key interface{}, val interface{}) bool {
		clock, ok := val.(runningClock)
		if !ok {
			return true
		}
		clocks = append(clocks, clock)

		return true
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(clocks)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonData)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
