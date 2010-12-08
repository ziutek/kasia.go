package main

import (
    "kasia"
    "web"
    "fmt"
)

type Ctx struct {
    a, b  string
}

type LocalCtx struct {
    b string
    c int
}

var (
    tpl *kasia.Template
    data = &Ctx{"Hello!", "not specified"}
)

func hello(web_ctx *web.Context, val string) {
    // Local data
    var ld *LocalCtx

    if len(val) > 0 {
        ld = &LocalCtx{val, len(val)}
    }

    // Rendering data
    err := tpl.Run(web_ctx, data, ld)
    if err != nil {
        fmt.Fprint(web_ctx, "%", err, "%")
    }
}

const tpl_txt = `
<html><body>
    $a<br>
    Parameter: $b
    $if c:
        , length: $c
    $end
</body></html>
`

func main() {
    // Main template
    tpl = kasia.New()
    err := tpl.Parse(tpl_txt)
    if err != nil {
        fmt.Println("Main template", err)
        return
    }

    // This example can work in strict mode
    tpl.Strict = true

    // Web.go
    web.Get("/(.*)", hello)
    web.Run("0.0.0.0:9999")
}

