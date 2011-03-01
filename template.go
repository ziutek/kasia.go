package kasia

import (
    "os"
    "io"
    "fmt"
    "bytes"
    "reflect"
)

type Template struct {
    elems      []Element
    Strict     bool
    EscapeFunc func(io.Writer, []byte) os.Error
}

type NestedTemplate struct {
    tpl *Template
    ctx []interface{}
}

type ContextItself []interface{}

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
        // Jesli jest dynamiczna, dla bezpieczenstwa, do jej okreslenia
        // uzywamy trybu strict.
        switch pe := pv.name.(type) {
        case nil:
            // Brak nazwy wiec elementem sciezki jest czyste wywolanie funkcji
            // lub sam kontekst.
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

    args, err := execArgs(wr, vf, ctx)
    if err != nil {
        return
    }
    stat := RUN_OK
    name_id = path[0]
    // name_id == nil tylko wtedy gdy na poczatku sciezki jest sam kontekst '$@'
    // lub czyste wywolanie funkcji '$(...)'
    if name_id != nil || vf.fun {
        // Poszukujemy zmiennej lub funkcji pasujacej do poczatku sicezki
        for ii := len(ctx); ii > 0; {
            ii--
            val, stat = getVarFun(
                reflect.NewValue(ctx[ii]), name_id, args, vf.fun,
            )
            if stat == RUN_OK {
                // Znalezlismy zmienna
                break
            }
            // Blad oznacza ze dana skladowa kontekstu nie pasuje do atrybutu,
            // czyli wystepuje jeden z ponizszych bledow:
            // - skladowa nie zawiera atrybutu o podanej nazwie,
            // - skladowa nie zawiera atrybutu o podanym indeksie,
            // - skladowa jest rowna nil,
            // - skladowa mie pasuje do zadanego wywolania funkcj (nie jest
            //   funkcja lub nie pasuje pod wzgledem argumentow).
            // Na obecna chwile getVarFun nie zwraca innych bledow.
        }
        if stat != RUN_OK {
            if strict {
                return nil, RunErr{vf.ln, stat, nil}
            }
            return nil, nil
        }
    } else {
        // Poczatek sciezki wskazuje na sam wielowarstwowy kontekst.
        // Traktujemy go jako normalna wartosc typu slice
        val = reflect.NewValue(ContextItself(ctx))
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
            _, err = wr.Write(el.txt)
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
            case []byte:
                // Raw text
                if el.filt && tpl.EscapeFunc != nil {
                    err = tpl.EscapeFunc(wr, vtn)
                } else {
                    _, err = wr.Write(vtn)
                }
            default:
                // Inna wartosc
                if el.filt && tpl.EscapeFunc != nil {
                    // Wysylamy z wykorzystaniem EscapeFunc
                    var buf bytes.Buffer
                    fmt.Fprint(&buf, vtn)
                    err = tpl.EscapeFunc(wr, buf.Bytes())
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
            switch vv := val.(type) {
            case reflect.ArrayOrSliceValue:
                vv_len := vv.Len()
                if vv_len != 0 {
                    for_tpl.elems = el.iter_block
                    // Tworzymy kontekst dla iteracyjnego bloku for
                    var val interface{}
                    iter := el.iter_inc
                    local_ctx := map[string]interface{}{
                        el.iter: &iter,
                        el.val:  &val,
                    }
                    for_ctx := append(ctx, local_ctx)
                    for vv_len += el.iter_inc; iter < vv_len; iter++ {
                        ev := vv.Elem(iter - el.iter_inc)
                        if ev == nil {
                            val = nil
                        } else {
                            val = ev.Interface()
                        }
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
            case *reflect.MapValue:
                if vv.Len() != 0 {
                    if el.iter_inc != 0 {
                        return RunErr{el.ln, RUN_INC_MAP_KEY, nil}
                    }
                    for_tpl.elems = el.iter_block
                    // Tworzymy kontekst dla iteracyjnego bloku for
                    var key, val interface{}
                    local_ctx := map[string]interface{}{
                        el.iter: &key,
                        el.val:  &val,
                    }
                    for_ctx := append(ctx, local_ctx)
                    for _, k := range vv.Keys() {
                        ev := vv.Elem(k)
                        if ev == nil {
                            val = nil
                        } else {
                            val = ev.Interface()
                        }
                        key = k.Interface()
                        err = for_tpl.Run(wr, for_ctx...)
                        if err != nil {
                            return
                        }
                    }
                } else {
                    // Pusty slownik
                    for_tpl.elems = el.else_block
                    err = for_tpl.Run(wr, ctx...)
                    if err != nil {
                        return
                    }
                }
             case *reflect.ChanValue:
                for_tpl.elems = el.iter_block
                // Tworzymy kontekst dla iteracyjnego bloku for
                var val interface{}
                iter := el.iter_inc
                local_ctx := map[string]interface{}{
                    el.iter: &iter,
                    el.val:  &val,
                }
                for_ctx := append(ctx, local_ctx)
                for {
                    ev := vv.Recv()
                    if vv.Closed() {
                        break
                    }
                    if ev == nil {
                        val = nil
                    } else {
                        val = ev.Interface()
                    }
                    err = for_tpl.Run(wr, for_ctx...)
                    if err != nil {
                        return
                    }
                    iter++
                }
                if iter == el.iter_inc {
                    // Nic nie odebralismy z kanalu
                    for_tpl.elems = el.else_block
                    err = for_tpl.Run(wr, ctx...)
                    if err != nil {
                        return
                    }
                }
           case nil:
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
                            el.iter: nil,
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

func (tpl *Template) Parse(txt []byte) (err os.Error) {
    lnum := 1
    tpl.elems, err = parse1(&txt, &lnum, make(map[string]string), "")
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

var (
    esc_amp  = []byte("&amp;")
    esc_apos = []byte("&apos;")
    esc_lt   = []byte("&lt;")
    esc_gt   = []byte("&gt;")
    esc_quot = []byte("&quot;")
)

func WriteEscapedHtml(wr io.Writer, txt []byte) (err os.Error) {
    var esc []byte
    last := 0
    for ii, bb := range txt {
        switch bb {
        case '&':
            esc = esc_amp
        case '\'':
            esc = esc_apos
        case '<':
            esc = esc_lt
        case '>':
            esc = esc_gt
        case '"':
            esc = esc_quot
        default:
            continue
        }
        if _, err = wr.Write(txt[last:ii]); err != nil {
            return
        }
        if _, err = wr.Write(esc); err != nil {
            return
        }
        last = ii + 1
    }
    _, err = wr.Write(txt[last:])
    return
}
