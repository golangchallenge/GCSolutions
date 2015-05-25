// Package app : temnplates provides support for using HTML
// based templates for responses.
package app

import (
	"bytes"
	"github.com/aishraj/gopherlisa/common"
	"html/template"
)

// executeTemplate executes the specified template with the specified variables.
func executeTemplate(context *common.AppContext, name string, params map[string]interface{}) []byte {
	context.Log.Println("Executing template named", name)
	t, err := template.ParseFiles("views/" + name + ".html")
	if err != nil {
		return nil
	}
	markup := new(bytes.Buffer)
	err = t.Execute(markup, params)
	if err != nil {
		context.Log.Panic(err)
		return nil
	}
	return markup.Bytes()
}
