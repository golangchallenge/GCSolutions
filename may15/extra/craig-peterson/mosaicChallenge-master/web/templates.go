package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
)

var t *template.Template

var devMode = flag.Bool("dev", false, "dev mode will load templates on each request, so that you can edit templates without restarting app")

func loadTemplates() (templates *template.Template, err error) {
	if !*devMode && t != nil {
		return t, nil
	}
	t, err = template.ParseGlob("templates/*")
	return t, err
}

type ErrNoTemplate struct {
	name string
}

func (e ErrNoTemplate) Error() string {
	return fmt.Sprintf("template not found: %s", e.name)
}

// render given template to w with header before and footer after.
func renderTemplate(w io.Writer, name string, data interface{}) (e error) {
	defer func() {
		if e != nil {
			log.Println(e)
		}
	}()
	t, err := loadTemplates()
	if err != nil {
		return err
	}
	tmpl := t.Lookup("header")
	if tmpl == nil {
		return ErrNoTemplate{"header"}
	}
	if err := tmpl.Execute(w, data); err != nil {
		return err
	}

	tmpl = t.Lookup(name)
	if tmpl == nil {
		return ErrNoTemplate{name}
	}
	if err := tmpl.Execute(w, data); err != nil {
		return err
	}

	tmpl = t.Lookup("footer")
	if tmpl == nil {
		return ErrNoTemplate{"footer"}
	}
	if err := tmpl.Execute(w, data); err != nil {
		return err
	}

	return nil
}
