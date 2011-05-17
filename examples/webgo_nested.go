package main

import (
    "fmt"
    "github.com/ziutek/kasia.go"
    "github.com/hoisie/web.go"
)

type Ctx struct {
    Cnt    int
    Txt    string
    Subtpl *kasia.Template // Nested template
    Subctx *SubCtx         // Subcontext
}

type SubCtx struct {
    Vals []string
}

var (
    tpl *kasia.Template
    data = &Ctx{0, "Hello!", nil, &SubCtx{}}
)

func hello(web_ctx *web.Context, val string) {
    // Note that it's not safe to modify global data in multithread enviroment
    data.Cnt++
    data.Subctx.Vals = append(data.Subctx.Vals, val)

    // Render data
    err := tpl.Run(web_ctx, data)
    if err != nil {
        fmt.Fprint(web_ctx, err)
    }
}


func main() {
    // Main template
    tpl = kasia.New()
    err := tpl.Parse([]byte("<html><body>Request #$Cnt: $Txt<br>\n" +
                            "$Subtpl.Nested(Subctx)</body></html>\n"))
    if err != nil {
        fmt.Println("Main template", err)
        return
    }

    // Nested template
    data.Subtpl = kasia.New()
    err = data.Subtpl.Parse([]byte("$for i+, v in Vals: $i: $v<br>\n$end"))
    if err != nil {
        fmt.Println("Nested template", err)
        return
    }

    // Web.go
    web.Get("/(.*)", hello)
    web.Run("0.0.0.0:9999")
}

