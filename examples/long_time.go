package main

import (
    "fmt"
    "time"
    "github.com/ziutek/kasia.go"
)

type ContextData struct {
    A   int
    B   string
    C   bool
    D   *[]interface{}
    E   map[string]interface{}
    F   func(int) interface{}
    G   *ContextData
}

func (cd ContextData) M1() int {
    return cd.A
}

func (cd *ContextData) M2(s string) bool {
    return cd.B == s
}


func main() {
    data_func := func(a int) interface{} {
        if a == 0 {
            return func(s string, i int) string {
                return fmt.Sprintf(">>%s, %d<<", s, i)
            }
        }
        return "'a != 0'"
    }

    data_slice := []interface{}{
        1,
        func(f float64) string {
            return fmt.Sprintf("First value >%f<.", f)
        },
        "text in slice",
        false,
    }

    data_struct := ContextData{
        2,
        "A",
        true,
        &data_slice, //data_slice[:len(data_slice)-1],
        nil,
        data_func,
        nil,
    }

    data_map := map[string]interface{}{
        "a":    1,
        "b":    "'&......................text in map....................&'\n",
        "c":    false,
        "d":    data_slice[1:],
        "e":    nil,
        "f":    &data_func,
        "g":    &data_struct,
    }

    data_map["e"] = data_map
    data_struct.E = data_map
    data_struct.G = &data_struct

    bgn := time.Nanoseconds()
    tpl, err := kasia.ParseFile("long_time.kt")
    if err != nil {
        fmt.Println(err)
        return
    }
    tpl.Strict = true
    end := time.Nanoseconds()
    fmt.Printf("Parsing: %d us\n", (end-bgn) / 1000)

    n := 1000
    for ii := 1;; ii++ {
        bgn := time.Nanoseconds()
        txt := ""
        for ii := 0; ii < n; ii++ {
            txt = tpl.Render(data_struct)
            if err != nil  {
                fmt.Println(err)
                return
            }
        }
        end := time.Nanoseconds()
        fmt.Printf("Rendered %d*%d B: %.0f r/s, %.0f kB/s\n",
            ii*n, len(txt), 1e9 * float64(n) / float64(end-bgn),
            1e6 * float64(len(txt) * n) / float64(end-bgn))

    }
}
