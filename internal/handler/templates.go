package handler

import (
	"github.com/hugmouse/scan24/static"
	"github.com/hugmouse/scan24/templates"
	"html/template"
	"log"
)

func init() {
	funcs := template.FuncMap{
		"include": func(path string) template.HTML {
			bb, err := static.FS.ReadFile(path)
			if err != nil {
				panic(err)
			}

			return template.HTML(bb)
		},
		"includeJS": func(path string) template.JS {
			bb, err := static.FS.ReadFile(path)
			if err != nil {
				panic(err)
			}

			return template.JS(bb)
		},
	}

	baseTmpl := template.New("").Funcs(funcs)

	var err error

	tmplIndex, err = baseTmpl.New("index.gohtml").ParseFS(templates.FS, "index.gohtml")
	if err != nil {
		log.Fatalf("Failed to parse index.gohtml: %v", err)
	}

	tmplResult, err = baseTmpl.New("result.gohtml").ParseFS(templates.FS, "result.gohtml")
	if err != nil {
		log.Fatalf("Failed to parse result.gohtml: %v", err)
	}

	tmplProgress, err = baseTmpl.New("progress.gohtml").ParseFS(templates.FS, "progress.gohtml")
	if err != nil {
		log.Fatalf("Failed to parse progress.gohtml: %v", err)
	}

	tmplError, err = baseTmpl.New("error.gohtml").ParseFS(templates.FS, "error.gohtml")
	if err != nil {
		log.Fatalf("Failed to parse error.gohtml: %v", err)
	}
}
