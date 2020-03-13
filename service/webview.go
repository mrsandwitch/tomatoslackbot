package service

import (
	"bushyang/tomatoslackbot/util"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
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
			Duration: record.Duration.String(),
			Tag:      record.Tag,
			Desc:     record.Desc,
		}
	}

	return readables
}

func webDataGen(records []ClockRecord) (*WebData, error) {
	webData := &WebData{}
	table := make(map[string][]recordReadable)
	readables := toReadable(records)

	for _, r := range readables {
		splits := strings.Split(r.Start, " ")
		if len(splits) < 2 {
			log.Fatal("Faile to split start time string")
		}

		key := splits[0]
		r.Start = splits[1]

		table[key] = append(table[key], r)
	}

	for key, value := range table {
		entry := WebDataEntry{
			Count:     len(value),
			Title:     key,
			Readables: value,
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
