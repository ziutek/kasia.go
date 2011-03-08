package main

import (
    "fmt"
    "github.com/ziutek/kasia.go"
    "github.com/hoisie/web.go"
)

type Ctx struct {
    cnt    int
    txt    string
    subtpl *kasia.Template // Nested template
    subctx *SubCtx         // Subcontext
}

type SubCtx struct {
    vals []string
}

var (
    tpl *kasia.Template
    data = &Ctx{0, "Hello!", nil, &SubCtx{}}
)

func hello(web_ctx *web.Context, val string) {
    // Note that it's not safe to modify global data in multithread enviroment
    data.cnt++
    data.subctx.vals = append(data.subctx.vals, val)

    // Render data
    err := tpl.Run(web_ctx, data)
    if err != nil {
        fmt.Fprint(web_ctx, err)
    }
}


func main() {
    // Main template
    tpl = kasia.New()
    err := tpl.Parse([]byte("<html><body>Request #$cnt: $txt<br>\n" +
                            "$subtpl.Nested(subctx)</body></html>\n"))
    if err != nil {
        fmt.Println("Main template", err)
        return
    }

    // Nested template
    data.subtpl = kasia.New()
    err = data.subtpl.Parse([]byte("$for i+, v in vals: $i: $v<br>\n$end"))
    if err != nil {
        fmt.Println("Nested template", err)
        return
    }

    // Web.go
    web.Get("/(.*)", hello)
    web.Run("0.0.0.0:9999")
}

