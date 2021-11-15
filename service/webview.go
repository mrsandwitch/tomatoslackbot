package service

import (
	"bushyang/tomatoslackbot/util"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"sort"
	"time"
)

type Webview struct {
	sender    *SenderService
	db        *DbService
	templates fs.FS
}

func InitWebviewService(sender *SenderService, db *DbService) *Webview {
	return &Webview{
		sender: sender,
		db:     db,
	}
}

type recordReadable struct {
	Start    string
	Duration string
	Tag      string
	Label    string
	Desc     string
}

type WebDataEntry struct {
	Count     int
	Title     string
	Readables []recordReadable
}

type WebData struct {
	Items []WebDataEntry
}

func toReadable(records []ClockRecord) []recordReadable {
	timeFormat := "15:04:05"

	readables := make([]recordReadable, len(records))
	for ix, record := range records {
		var label string
		if record.Tag == "work" {
			label = "primary"
		} else if record.Tag == "spare" {
			label = "success"
		}
		readables[ix] = recordReadable{
			Start:    record.Start.Format(timeFormat),
			Duration: fmt.Sprintf("%dm", record.Duration/time.Minute),
			Tag:      record.Tag,
			Desc:     record.Desc,
			Label:    label,
		}
	}

	return readables
}

type group struct {
	Id     int
	Title  string
	record []ClockRecord
}

// For sorting
type ByDate []group

func (a ByDate) Len() int           { return len(a) }
func (a ByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return a[i].Id > a[j].Id }

func webDataGen(records []ClockRecord) (*WebData, error) {
	webData := &WebData{}
	table := make(map[int][]ClockRecord)
	groupFormat := "2006-01-02"

	for _, r := range records {
		groupId := r.Start.Year()*1000 + r.Start.YearDay()
		table[groupId] = append(table[groupId], r)
	}

	var groups []group
	for key, value := range table {
		grp := group{
			Title:  value[0].Start.Format(groupFormat),
			Id:     key,
			record: value,
		}
		groups = append(groups, grp)
	}

	sort.Sort(ByDate(groups))

	for _, group := range groups {
		readables := toReadable(group.record)

		entry := WebDataEntry{
			Count:     len(group.record),
			Title:     group.Title,
			Readables: readables,
		}
		webData.Items = append(webData.Items, entry)
	}

	return webData, nil
}

func (web *Webview) WebUrlGet(w http.ResponseWriter, req *http.Request) {
	uri, err := util.GetDefaultUri()
	if err != nil {
		log.Fatal(err)
	}

	url := uri + "/web"
	text := fmt.Sprintf("View the record at:\n %s\n", url)

	_, err = web.sender.SendMsg(text)
	if err != nil {
		log.Println(err)
	}
}

func (web *Webview) TestPage(w http.ResponseWriter, req *http.Request) {
	records, err := web.db.ClockRecordGet()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	webData, err := webDataGen(records)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFS(web.templates, "templates/index.html")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	err = t.Execute(w, webData)
	if err != nil {
		log.Println(err)
		return
	}
}

func (web *Webview) RecordGet(w http.ResponseWriter, req *http.Request) {
	records, err := web.db.ClockRecordGet()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	webData, err := webDataGen(records)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(webData)
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
