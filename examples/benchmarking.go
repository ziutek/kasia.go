package main

import (
    "time"
    "os"
    "fmt"
    "github.com/ziutek/kasia.go"
)

const bench_kt = `
$for i, v in arr:
    $i: $v
$end
`

var bctx = struct {
    arr [1000]int
}{}

type DevNull struct{}
func (*DevNull) Write(p []byte) (int, os.Error) {
    return len(p), nil
}

func main() {
    // Szykujemy dane
    for ii := 0; ii < len(bctx.arr); ii++ {
        bctx.arr[ii] = ii*ii
    }
    // Tworzymy szablon
    tpl := kasia.New()
    tpl.EscapeFunc = nil
    tpl.Parse([]byte(bench_kt))
    tpl.Strict = true
    var devnull DevNull
    start := time.Nanoseconds()
    n := 100
    for ii := 0; ii < n; ii++ {
        err := tpl.Run(&devnull, bctx)
        if err != nil {
            panic(err)
        }
    }
    stop := time.Nanoseconds()
    fmt.Printf("%d ns/op\n", (stop - start) / int64(n))
}
