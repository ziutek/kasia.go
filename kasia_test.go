package kasia

import (
    "testing"
    "bytes"
    "fmt"
)

type Test struct {
    txt    string
    expect string
    strict bool
    ctx    interface{}
}

type S1 struct {
    a int
    b float
}

var (

f1 = func() int { return 0 }
tpl1, _ = Parse("$a $b")

tests = []Test{
{
    // Tekst bez zmiennych
    `Łódź, róża, łąka, $'$"$$.`,
    `Łódź, róża, łąka, '"$.`, true,
    nil,
},{
    // map: Zmienna i funkcja
    `$aa, $bc(1)`,
    `Ala, 2`, true,
    map[string]interface{}{"aa": "Ala", "bc": func(i int)int{return 2*i}},
},{
    // map: Zmienna i funkcja, indeksy
    `$["aa"], $['bc'](1)`,
    `Ala, 2`, true,
    map[string]interface{}{"aa": "Ala", "bc": func(i int)int{return 2*i}},
},{
    // struct: 'if', funkcja zwracajaca funkcje, complex
    `$if f(1)('str')(2.1) == c: OK $end $f(a)("eeee")(b) $if b==2.1:$a$b$c$end`,
    ` OK  (6+2.1i) 22.1(4+2.1i)`, true,
    struct {
        f func(int)func(string)func(float)complex
        a int
        b float
        c complex
    } {
        func(i int) func(string)func(float)complex {
            return func(s string) func(float)complex {
                return func(f float) complex {
                    return cmplx(float(i + len(s)), f)
                }
            }
        },
        2,
        2.1,
        cmplx(4, 2.1),
    },
},{
    // struct: 'if' jednoargumentowy, funkcja bez argumentow
    `$if f: $f() $end $if f(): First $else: Second $end`,
    ` 0   Second `, true,
    struct {f *func()int}{&f1},
},{
    // func:
    `$((((((((1)))))))) $(0)`,
    `9 1`, true,
    func (i int) int {i++; return i},
},{
    // map: DotDotDot function, int args
    `$fun(1), $fun(1, 2), $fun(1, 2, 3)`,
    `[1], [1 2], [1 2 3]`, true,
    map[string]func(...int)string {
        "fun": func(n ...int) string {
            return fmt.Sprint([]int(n))
        },
    },
},{
    // map: DotDotDot function, interface args
    `$fun(1) $fun(1, 2.2) $fun(1, 2.2, "text") $fun(1, 2.2, "text", val)`,
    "1\n 1 2.2\n 1 2.2 text\n 1 2.2 text true\n", true,
    map[string]interface{} {
        "fun": func(n ...interface{}) string {
            return fmt.Sprintln(n...)
        },
        "val": true,
    },
},{
    // slice:
    `$[0], $[1], $[2], $[-1], $[-2], $[-3]`,
    `A, B, C, C, B, A`, true,
    []string {"A", "B", "C"},
},{
    // struct: 'for'
    "$for i+, v in arr:$i: $v.a, $v.b\n$end\n$for i,v in x:A$else:B$end\n\n",
    "1: 0, 0.5\n2: 1, 0.25\n3: 2, 0.125\nB\n", true,
    struct{arr []S1}{[]S1{S1{0, 1/2.0}, S1{1, 1/4.0}, S1{2, 1/8.0}}},
},{
    //map: Nested templates
    "$tpl\n $tpl.Nested(ctx1)\n $tpl.Nested(ctx1, ctx2)\n",
    "1 1.1\n 2 2.2\n 2 3.3\n", true,
    map[string]interface{} {
        "tpl": tpl1,
        "a":   1,
        "b":   1.1,
        "ctx1": S1{2, 2.2},
        "ctx2": map[string]float{"b": 3.3},
    },
},
}

)

func TestAll(t *testing.T) {
    for _, te := range tests {
        out, err := RenderString(te.txt, te.strict, te.ctx)
        strict := ""
        if te.strict {
            strict = "!"
        }
        if err != nil {
            t.Fatalf("`%s`%s err: %s", te.txt, strict, err)
        } else if out != te.expect {
            t.Fatalf("`%s`%s out `%s` expect `%s`",
                te.txt, strict, out, te.expect)
        }
    }
}


const bench_kt = `
$for i, v in arr:
    $i: $v.a, $v.b
$end
`

var bctx = struct {
    arr [1000]S1
}{}


func BenchmarkFor(b *testing.B) {
    b.StopTimer()
    // Szykujemy dane
    for ii := 0; ii < len(bctx.arr); ii++ {
        bctx.arr[ii].a = ii*ii
        bctx.arr[ii].b = 1.0 / (1.0 + float(ii))
    }
    // Tworzymy szablon
    tpl := New()
    tpl.Parse(bench_kt)
    tpl.Strict = true
    var buf bytes.Buffer
    // Startujemy test
    b.StartTimer()
    for ii := 0; ii < b.N; ii++ {
        buf.Reset()
        err := tpl.Run(&buf, bctx)
        if err != nil {
            panic(err)
        }
    }
}
