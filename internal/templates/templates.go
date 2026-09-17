package templates

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/dougdomingos/test-prettify/internal/model"
)

//go:embed base.html layout/*.html pages/*.html styles.css
var files embed.FS

// templateData is the envelope passed to every template execution.
// CSS is at the top level to match the {{ .CSS }} reference in base.html.
// Data carries page-specific payload.
type templateData struct {
	CSS  template.CSS
	Data any
}

// PageRenderer parses and executes HTML templates with embedded assets.
type PageRenderer struct {
	css template.CSS
}

// NewPageRenderer creates a Renderer by loading the embedded CSS asset.
func NewPageRenderer() (*PageRenderer, error) {
	raw, err := files.ReadFile("styles.css")
	if err != nil {
		return nil, err
	}

	return &PageRenderer{css: template.CSS(raw)}, nil
}

// splitLines splits output into individual rows, dropping the trailing empty
// element produced by a final newline while keeping interior blank lines.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}

	lines := strings.Split(s, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

// ListPages returns the names of all available page templates, without extension.
func ListPages() ([]string, error) {
	entries, err := fs.ReadDir(files, "pages")
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".html" {
			names = append(names, strings.TrimSuffix(entry.Name(), ".html"))
		}
	}

	return names, nil
}

// statusClass maps a source line status to its CSS modifier class.
func statusClass(status model.LineStatus) string {
	switch status {
	case model.LineCovered:
		return "covered"
	case model.LineUncovered:
		return "uncovered"
	case model.LineMixed:
		return "mixed"
	default:
		return "neutral"
	}
}

// gradeClass maps a coverage percentage to its indicator color class.
func gradeClass(percent float64) string {
	return model.CoverageGrade(percent)
}

// fileID derives a stable, URL-safe anchor id from a module-relative file path
// so overview links can jump to the matching coverage file card.
func fileID(path string) string {
	var b strings.Builder
	b.WriteString("file-")
	for _, r := range path {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteByte('-')
		}
	}
	return b.String()
}

// Render executes the named page template and writes the output to w.
// pageName must match a file under pages/ without the .html extension.
func (r *PageRenderer) Render(w io.Writer, pageName string, data any) error {
	tmpl, err := template.New(pageName).Funcs(template.FuncMap{
		"splitLines":  splitLines,
		"statusClass": statusClass,
		"gradeClass":  gradeClass,
		"fileID":      fileID,
	}).ParseFS(files,
		"base.html",
		"layout/header.html",
		"layout/sidebar.html",
		"pages/"+pageName+".html",
	)

	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(w, "base.html", templateData{
		CSS:  r.css,
		Data: data,
	})
}
