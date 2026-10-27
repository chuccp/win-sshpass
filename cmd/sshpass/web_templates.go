package main

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
)

//go:embed templates/index.html static/*
var webTemplateFiles embed.FS

var webAssetFS = mustWebAssetFS()

type webTemplates struct {
	index *template.Template
}

func mustLoadWebTemplates() *webTemplates {
	templateFile, err := webTemplateFiles.ReadFile("templates/index.html")
	if err != nil {
		panic(err)
	}
	return &webTemplates{
		index: template.Must(template.New("index.html").Parse(string(templateFile))),
	}
}

func (t *webTemplates) execute(w io.Writer, data webIndexData) error {
	return t.index.Execute(w, data)
}

func mustWebAssetFS() fs.FS {
	assets, err := fs.Sub(webTemplateFiles, "static")
	if err != nil {
		panic(err)
	}
	return assets
}
