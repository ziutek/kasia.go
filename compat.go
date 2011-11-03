package kasia

import (
	"io"
	"io/ioutil"
	"bytes"
)

// Renderowanie do tekstu lub tablicy bajtow

func (tpl *Template) RenderBytes(ctx ...interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := tpl.Run(&buf, ctx...)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (tpl *Template) RenderString(ctx ...interface{}) (string, error) {
	bytes, err := tpl.RenderBytes(ctx...)
	return string(bytes), err
}

func RenderBytes(txt string, strict bool, ctx ...interface{}) ([]byte, error) {

	tpl := New()
	tpl.Strict = strict

	err := tpl.Parse([]byte(txt))
	if err != nil {
		return nil, err
	}

	return tpl.RenderBytes(ctx...)
}

func RenderString(txt string, strict bool, ctx ...interface{}) (string, error) {

	bytes, err := RenderBytes(txt, strict, ctx...)
	return string(bytes), err
}

// Funkcje i metody kompatybilne z Go template

func (tpl *Template) Execute(wr io.Writer, data interface{}) error {
	return tpl.Run(wr, data)
}

func (tpl *Template) ParseFile(filename string) (err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	return tpl.Parse(data)
}

func Parse(txt string) (tpl *Template, err error) {
	tpl = New()
	err = tpl.Parse([]byte(txt))
	if err != nil {
		tpl = nil
	}
	return
}

func MustParse(txt string) (tpl *Template) {
	tpl, err := Parse(txt)
	if err != nil {
		panic(err)
	}
	return
}

func ParseFile(filename string) (tpl *Template, err error) {
	tpl = New()
	err = tpl.ParseFile(filename)
	if err != nil {
		tpl = nil
	}
	return
}

func MustParseFile(filename string) (tpl *Template) {
	tpl, err := ParseFile(filename)
	if err != nil {
		panic(err)
	}
	return
}

// Funkcje i metody kompatybilne z mustache.go

func (tpl *Template) Render(ctx ...interface{}) (out string) {
	var err error
	out, err = tpl.RenderString(ctx...)
	if err != nil {
		panic(err)
	}
	return
}

func Render(txt string, ctx ...interface{}) (out string) {
	var err error
	out, err = RenderString(txt, false, ctx...)
	if err != nil {
		panic(err)
	}
	return
}
