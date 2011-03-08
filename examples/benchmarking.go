package main

import (
    "time"
    "os"
    "fmt"
    "template"
    "github.com/ziutek/kasia.go"
)

const (
bench_kt = `
$for i, v in Arr:
    $i: $v.A, $v.B
$end
`
bench_tpl = `
{.repeated section Arr}
    {.section @}
        {I}: {A} {B}
    {.end}
{.end}
`
)

var bctx = struct {
    Arr [1000]map[string]interface{}
}{}

type DevNull struct{}
func (*DevNull) Write(p []byte) (int, os.Error) {
    return len(p), nil
}
var devnull DevNull

func kbench(n int) int64 {
    tpl := kasia.MustParse(bench_kt)
    tpl.EscapeFunc = nil
    start := time.Nanoseconds()
    for ii := 0; ii < n; ii++ {
        err := tpl.Run(&devnull, bctx)
        if err != nil {
            panic(err)
        }
    }
    return int64(n) * 1e12 / (time.Nanoseconds() - start)
}

func tbench(n int) int64 {
    tpl := template.MustParse(bench_tpl, nil)
    start := time.Nanoseconds()
    for ii := 0; ii < n; ii++ {
        err := tpl.Execute(&devnull, bctx)
        if err != nil {
            panic(err)
        }
    }
    return int64(n) * 1e12 / (time.Nanoseconds() - start)
}

func main() {
    // Szykujemy dane
    for ii := 0; ii < len(bctx.Arr); ii++ {
        bctx.Arr[ii] = map[string]interface{}{
            "I": ii, "A": ii * 2,
            "B": `aąbcćdeęfghijklłmnńoóprsśtuvwxżż
                  AĄBCĆDEĘFGHIJKLŁMNŃOÓPRSŚTUVWXYŻŹ`,
        }
    }
    n := 100
    fmt.Printf("Kasia: %d, template: %d [r/s]\n", kbench(n), tbench(n))
}
