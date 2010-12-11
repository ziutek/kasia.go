package kasia

import "testing"

type Test struct {
    txt    string
    expect string
    strict bool
    ctx    interface{}
}


var (

f1 = func() int { return 0 }

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
    // slice:
    `$[0], $[1], $[2], $[-1], $[-2], $[-3]`,
    `A, B, C, C, B, A`, true,
    []string {"A", "B", "C"},
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
