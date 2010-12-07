package kasia

import (
    "io"
    "io/ioutil"
    "os"
    "bytes"
)

// Funkcje i metody kompatybilne z Go template.

func (tpl *Template) Execute(data interface{}, wr io.Writer) os.Error {
    return tpl.Run(wr, data)
}

func (tpl *Template) ParseFile(filename string) (err os.Error) {
    data, err := ioutil.ReadFile(filename)
    return tpl.Parse(string(data))
}

func Parse(str string) (tpl *Template, err os.Error) {
    tpl = New()
    err = tpl.Parse(str)
    if err != nil {
        tpl = nil
    }
    return
}

func ParseFile(filename string) (tpl *Template, err os.Error) {
    tpl = New()
    err = tpl.ParseFile(filename)
    if err != nil {
        tpl = nil
    }
    return
}

// Funkcje i metody kompatybilne z mustache.go

func (tpl *Template) Render(ctx ...interface{}) string {
    var buf bytes.Buffer
    err := tpl.Run(&buf, ctx...)
    if err != nil {
        panic(err)
    }
    return buf.String()
}
