package kasia

import (
    "io"
    "io/ioutil"
    "os"
    "bytes"
)

// Renderowanie do tekstu lub tablicy bajtow

func (tpl *Template) RenderBytes(ctx ...interface{}) ([]byte, os.Error) {
    var buf bytes.Buffer
    err := tpl.Run(&buf, ctx...)
    if err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

func (tpl *Template) RenderString(ctx ...interface{}) (string, os.Error) {
    bytes, err := tpl.RenderBytes(ctx...)
    return string(bytes), err
}

func RenderBytes(txt string, strict bool, ctx ...interface{}) (
        []byte, os.Error) {

    tpl := New()
    tpl. Strict = strict

    err := tpl.Parse(txt)
    if err != nil {
        return nil, err
    }

    return tpl.RenderBytes(ctx...)
}

func RenderString(txt string, strict bool, ctx ...interface{}) (
        string, os.Error) {

    bytes, err := RenderBytes(txt, strict, ctx...)
    return string(bytes), err
}

// Funkcje i metody kompatybilne z Go template

func (tpl *Template) Execute(data interface{}, wr io.Writer) os.Error {
    return tpl.Run(wr, data)
}

func (tpl *Template) ParseFile(filename string) (err os.Error) {
    data, err := ioutil.ReadFile(filename)
    return tpl.Parse(string(data))
}

func Parse(txt string) (tpl *Template, err os.Error) {
    tpl = New()
    err = tpl.Parse(txt)
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

func (tpl *Template) Render(ctx ...interface{}) (out string) {
    var err os.Error
    out, err = tpl.RenderString(ctx...)
    if err != nil {
        panic(err)
    }
    return
}


func Render(txt string, ctx ...interface{}) (out string) {
    var err os.Error
    out, err = RenderString(txt, false, ctx...)
    if err != nil {
        panic(err)
    }
    return
}
