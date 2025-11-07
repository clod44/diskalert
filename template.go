package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
)

//go:embed templates/index.html
var content embed.FS

var tpl *template.Template 

type HomeData struct {
	DiskStatus DiskStatus
}

func init() {
	var err error
	funcMap := template.FuncMap{
        "marshalJSON": marshalJSON,
    }

	tpl, err = template.New("index.html").Funcs(funcMap).ParseFS(content, "templates/index.html")
	if err != nil {
		log.Fatalf("Template parsing failed: %v", err)
	}
	log.Println("Template successfully compiled and ready.")
}


func marshalJSON(data interface{}) (template.JS, error) {
    jsonBytes, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return "", err
    }
    return template.JS(jsonBytes), nil 
}
