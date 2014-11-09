package templatex

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var (
	Layout  string
	IsDebug bool
	Funcs   template.FuncMap
)

func init() {
	Funcs = make(template.FuncMap)
}

func New(root string) func(io.Writer, string, interface{}) error {
	funcs := template.FuncMap{
		`yield`: func() error {
			return fmt.Errorf(`yield can't call`)
		},
		`current`: func() error {
			return fmt.Errorf(`current can't call`)
		},
	}

	for n, f := range Funcs {
		funcs[n] = f
	}

	var tpl *template.Template
	return func(w io.Writer, name string, v interface{}) error {
		if tpl == nil || IsDebug {
			tpl = template.Must(template.New(root).Funcs(funcs).Parse(`Templatex`))

			err := filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
				if fi.IsDir() {
					return nil
				}

				if b, err := ioutil.ReadFile(path); err != nil {
					return err
				} else {
					nt := tpl.New(fi.Name())
					_, err := nt.Funcs(funcs).Parse(string(b))
					return err
				}
			})
			if err != nil {
				return err
			}
		}

		if tc, err := tpl.Clone(); err != nil {
			log.Printf(`Can't clone template, %s`, err)
			return err
		} else if len(Layout) > 0 {
			lfs := template.FuncMap{
				`yield`: func() (template.HTML, error) {
					if yc, err := tpl.Clone(); err != nil {
						return template.HTML(``), err
					} else {
						var buf bytes.Buffer
						if err := yc.ExecuteTemplate(&buf, name, v); err != nil {
							return template.HTML(``), err
						} else {
							return template.HTML(buf.String()), nil
						}
					}
				},
				`current`: func() string {
					return name
				},
			}
			return tc.Funcs(lfs).ExecuteTemplate(w, Layout, v)
		} else {
			return tc.ExecuteTemplate(w, name, v)
		}
	}
}
