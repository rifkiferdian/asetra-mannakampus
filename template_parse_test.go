package main

import (
	"html/template"
	"testing"
)

func TestApplicationTemplatesParse(t *testing.T) {
	_, err := template.New("app").Funcs(template.FuncMap{
		"no":      func(a, b int) int { return a + b },
		"baseURL": func(path string) string { return path },
	}).ParseGlob("templates/**/*")
	if err != nil {
		t.Fatalf("templates must parse: %v", err)
	}
}
