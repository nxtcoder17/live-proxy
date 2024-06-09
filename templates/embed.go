package templates

import (
	"embed"
	"html/template"
	"io"
)

//go:embed *
var templatesDir embed.FS

type Page string

const (
	HomePage Page = "index.html"
)

type HomePageArgs struct {
	Title         string
	WebsocketPath string
}

type Template struct {
	*template.Template
}

func NewTemplate() (*Template, error) {
	t := template.New("htmx")
	// t = t.Funcs(sprig.TxtFuncMap())
	if _, err := t.ParseFS(templatesDir, "htmx/*.html"); err != nil {
		return nil, err
	}

	return &Template{Template: t}, nil
}

func (t *Template) Render(wr io.Writer, name Page, data any) error {
	return t.ExecuteTemplate(wr, string(name), data)
}
