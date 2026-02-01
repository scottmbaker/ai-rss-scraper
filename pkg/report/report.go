package report

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/scottmbaker/ai-rss-scraper/pkg/email"
	"github.com/scottmbaker/ai-rss-scraper/pkg/storage"
	"github.com/scottmbaker/ai-rss-scraper/pkg/utils"
)

// Report holds the data to be rendered in the template.
type Report struct {
	Title    string
	Articles []storage.Article
}

// NewReport creates a new Report instance.
func NewReport(title string, articles []storage.Article) *Report {
	return &Report{
		Title:    title,
		Articles: articles,
	}
}

// The template is hardcoded here as it eliminates the need to include a separate file.
// TODO: Make it configurable.
const reportTemplate = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>{{.Title}}</title>
	<style>
		body { font-family: sans-serif; max-width: 900px; margin: 2em auto; padding: 0 1em; background: #f4f4f4; color: #333; }
		h1 { text-align: center; color: #444; }
		.article { background: #fff; padding: 1.5em; margin-bottom: 1.5em; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.header { display: flex; justify-content: space-between; align-items: baseline; border-bottom: 2px solid #eee; padding-bottom: 0.5em; margin-bottom: 1em; }
		.title { font-size: 1.4em; font-weight: bold; }
		.title a { text-decoration: none; color: #2c3e50; }
		.title a:hover { color: #3498db; }
		.link { font-size: 0.85em; color: #3498db; margin-top: 0.2em; }
		.link a { text-decoration: none; color: #3498db; }
		.link a:hover { text-decoration: underline; }
		.meta { font-size: 0.85em; color: #888; text-align: right; }
		.score { font-weight: bold; color: #e67e22; font-size: 1.1em; }
		.analysis { font-style: italic; background: #f9f9f9; padding: 1em; border-left: 4px solid #3498db; margin: 1em 0; white-space: pre-wrap; }
		.description { line-height: 1.6; }
		.content { display: none; margin-top: 1em; padding-top: 1em; border-top: 1px dashed #ccc; font-size: 0.9em; color: #555; }
		.toggle-content { cursor: pointer; color: #3498db; font-size: 0.9em; user-select: none; }
		.toggle-content:hover { text-decoration: underline; }
	</style>
	<script>
		function toggleContent(id) {
			var el = document.getElementById('content-' + id);
			if (el.style.display === 'block') {
				el.style.display = 'none';
			} else {
				el.style.display = 'block';
			}
		}
	</script>
</head>
<body>
	<h1>{{.Title}}</h1>
	{{range .Articles}}
	<div class="article">
		<div class="header">
			<div>
				<div class="title"><a href="{{.Link}}" target="_blank">{{.Title}}</a></div>
				<div class="link"><a href="{{.Link}}" target="_blank">{{.Link}}</a></div>
			</div>
			<div class="meta">
				<span class="score">Score: {{.Score}}</span><br>
				{{.PublishedDate.Format "2006-01-02 15:04"}}<br>
				<span style="font-size:0.8em">{{.Model}}</span>
			</div>
		</div>
		{{if .Analysis}}
		<div class="analysis"><strong>Analysis:</strong><br>{{.Analysis}}</div>
		{{end}}
		<div class="description">{{.Description}}</div>
		{{if .Content}}
		<div class="toggle-content" onclick="toggleContent('{{.GUID}}')">Show/Hide Full Content</div>
		<div id="content-{{.GUID}}" class="content">{{.Content}}</div>
		{{end}}
	</div>
	{{else}}
	<p style="text-align:center">No articles found.</p>
	{{end}}
</body>
</html>
`

// GenerateHTML generates the HTML report string from the given articles.
func (r *Report) GenerateHTML() (string, error) {
	tmpl, err := template.New("report").Parse(reportTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing report template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r); err != nil {
		return "", fmt.Errorf("error executing report template: %w", err)
	}

	return buf.String(), nil
}

// GenerateFile creates an HTML report from the given articles and writes it to the specified file.
func (r *Report) GenerateFile(filename string) error {
	html, err := r.GenerateHTML()
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating report file %s: %w", filename, err)
	}
	defer utils.CloseFileBOF(f)

	if _, err := f.WriteString(html); err != nil {
		return fmt.Errorf("error writing report file: %w", err)
	}

	path, _ := filepath.Abs(filename)
	fmt.Printf("Report generated at: %s\n", path)
	return nil
}

// GenerateEmail sends the report via email.
func (r *Report) GenerateEmail(to, from, smarthost, identity, username, password string) error {
	if smarthost == "" || to == "" || from == "" {
		return fmt.Errorf("email configuration missing (smarthost, to, from)")
	}

	html, err := r.GenerateHTML()
	if err != nil {
		return fmt.Errorf("error generating html for email: %w", err)
	}

	if err := email.Send([]string{to}, from, r.Title, html, smarthost, identity, username, password); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
