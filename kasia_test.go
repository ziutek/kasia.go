package kasia

import (
    "testing"
    "fmt"
    "os"
    "template"
)

type Test struct {
    txt    string
    expect string
    strict bool
    ctx    interface{}
}

type S1 struct {
    A int
    B float64
}

var (

f1 = func() int { return 0 }
tpl1, _ = Parse("$A $B")

tests = []Test{
{
    // Text in UTF-8 encoding
    "Łódź, róża, łąka, $'$\"$`$$.",
    "Łódź, róża, łąka, '\"`$.", true,
    nil,
},{
    // Comments
    "This is a $#tttt,\n $a, $c\n $# $$ #$ comment!",
    "This is a  comment!", true,
    nil,
},{
    // map: []byte
    `$aa`,
    `Ala`, true,
    map[string]interface{}{"aa": []byte("Ala")},
},{
    // map: Variable and function
    `$aa, $bc(1)`,
    `Ala, 2`, true,
    map[string]interface{}{"aa": "Ala", "bc": func(i int)int{return 2*i}},
},{
    // map: Variable and function, string key
    `$["aa"], $['bc'](1)`,
    `Ala, 2`, true,
    map[string]interface{}{"aa": "Ala", "bc": func(i int)int{return 2*i}},
},{
    // map: Integer key, not strict mode
    `'$[1]', '$[-101]', '$[2]'`,
    `'a', 'b', ''`, false,
    map[int]string{1: "a", -101: "b"},
},{
    // map: Float key, not strict mode
    `'$[0.1]', '$[-1.1]', '$[2.2]'`,
    `'a', 'b', ''`, false,
    map[float64]string{0.1: "a", -1.1: "b"},
},{
    // map: if elif
    `$if Kind.a: A $elif Kind.b: B $else: unk $end`,
    ` B `, true,
    struct{Kind map[string]bool}{map[string]bool{"b": true}},
},{
    // map: Any key, any value
    "$[1]\n$[2.2]\n$[i]\n$[t]\n$i\n$t\n",
    "a\nb\n-1.9\n100\n(0+1i)\ntrue\n", true,
    map[interface{}]interface{} {
        1           : "a",
        2.2         : "b",
        complex(0,1)  : -1.9,
        true        : 100,
        "i"         : complex(0,1),
        "t"         : true,
        },
},{
    // struct: 'if', funkcja zwracajaca funkcje, complex
    `$if F(1)('str')(2.1) == C:OK $end $F(A)("eeee")(B) $if B==2.1:$A$B$C$end`,
    `OK  (6+2.1i) 22.1(4+2.1i)`, true,
    struct {
        F func(int)func(string)func(float64)complex128
        A int
        B float64
        C complex128
    } {
        func(i int) func(string)func(float64)complex128 {
            return func(s string) func(float64)complex128 {
                return func(f float64) complex128 {
                    return complex(float64(i + len(s)), f)
                }
            }
        },
        2,
        2.1,
        complex(4, 2.1),
    },
},{
    // struct: 'if' jednoargumentowy, funkcja bez argumentow
    `$if F: $F() $end $if F(): First $else: Second $end`,
    ` 0   Second `, true,
    struct {F *func()int}{&f1},
},{
    // func:
    `$((((((((1)))))))) $(0) $(1) $((-2))`,
    `9 1 2 0`, true,
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
    "$fun(1) $fun(1, 2.2) $fun(1, 2.2, 'text') $fun(1, 2.2, `text`, val)",
    "1\n 1 2.2\n 1 2.2 text\n 1 2.2 text true\n", true,
    map[string]interface{} {
        "fun": func(n ...interface{}) string {
            return fmt.Sprintln(n...)
        },
        "val": true,
    },
},{
    // slice:
    `$[0], $[1], $[2]`,
    `A, B, C`, true,
    []string {"A", "B", "C"},
},{
    // struct: 'for'
    "$for i+, v in Arr:$i: $v.A, $v.B\n$end\n$for i,v in x:A$else:B$end\n\n",
    "1: 0, 0.5\n2: 1, 0.25\n3: 2, 0.125\nB\n", true,
    struct{Arr []S1}{[]S1{S1{0, 1/2.0}, S1{1, 1/4.0}, S1{2, 1/8.0}}},
},{
    //map: Nested templates
    "$tpl\n $tpl.Nested(ctx1)\n $tpl.Nested(ctx1, ctx2)\n",
    "1 1.1\n 2 2.2\n 2 3.3\n", true,
    map[string]interface{} {
        "tpl": tpl1,
        "A":   1,
        "B":   1.1,
        "ctx1": S1{2, 2.2},
        "ctx2": map[string]float64{"B": 3.3},
    },
},{
    //int
    `$@[0] $@[0].`, `7 7.`, true, 7,
},{
    //string: HTML escaping
    `$@[0] $:@[0]`, `&amp;&apos;&lt;&gt;&quot; &'<>"`, true, `&'<>"`,
},{
    //map: iterate over map
    "$for k, v in @[0]: $k: $v $end",
    " a: 11  34: aa ", true,
    map[interface{}]interface{} {"a": 11, 34: "aa"},
},
}

)

func check(t *testing.T, tpltxt, exp string, strict bool, ctx ...interface{}) {
    out, err := RenderString(tpltxt, strict, ctx...)
    str := ""
    if strict {
        str = "!"
    }
    if err != nil {
        t.Fatalf("`%s`%s err: %s", tpltxt, str, err)
    } else if out != exp {
        t.Fatalf("`%s`%s out `%s` expect `%s`", tpltxt, str, out, exp)
    }
}


func TestAll(t *testing.T) {
    for _, te := range tests {
        check(t, te.txt, te.expect, te.strict, te.ctx)
    }
}


type Environment struct {
    MsgHost, MsgPath, SubmitPath, StaticPath string
}
var env = Environment {
    MsgHost:    "www.lnet.pl",
    MsgPath:    "/message/",
    SubmitPath: "/submit/",
    StaticPath: "/static/",
}


type MsgCtx struct {
    StartData string
    Tekst      string
    Tpl        *Template
}

func TestMultiCtx(t *testing.T) {
    divs := map[string]string{"content": "KKKKK"}
    tpl, _ := Parse("$[0].content $[1].StaticPath")
    ctx := MsgCtx{"1234-22-33", "abcd", tpl}

    tpl_txt := "$content $StartData $Tekst $MsgHost $MsgPath $Tpl.Nested(@)"
    expect := "KKKKK 1234-22-33 abcd www.lnet.pl /message/ KKKKK /static/"

    check(t, tpl_txt, expect, true, divs, env, ctx)
}


func TestMultiFuncCtx(t *testing.T) {
    fun1 := func(a int) int { return 2 * a }
    fun2 := func() string { return "N" }
    fun3 := func(a int, s string) string { return fmt.Sprintf("%d:%s", a, s) }

    tpl_txt := "$(1) $() $(3, 'aaa')"
    expect := "2 N 3:aaa"

    check(t, tpl_txt, expect, true, fun1, fun2, fun3, "bla bla")
}

func TestMultiSlice(t *testing.T) {
    s1 := []float32{1.1, 2.2, 3.3}
    s2 := []string{"a", "b"}
    s3 := []int{1}

    tpl_txt := "$[0] $[1] $[2]"
    expect := "1 b 3.3"

    check(t, tpl_txt, expect, true, s1, s2, s3, "bla bla")
}

func TestCtxStackUnstr(t *testing.T) {
    tpl_txt := "$@[0] $@[1] $@[2]"
    expect := "2 Ala 3.14159"

    check(t, tpl_txt, expect, true, 2, "Ala", 3.14159)
}

func TestCtxStackBoolean(t *testing.T) {
	tpl_txt := "$if @[0]: Should not be reached $end"
	expect := ""

	check(t, tpl_txt, expect, true, false) 

}

func xrange(start, stop int) <-chan int {
    c := make(chan int)
    go func() {
        for i := start; i < stop; i++ {
            c <- i
        }
        close(c)
    }()
    return c
}


func TestChannel(t *testing.T) {
    tpl_txt := "$for i+,v in xrange(4,8): $i,$v $end"
    expect := " 1,4  2,5  3,6  4,7 "
    ctx := map[string]interface{}{"xrange": xrange}
    check(t, tpl_txt, expect, true, ctx)
}


type Tmf struct {
    m int
}
func (t *Tmf) Mulf(i int) func(int)int {
    return func(c int)int { return t.m * i * c }
}
func (t *Tmf) Mul(i int) func()int {
    return func()int { return t.m * i }
}

func TestMethodFunc(t *testing.T) {
    ctx := &Tmf{2}
    tpl_txt := "$Mulf(3)(5) $Mul(3)"
    expect := "30 6"
    check(t, tpl_txt, expect, true, ctx)
}

// BENCHMARKING

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
    ktpl *Template
    gtpl *template.Template
)

func BenchmarkKasia(b *testing.B) {
    for ii := 0; ii < b.N; ii++ {
        err := ktpl.Run(&devnull, bctx)
        if err != nil {
            panic(err)
        }
    }
}

func BenchmarkTemplates(b *testing.B) {
    for ii := 0; ii < b.N; ii++ {
        err := gtpl.Execute(&devnull, bctx)
        if err != nil {
            panic(err)
        }
    }
}

func init() {
    ktpl = MustParse(bench_kt)
    ktpl.EscapeFunc = nil
    gtpl = template.MustParse(bench_tpl, nil)
    for ii := 0; ii < len(bctx.Arr); ii++ {
        bctx.Arr[ii] = map[string]interface{}{
            "I": ii, "A": ii * 2,
            "B": `aąbcćdeęfghijklłmnńoóprsśtuvwxżż
                  AĄBCĆDEĘFGHIJKLŁMNŃOÓPRSŚTUVWXYŻŹ`,
        }
    }
}
