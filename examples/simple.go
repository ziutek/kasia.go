package main

import (
    "os"
    "fmt"
    "github.com/ziutek/kasia.go"
)

type Ctx struct {
    H, W string
}

func main() {
    ctx := &Ctx{"Hello", "world"}

    tpl, err := kasia.Parse("$H $W!\n")
    if err != nil {
        fmt.Println(err)
        return
    }

    err = tpl.Run(os.Stdout, ctx)
    if err != nil {
        fmt.Println(err)
    }
}
