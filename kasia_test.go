package kasia

import "testing"

type Test struct {
    txt    string
    strict bool
    ctx    interface{}
    expect string
}

var tests = []Test{
{
    "Łódź, róża, łąka", true,
    nil,
    "Łódź, róża, łąka",
},{
    "$aa, $bc(1)", true,
    map[string]interface{}{"aa": "Ala", "bc": func(i int)int{return 2*i}},
    "Ala, 2",
},{
    `$["aa"], $['bc'](1)`, true,
    map[string]interface{}{"aa": "Ala", "bc": func(i int)int{return 2*i}},
    "Ala, 2",
},
}

func TestAll(t *testing.T) {
    for _, te := range tests {
        out, err := RenderTxt(te.txt, te.strict, te.ctx)
        strict := ""
        if te.strict {
            strict = "/strict"
        }
        if err != nil {
            t.Fatalf("|%s|%s err: %s", te.txt, strict, err)
        } else if out != te.expect {
            t.Fatalf("|%s|%s out:|%s| expect:|%s|",
                te.txt, strict, out, te.expect)
        }
    }
}
