package main

import (
    "os"
    "fmt"
    "github.com/ziutek/kasia.go"
)

type Ctx struct {
    h, w string
}

func main() {
    ctx := &Ctx{"Hello", "world"}

    tpl, err := kasia.Parse("$h $w!\n")
    if err != nil {
        fmt.Println(err)
        return
    }

    err = tpl.Run(os.Stdout, ctx)
    if err != nil {
        fmt.Println(err)
    }
}
