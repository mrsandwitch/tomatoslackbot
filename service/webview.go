package service

import (
	"bushyang/tomatoslackbot/util"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"
)

type Webview struct {
	sender *SenderService
	db     *DbService
}

func InitWebviewService(sender *SenderService, db *DbService) *Webview {
	return &Webview{
		sender: sender,
		db:     db,
	}
}

func getTpl() string {
	return `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Tomato Clock Record</title>
	</head>
	<body>
		{{ range .Items }}
			<h2>{{.Title}} ({{.Count}})</h2>
			{{range .Readables}}
				<div>
					<span>{{ .Start }}</span>
					<span>{{ .Duration }}</span>
					<span>{{ .Tag }}</span>
				</div>
			{{ end }}
		{{ end }}
	</body>
</html>
`
}

type recordReadable struct {
	Start    string
	Duration string
	Tag      string
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
	timeFormat := "2006-01-02 15:04:05"

	readables := make([]recordReadable, len(records))
	for ix, record := range records {
		readables[ix] = recordReadable{
			Start:    record.Start.Format(timeFormat),
			Duration: fmt.Sprintf("%dm", record.Duration/time.Minute),
			Tag:      record.Tag,
			Desc:     record.Desc,
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

func (web *Webview) WebShow(w http.ResponseWriter, req *http.Request) {
	records, err := web.db.ClockRecordGet()
	if err != nil {
		log.Fatal(err)
	}

	webData, err := webDataGen(records)
	if err != nil {
		log.Fatal(err)
	}

	tpl := getTpl()

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(w, webData)
	if err != nil {
		log.Fatal(err)
	}
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
