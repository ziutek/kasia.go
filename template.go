package kasia

import (
    "os"
    "io"
    "fmt"
    str "strings"
    "bytes"
    "reflect"
)

type Template struct {
    elems      []Element
    Strict     bool
    EscapeFunc func(io.Writer, string) os.Error
}

type NestedTemplate struct {
    tpl *Template
    ctx []interface{}
}

// Okresla wartosc parametru.
func execParam(wr io.Writer, par interface{}, ctx []interface{}, ln int,
        strict bool) (ret reflect.Value, err os.Error) {
    switch pv := par.(type) {
    case *reflect.IntValue:
        // Argumentem jest liczba staloprzecinowa
        ret = pv

    case  *reflect.FloatValue:
        // Argumentem jest liczba zmiennoprzecinkowa
        ret = pv

    case []Element:
        // Argumentem jest sparsowany tekst
        var buf bytes.Buffer
        tpl := Template{pv, true, nil}
        err = tpl.Run(&buf, ctx...)
        if err != nil {
            return
        }
        ret = reflect.NewValue(buf.String())

    case *VarFunElem:
        // Argumentem jest zmienna
        ret, err = execVarFun(wr, pv, ctx, strict)
        if err != nil {
            return
        }

    default:
        panic(fmt.Sprintf(
            "tmpl:exec, line %d: Unknown parameter type!", ln))
    }
    return
}

// Okresla wartosci parametrow funkcji.
func execArgs(wr io.Writer, vf *VarFunElem, ctx []interface{}) (
        args []reflect.Value, err os.Error) {

    var arg reflect.Value
    for ii := range vf.args {
        arg, err = execParam(wr, vf.args[ii], ctx, vf.ln, true)
        if err != nil {
            return
        }
        args = append(args, arg)
    }
    return
}

// Zwraca zmienna lub nil jesli zmienna nie istnieje.
func execVarFun(wr io.Writer, vf *VarFunElem, ctx []interface{}, strict bool) (
        val reflect.Value, err os.Error) {

    var (
        path []reflect.Value
        name_id reflect.Value
    )
    // Ustalamy pelna sciezke do zmiennej w kontekscie.
    for pv := vf; pv != nil; pv = pv.next {
        // Okreslenie nazwy/indeksu zmiennej w kontekscie.
        // Jesli jest dynamiczna, dla bezpieczenstwa, do jaj okreslenia
        // uzywamy trybu strict.
        switch pe := pv.name.(type) {
        case nil:
            // Brak nazwy wiec elementem sciezki jest czyste wywolanie funkcji
            name_id = nil

        case *reflect.StringValue, *reflect.IntValue, *reflect.FloatValue:
            // Elementem sciezki jest:
            //   *reflect.StringValue: nazwa tekstowa lub indeks tekstowy,
            //   *reflect.IntValue:    indeks calkowity,
            //   *reflect.FloatValue:  indeks zmiennorzecinkowy.
            name_id = pe.(reflect.Value)

        case []Element:
            // Elementem sciezki jest indeks tekstowy.
            var buf bytes.Buffer
            tpl := Template{pe, true, nil}
            err = tpl.Run(&buf, ctx...)
            if err != nil {
                return
            }
            name_id = reflect.NewValue(buf.String())

        case *VarFunElem:
            // Elementem sciezki jest zmienna/funkcja ktora zwraca indeks
            name_id, err = execVarFun(wr, pe, ctx, true)
            if err != nil {
                return
            }

        default:
            panic(fmt.Sprintf(
                "tmpl:exec, line %d: unknown type in var/fun path!",
                vf.ln))
        }
        path = append(path, name_id)
    }

    // Poszukujemy poczatku sciezki do zmiennej przechodzac poszczegolne
    // warstwy kontekstu, rozpoczynajac od warstwy najbardziej lokalnej.
    var args []reflect.Value
    args, err = execArgs(wr, vf, ctx)
    if err != nil {
        return
    }
    name_id = path[0]
    stat := RUN_OK
    var run_error bool
    for ii := len(ctx); ii > 0; {
        ii--
        val, stat = getVarFun(reflect.NewValue(ctx[ii]), name_id, args, vf.fun)
        run_error = (stat != RUN_NOT_FOUND && stat != RUN_NIL_CTX)
        if stat == RUN_OK || run_error {
            // Znalezlismy zmienna lub wystapil blad
            break
        }
    }
    if stat != RUN_OK {
            if strict || run_error {
                return nil, RunErr{vf.ln, stat, nil}
            }
            return nil, nil
    }
    // Poczatek sciezki do zmiennej znaleziony - przechodzimy reszte.
    for _, name_id = range path[1:] {
        vf = vf.next
        args, err = execArgs(wr, vf, ctx)
        if err != nil {
            return
        }
        val, stat = getVarFun(val, name_id, args, vf.fun)
        if stat != RUN_OK {
            if strict || (stat != RUN_NOT_FOUND && stat != RUN_NIL_CTX) {
                return nil, RunErr{vf.ln, stat, nil}
            }
            return nil, nil
        }
    }
    return
}

// Public methods

func (tpl *Template) Run(wr io.Writer, ctx ...interface{}) (err os.Error){
    for _, va := range tpl.elems {
        switch el := va.(type) {
        case *TxtElem:
            _, err = io.WriteString(wr, el.txt)
            if err != nil {
                return
            }

        case *VarFunElem:
            var val reflect.Value
            val, err = execVarFun(wr, el, ctx, tpl.Strict)
            if err != nil {
                return
            }
            // Dereferencja zwroconej wartosci
            dereference(&val)
            if val == nil {
                break
            }
            switch vtn := val.Interface().(type) {
            case Template:
                // Zagniezdzony szablon
                err = vtn.Run(wr, ctx...)
                if err != nil {
                    return RunErr{el.ln, RUN_NESTED, err}
                }
            case NestedTemplate:
                // Zagniezdzony szablon z wlasnym kontekstem
                err = vtn.tpl.Run(wr, vtn.ctx...)
                if err != nil {
                    return RunErr{el.ln, RUN_NESTED, err}
                }
            default:
                // Zwykla zmienna
                if el.filt && tpl.EscapeFunc != nil {
                    // Wysylamy z wykorzystaniem EscapeFunc
                    err = tpl.EscapeFunc(wr, fmt.Sprint(vtn))
                } else {
                    // Wysylamy
                    _, err = fmt.Fprint(wr, vtn)
                }
            }

        case *IfElem:
            var v1, v2 reflect.Value
            // Parametry musza istniec jesli porownujemy je ze soba.
            v1, err = execParam(wr, el.arg1, ctx, el.ln, el.cmp != if_nocmp)
            if err != nil {
                return
            }
            var tf bool
            if el.cmp == if_nocmp {
                tf = getBool(v1)
            } else {
                v2, err = execParam(wr, el.arg2, ctx, el.ln, true)
                if err != nil {
                    return
                }
                var stat int
                tf, stat = getCmp(v1, v2, el.cmp)
                if stat != RUN_OK {
                    return RunErr{el.ln, stat, nil}
                }
            }
            // Aby wyswietlic blok if'a tworzymy kopie glownego szablonu.
            ift := *tpl
            // W kopii podmieniamy liste elementow na liste wybranego bloku.
            if tf {
                ift.elems = el.true_block
            } else {
                ift.elems = el.false_block
            }
            // Renderujemy szablon wybranego bloku.
            err = ift.Run(wr, ctx...)
            if err != nil {
                return
            }

        case *ForElem:
            var val reflect.Value
            val, err = execVarFun(wr, el.arg, ctx, false)
            if err != nil {
                return
            }
            // Tworzymy kopie glownego szablonu dla blokow for'a.
            for_tpl := *tpl
            // Dereferencja argumentu
            dereference(&val)
            // W tym momencie val moze zawierac: nil, tablice lub skalar
            switch vv := val.(type) {
            case reflect.ArrayOrSliceValue:
                // Zmienna jest typu tablicowego
                if vv.Len() != 0 {
                    // Niepusta tablica.
                    for_tpl.elems = el.iter_block
                    // Tworzymy kontekst dla iteracyjnego bloku for.
                    local_ctx := map[string]interface{}{}
                    for_ctx := append(ctx, local_ctx)
                    for ii := 0; ii < vv.Len(); ii++ {
                        ev := vv.Elem(ii)
                        if ev == nil {
                            local_ctx[el.val] = nil
                        } else {
                            local_ctx[el.val] = ev.Interface()
                        }
                        local_ctx[el.iter] = ii + el.iter_inc
                        err = for_tpl.Run(wr, for_ctx...)
                        if err != nil {
                            return
                        }
                    }
                } else {
                    // Pusta tablica
                    for_tpl.elems = el.else_block
                    err = for_tpl.Run(wr, ctx...)
                    if err != nil {
                        return
                    }
                }
            case nil:
                // nil
                for_tpl.elems = el.else_block
                err = for_tpl.Run(wr, ctx...)
                if err != nil {
                    return
                }
            default:
                // Zmienna jest skalarem, różnym od nil
                for_tpl.elems = el.iter_block
                err = for_tpl.Run(
                    wr,
                    append(
                        ctx,
                        map[string]interface{} {
                            el.val:  vv.Interface(),
                            el.iter: el.iter_inc,
                        },
                    )...
                )
                if err != nil {
                    return
                }
            }

        default:
            panic("tmpl:exec: Unknown element!")
        }
    }
    return
}

func (tpl *Template) Parse(str string) (err os.Error) {
    lnum := 1
    tpl.elems, err = parse1(&str, &lnum, "")
    if err != nil {
        return
    }
    //fmt.Println("-----------------------------------")
    //fmt.Print(revParse1(tpl.elems))
    //fmt.Println("-----------------------------------")

    tpl.elems, err = parse2(&tpl.elems, MAIN_BLK)
    //fmt.Print(revParse2(tpl.elems))
    //fmt.Println("-----------------------------------")
    return
}

func (tpl *Template) Nested(ctx ...interface{}) *NestedTemplate {
    return &NestedTemplate{tpl, ctx}
}

// Public functions

func New() *Template {
    var tpl Template
    tpl.EscapeFunc = WriteEscapedHtml
    return &tpl
}

func WriteEscapedHtml(wr io.Writer, txt string) (err os.Error) {
    var esc string
    for len(txt) > 0 {
        ii := str.IndexAny(txt, "&'<>\"")
        if ii == -1 {
            _, err = io.WriteString(wr, txt)
            return
        } else {
            _, err = io.WriteString(wr, txt[0:ii])
            if err != nil {
                return
            }
        }
        switch txt[ii] {
        case '&':
            esc = "&amp;"
        case '\'':
            esc = "&apos;"
        case '<':
            esc = "&lt;"
        case '>':
            esc = "&gt;"
        case '"':
            esc = "&quot;"
        default:
            panic("unknown escape character")
        }
        _, err = io.WriteString(wr, esc)
        txt = txt[ii+1:]
    }
    return
}
