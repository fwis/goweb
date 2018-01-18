package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Included helper functions for use when rendering HTML.
var helperFuncs = template.FuncMap{
	"yield": func() (string, error) {
		return "", fmt.Errorf("yield called with no layout defined")
	},
	"current": func() (string, error) {
		return "", nil
	},
}

// Delims represents a set of Left and Right delimiters for HTML template rendering.
type Delims struct {
	// Left delimiter, defaults to {{.
	Left string
	// Right delimiter, defaults to }}.
	Right string
}

type TplOption struct {
	// Directory to load templates. Default is "tpl".
	Directory string
	// Layout template name. Will not render a layout if blank (""). Defaults to blank ("").
	Layout string
	// Extensions to parse template files from. Defaults to [".tpl"].
	Extensions []string
	// Funcs is a slice of FuncMaps to apply to the template upon compilation. This is useful for helper functions. Defaults to [].
	FuncMap template.FuncMap
	// Delims sets the action delimiters to the specified strings in the Delims struct.
	Delims Delims
	// If IsDevelopment is set to true, this will recompile the templates on every request. Default is false.
	IsDevelopment bool
}

// Render is a service that provides functions for easily writing JSON, XML,
// binary data, and HTML templates out to a HTTP Response.
type HTMLRenderEngine struct {
	// Customize Secure with an Options struct.
	Option    TplOption
	templates *template.Template
}

func NewDefaultHTMLRenderEngine(tplRootDir string, f template.FuncMap) *HTMLRenderEngine {
	r := &HTMLRenderEngine{
		Option: TplOption{},
	}
	r.Option.Directory = tplRootDir
	r.Option.FuncMap = f
	r.prepareOptions()
	r.compileTemplates()

	return r
}

func (r *HTMLRenderEngine) prepareOptions() {
	if len(r.Option.Directory) == 0 {
		r.Option.Directory = "tpl"
	}
	if len(r.Option.Extensions) == 0 {
		r.Option.Extensions = []string{".tpl"}
	}
}

func (r *HTMLRenderEngine) compileTemplates() {
	dir := r.Option.Directory
	r.templates = template.New(dir)
	r.templates.Delims(r.Option.Delims.Left, r.Option.Delims.Right)

	// Walk the supplied directory and compile any files that match our extension list.
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		ext := ""
		if strings.Index(rel, ".") != -1 {
			ext = "." + strings.Join(strings.Split(rel, ".")[1:], ".")
		}

		for _, extension := range r.Option.Extensions {
			if ext == extension {

				buf, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}

				name := (rel[0 : len(rel)-len(ext)])
				tmpl := r.templates.New(filepath.ToSlash(name))

				// Add our funcmaps.
				for key, tfunc := range r.Option.FuncMap {
					helperFuncs[key] = tfunc
				}

				// Break out if this parsing fails. We don't want any silent server starts.
				template.Must(tmpl.Funcs(helperFuncs).Parse(string(buf)))
				break
			}
		}

		return nil
	})
}

// HTML builds up the response from the specified template and bindings.
func (r *HTMLRenderEngine) Render(name string, tplLayout string, binding interface{}) ([]byte, error) {
	// If we are in development mode, recompile the templates on every HTML request.
	if r.Option.IsDevelopment {
		r.compileTemplates()
	}

	// DO NOT Assign default layout.
	// if tplLayout == "" && len(r.Option.Layout) > 0 {
	// tplLayout = r.Option.Layout
	// }

	if len(tplLayout) > 0 {
		r.addYield(name, binding)
		name = tplLayout
	}

	out := new(bytes.Buffer)
	err := r.templates.ExecuteTemplate(out, name, binding)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func (r *HTMLRenderEngine) execute(name string, binding interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	return buf, r.templates.ExecuteTemplate(buf, name, binding)
}

func (r *HTMLRenderEngine) addYield(name string, binding interface{}) {
	funcs := template.FuncMap{
		"yield": func() (template.HTML, error) {
			buf, err := r.execute(name, binding)
			// Return safe HTML here since we are rendering our own template.
			return template.HTML(buf.String()), err
		},
		"current": func() (string, error) {
			return name, nil
		},
	}
	r.templates.Funcs(funcs)
}
