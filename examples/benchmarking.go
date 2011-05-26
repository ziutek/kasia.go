package main

import (
    "os"
    "fmt"
    "template"
    "testing"
    "github.com/ziutek/kasia.go"
)

const (
bench_kt = `
$for i, v in Arr:
    $i: $v.A $v.B
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

var (
    devnull DevNull
    gtpl *template.Template
    ktpl *kasia.Template
)

func kbench(b *testing.B) {
    for ii := 0; ii < b.N; ii++ {
        err := ktpl.Run(&devnull, bctx)
        if err != nil {
            panic(err)
        }
    }
}

func tbench(b *testing.B) {
    for ii := 0; ii < b.N; ii++ {
        err := gtpl.Execute(&devnull, bctx)
        if err != nil {
            panic(err)
        }
    }
}

func main() {
    ktpl = kasia.MustParse(bench_kt)
    ktpl.EscapeFunc = nil
    gtpl = template.MustParse(bench_tpl, nil)
    for ii := 0; ii < len(bctx.Arr); ii++ {
        bctx.Arr[ii] = map[string]interface{}{
            "I": ii, "A": ii * 2,
            "B": `aąbcćdeęfghijklłmnńoóprsśtuvwxżż
                  AĄBCĆDEĘFGHIJKLŁMNŃOÓPRSŚTUVWXYŻŹ`,
        }
    }
    fmt.Println("Kasia:")
    fmt.Println(testing.Benchmark(kbench))
    fmt.Println("Go template:")
    fmt.Println(testing.Benchmark(tbench))
}
