package views

import (
	"embed"
	"fmt"
	"html/template"
	"io"
)

//go:embed layouts pages
var files embed.FS

var views = map[string]*template.Template{}

func init() {
	loadViews("badges", "player", "index", "not_ready")
}

func loadViews(names ...string) {
	for _, name := range names {
		file := fmt.Sprintf("pages/%s.tmpl", name)
		views[name] = template.Must(template.ParseFS(files, "layouts/main.tmpl", file))
	}
}

type Data struct {
	Refresh int
	Data    interface{}
}

func View(w io.Writer, name string, data Data) {
	view, ok := views[name]
	if !ok {
		panic("View does not exist : " + name)
	}
	if err := view.ExecuteTemplate(w, "main", data); err != nil {
		fmt.Printf("Failed to render view %s:%s\n", name, err)
	}
}
