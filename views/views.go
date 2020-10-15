package views

import (
	"fmt"
	"html/template"
	"io"
)

var views = map[string]*template.Template{}

func init() {
	loadViews("badges", "player", "index", "not_ready")
}

func loadViews(names ...string) {
	for _, name := range names {
		file := fmt.Sprintf("./views/pages/%s.tmpl", name)
		views[name] = template.Must(template.ParseFiles("./views/layouts/main.tmpl", file))
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
