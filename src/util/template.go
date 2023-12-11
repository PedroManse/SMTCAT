package util

import (
	"html/template"
	"net/http"
)

type HttpWriter = http.ResponseWriter
type HttpReq = *http.Request
type Plugin = func(w HttpWriter, r HttpReq, info map[string]any) (render bool, addinfo any)
type GOTMPlugin struct {
	Name string
	Plug Plugin
}

/* example

// make plugin
func number_plugin(w HttpWriter, r HttpReq, info map[string]any) any {
	return 4
}
// name plugin
var GOTM_number GOTMPlugin = {"number", number_plugin}

// create templated page
// pre-populate the request's info map
index := TemplatePage(
	"html/index.gohtml", map[string]any{"preinfo":"Hi!"},
	[]GOTMPlugin{GOTM_number}
)

// when index.ServeHTTP is called, number_plugin will be executed
// then it's return value (4) is set in the request's info map
// info map[string]any -> {"number": 4, "preinfo": "Hi!"}
// the info map will be sent to the template's file execution

*/

// serve file
type StaticFile struct {
	Filename string
}

// execute plugins, (possibly) template and serve gohtml file
type TemplatedPage struct {
	Template *template.Template
	Info map[string]any
	Plugins []GOTMPlugin
}

// execute plugins, execute custom function, (possibly) template and serve gohtml file
type LogicedPage struct {
	Template *template.Template
	Info map[string]any
	Plugins []GOTMPlugin
	Fn Plugin
}

func (s StaticFile) ServeHTTP (w HttpWriter, r HttpReq) {
	http.ServeFile(w, r, s.Filename)
}

func TemplatePage(filename string, info map[string]any, plugins []GOTMPlugin) TemplatedPage {
	if info == nil {
		info = make(map[string]any)
	}

	tmpl := template.Must(
		template.Must(
			template.ParseFiles(filename),
		).ParseGlob("templates/*.gohtml"),
	)

	return TemplatedPage{
		tmpl, info, plugins,
	}
}

func (s TemplatedPage) ServeHTTP (w HttpWriter, r HttpReq) {
	var render = true
	var prender bool
	for _, plug := range s.Plugins {
		prender, s.Info[plug.Name] = plug.Plug(w, r, s.Info)
		render = render&&prender
	}
	if (render) {
		s.Template.Execute(w, s.Info)
	}
}

func (s LogicedPage) ServeHTTP (w HttpWriter, r HttpReq) {
	var render = true
	var prender bool
	for _, plug := range s.Plugins {
		prender, s.Info[plug.Name] = plug.Plug(w, r, s.Info)
		render = render&&prender
	}
	prender, s.Info["logic"] = s.Fn(w, r, s.Info)
	render = render && prender && (s.Template != nil)

	if (render) {
		s.Template.Execute(w, s.Info)
	}
}

func LogicPage(
	filename string,
	info map[string]any,
	plugins []GOTMPlugin,
	fn Plugin,
) (LogicedPage) {
	if info == nil {
		info = make(map[string]any)
	}

	var tmpl *template.Template = nil
	if (filename != "") {
		tmpl = template.Must(
			template.Must(
				template.ParseFiles(filename),
			).ParseGlob("templates/*.gohtml"),
		)
	}

	return LogicedPage{
		tmpl, info, plugins, fn,
	}
}

