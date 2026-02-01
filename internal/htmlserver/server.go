package htmlserver

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/scottmbaker/ai-rss-scraper/pkg/storage"
)

type Server struct {
	host string
	port int
	db   *storage.DB
}

func NewServer(host string, port int, db *storage.DB) *Server {
	return &Server{
		host: host,
		port: port,
		db:   db,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleList)
	mux.HandleFunc("/action", s.handleAction)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	log.Printf("Starting web server at http://%s", addr)
	return http.ListenAndServe(addr, mux)
}

const listTemplate = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>AI RSS Scraper - Articles</title>
	<style>
		body { font-family: sans-serif; margin: 2em; }
		table { width: 100%; border-collapse: collapse; }
		th, td { text-align: left; padding: 8px; border-bottom: 1px solid #ddd; }
		th { background-color: #f2f2f2; }
		.actions { margin-bottom: 1em; padding: 1em; background: #eee; border-radius: 4px; display: flex; justify-content: space-between; align-items: center; }
		button { padding: 0.5em 1em; cursor: pointer; margin-right: 0.5em; }
		.score-high { color: green; font-weight: bold; }
		.score-low { color: #888; }
		.filter { font-size: 0.9em; }
	</style>
	<script>
		function toggleAll(source) {
			checkboxes = document.getElementsByName('guids');
			for(var i=0, n=checkboxes.length;i<n;i++) {
				checkboxes[i].checked = source.checked;
			}
		}
		function updateFilter() {
			var reported = document.getElementById('reportedOnly').checked;
			window.location.href = "/?reported=" + reported;
		}
	</script>
</head>
<body>
	<h1>Articles</h1>
	<form action="/action" method="POST">
		<div class="actions">
			<div>
				<button type="submit" name="action" value="rescore">Rescore Selected</button>
				<button type="submit" name="action" value="reset-reported">Reset Reported Status</button>
			</div>
			<div class="filter">
				<input type="checkbox" id="reportedOnly" onclick="updateFilter()" {{if .ReportedOnly}}checked{{end}}>
				<label for="reportedOnly">Show Reported Only</label>
			</div>
		</div>
		<table>
			<thead>
				<tr>
					<th><input type="checkbox" onClick="toggleAll(this)"></th>
					<th>Score</th>
					<th>Title</th>
					<th>Date</th>
					<th>Reported</th>
				</tr>
			</thead>
			<tbody>
				{{range .Articles}}
				<tr>
					<td><input type="checkbox" name="guids" value="{{.GUID}}"></td>
					<td>
						{{if .Score}}
							<span class="{{if ge (toInt .Score) 50}}score-high{{else}}score-low{{end}}">{{.Score}}</span>
						{{else}}
							-
						{{end}}
					</td>
					<td><a href="{{.Link}}" target="_blank">{{.Title}}</a></td>
					<td>{{.PublishedDate.Format "2006-01-02 15:04"}}</td>
					<td>{{if .Reported}}Yes{{else}}No{{end}}</td>
				</tr>
				{{end}}
			</tbody>
		</table>
	</form>
</body>
</html>
`

type ListData struct {
	Articles     []storage.Article
	ReportedOnly bool
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	reportedOnly := r.URL.Query().Get("reported") == "true"

	// Fetch recent articles (limit 100 for now to keep it snappy)
	articles, err := s.db.ListArticlesFiltered(100, reportedOnly)
	if err != nil {
		http.Error(w, "Error fetching articles: "+err.Error(), http.StatusInternalServerError)
		return
	}

	funcMap := template.FuncMap{
		"toInt": func(s string) int {
			i, _ := strconv.Atoi(s)
			return i
		},
		"ge": func(a, b int) bool {
			return a >= b
		},
	}

	tmpl, err := template.New("list").Funcs(funcMap).Parse(listTemplate)
	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := ListData{
		Articles:     articles,
		ReportedOnly: reportedOnly,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	action := r.FormValue("action")
	guids := r.Form["guids"]

	if len(guids) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var err error
	switch action {
	case "rescore":
		err = s.db.ClearScores(guids)
	case "reset-reported":
		err = s.db.ResetReportedArticles(guids)
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Error performing action: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
