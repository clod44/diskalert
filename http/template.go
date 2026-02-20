package http

import (
	"diskalert/diskmon"
	"encoding/json"
	"html/template"
	"log"
)


var tpl *template.Template 

type HomeData struct {
	DiskStatus diskmon.DiskStatus
}

func init() {
	var err error
	funcMap := template.FuncMap{
		"marshalJSON": marshalJSON,
	}

	tpl, err = template.New("index.html").Funcs(funcMap).ParseFS(embed.HttpTemplates, "templates/index.html")
	if err != nil {
		log.Println("[ERROR] Template parsing failed: %v", err)
	}
	log.Println("Template successfully compiled and ready.")
}

func InitializeTemplates() {
	var err error
	funcMap := template.FuncMap{
		"marshalJSON": marshalJSON,
	}

	tpl, err = template.New("index.html").Funcs(funcMap).ParseFS(embed.HttpTemplates, "templates/index.html")
	if err != nil {
		log.Println("[ERROR] Template parsing failed: %v", err)
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