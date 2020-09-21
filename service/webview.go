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

func getRecordPageTpl() string {
	return `
<!DOCTYPE html>
<html>
	<head>
		<title>Tomato Clock Record</title>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">

		<!-- UIkit CSS -->
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/uikit@3.3.6/dist/css/uikit.min.css" />

		<!-- UIkit JS -->
		<script src="https://cdn.jsdelivr.net/npm/uikit@3.3.6/dist/js/uikit.min.js"></script>
		<script src="https://cdn.jsdelivr.net/npm/uikit@3.3.6/dist/js/uikit-icons.min.js"></script>
	</head>
	<body>
		<div class="uk-child-width-1-3@m uk-child-width-1-2@s" uk-grid="masonry: true">
			{{ range .Items }}
			<div>
				<div class="uk-card uk-card-default uk-card-body">
		    	    <h3 class="uk-card-title">
						<div class="uk-card-badge uk-label">{{.Count}}</div>
						{{.Title}} 
					</h3>
					<table class="uk-table uk-table-small uk-text-nowrap">
						<tbody>
							{{range .Readables}}
							<tr>
								<td class="uk-width-1-4">{{.Start}}</td>
								<td class="uk-width-1-4">{{.Duration}}</td>
								<td class="uk-text-{{.Label}}">{{.Tag}}</td>
							</tr>
							{{ end }}
						</tbody>
					</table>
				</div>
			</div>
			{{ end }}
		</div>
	</body>
</html>
`
}

func getClockPageTpl() string {
	return `
<!DOCTYPE html>
<html>
	<head>
		<title>Tomato Clock Record</title>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">

		<!-- UIkit CSS -->
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/uikit@3.3.6/dist/css/uikit.min.css" />

		<!-- UIkit JS -->
		<script src="https://cdn.jsdelivr.net/npm/uikit@3.3.6/dist/js/uikit.min.js"></script>
		<script src="https://cdn.jsdelivr.net/npm/uikit@3.3.6/dist/js/uikit-icons.min.js"></script>
	</head>
	<body>
		<iframe name="dummyframe" id="dummyframe" style="display: none;"></iframe>

		<form action="/tomato" method="post" target="dummyframe">
			<div style="padding:10px;">
				<button class="uk-button uk-button-default" name="ctlStr" value="w 10m">10 (work)</button>
				<button class="uk-button uk-button-primary" name="ctlStr" value="s 10m" style="margin-left:10px;">10 (spare)</button>
			</div>
			<div style="padding:10px;">
				<button class="uk-button uk-button-default" name="ctlStr" value="w 25m">25 (work)</button>
				<button class="uk-button uk-button-primary" name="ctlStr" value="s 25m" style="margin-left:10px;">25 (spare)</button>
			</div>
		</form>
	</body>
</html>
`
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

func (web *Webview) RecordPage(w http.ResponseWriter, req *http.Request) {
	records, err := web.db.ClockRecordGet()
	if err != nil {
		log.Fatal(err)
	}

	webData, err := webDataGen(records)
	if err != nil {
		log.Fatal(err)
	}

	tpl := getRecordPageTpl()

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(w, webData)
	if err != nil {
		log.Fatal(err)
	}
}

func (web *Webview) ClockPage(w http.ResponseWriter, req *http.Request) {
	tpl := getClockPageTpl()

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(w, nil)
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
