package kasia

import (
    "os"
    "fmt"
    "reflect"
)

const (
    MAIN_BLK = iota
    IF_BLK
    FOR_BLK
    ELSE_BLK
)

func parse2Param(par *interface{}, ln int) (err os.Error) {
    switch pv := (*par).(type) {
    case reflect.Value:
        vk := pv.Kind()
        if vk == reflect.Int || vk == reflect.Int8 || vk == reflect.Int16 ||
                vk == reflect.Int32 || vk == reflect.Int64 ||
                vk == reflect.Float32 || vk == reflect.Float64 {
            // Argumentem jest liczba - nic nie robimy.
        } else {
            panic(fmt.Sprintf("tmpl:parse2, line %d: Unknown parameter type: " +
                "%s!", ln, vk))
        }

    case []Element:
        // Argumentem jest wstepnie sparsowany tekst
        *par, err = parse2(&pv, MAIN_BLK)

    case *VarFunElem:
        // Argumentem jest zmienna
        err = parse2VarFun(pv)
        *par = pv

    default:
        panic(fmt.Sprintf("tmpl:parse2, line %d: Unknown parameter type: %T!",
            ln, pv))
    }
    return
}

// Parsuje w miejscu element typu VarFunElem
func parse2VarFun(vf *VarFunElem) (err os.Error) {
    // Parsujemy sciezke do zmiennej/funkcji
    for vf != nil {
        switch pe := vf.name.(type) {
        case nil:
            // Samowywolanie - nic nie robimy.
        case reflect.Value:
            vk := pe.Kind()
            if vk == reflect.String || vk == reflect.Int ||
                    vk == reflect.Int8 || vk == reflect.Int16 ||
                    vk == reflect.Int32 || vk == reflect.Int64 ||
                    vk == reflect.Float32 || vk == reflect.Float64 {
                // Nazwa, indeks liczbowy - nic nie robimy.
            } else {
                panic(fmt.Sprintf("tmpl:parse2, line %d: Unknown type (%s) " +
                    "of index in var/fun path! ", vf.ln, vk))
            }

        case []Element:
            // Indeks tekstowy po sparsowaniu.
            vf.name, err = parse2(&pe, MAIN_BLK)
            if err != nil {
                return
            }

        case *VarFunElem:
            // Indeks bedacy wynikiem funkcji
            err = parse2VarFun(pe)
            if err != nil {
                return
            }
            vf.name = pe

        default:
            panic(fmt.Sprintf(
                "tmpl:parse2, line %d: Unknown type (%T) in var/fun path!",
                vf.ln, pe))
        }

        // Parsujemy argumenty funkcji
        for ii := range vf.args {
            err = parse2Param(&vf.args[ii], vf.ln)
            if err != nil {
                return
            }
        }
        vf = vf.next
    }
    return
}

// Zwraca pierwszy element i odpowiednio skraca wejsciawa tablice elementow.
func parse2FirstElem(in_els *[]Element) (out_el Element, err os.Error) {
    va := (*in_els)[0]
    *in_els = (*in_els)[1:]

    switch el := va.(type) {
    case *TxtElem:
        out_el = el

    case *VarFunElem:
        err = parse2VarFun(el)
        if err != nil {
            return
        }
        out_el = el

    case *IfElem:
        err = parse2Param(&el.arg1, el.ln)
        if err != nil {
            return
        }
        if el.cmp != if_nocmp {
            err = parse2Param(&el.arg2, el.ln)
            if err != nil {
                return
            }
        }
        el.true_block, err = parse2(in_els, IF_BLK)
        if err != nil {
            return
        }
        if len(*in_els) == 0 {
            // Nie znaleziono elif, else lub end
            err = ParseErr{el.ln, PARSE_IF_NEND}
            return
        }
        // Tutaj (*in_els)[0] moze zawierac elif, else lub end 
        switch stop_el := (*in_els)[0].(type) {
        case *EndElem:
            *in_els = (*in_els)[1:] // Usowamy end
        case *ElseElem:
            *in_els = (*in_els)[1:] // Usowamy else
            el.false_block, err = parse2(in_els, ELSE_BLK)
            if err != nil {
                return
            }
            *in_els = (*in_els)[1:] // Usowamy end
        case *ElifElem:
            // Zmieniamy elif w if
            (*in_els)[0] = &IfElem{
                stop_el.ln, stop_el.cmp, stop_el.arg1, stop_el.arg2, nil, nil,
            }
            el.false_block = []Element{nil}
            el.false_block[0], err = parse2FirstElem(in_els)
            if err != nil {
                return
            }
        default:
            panic(fmt.Sprintf(
                "tmpl:parse2, line %d: Unknown element after 'if'!", el.ln))
        }
        out_el = el

    case *ForElem:
        err = parse2VarFun(el.arg)
        if err != nil {
            return
        }
        el.iter_block, err = parse2(in_els, FOR_BLK)
        if err != nil {
            return
        }
        if len(*in_els) == 0 {
            // Nie znaleziono else lub end
            err = ParseErr{el.ln, PARSE_FOR_NEND}
            return
        }
        // Tutaj (*in_els)[0] moze zawierac else lub end 
        switch stop_el := (*in_els)[0].(type) {
        case *EndElem:
            *in_els = (*in_els)[1:] // Usowamy end
        case *ElseElem:
            *in_els = (*in_els)[1:] // Usowamy else
            el.else_block, err = parse2(in_els, ELSE_BLK)
            if err != nil {
                return
            }
            *in_els = (*in_els)[1:] // Usowamy end
        default:
            panic(fmt.Sprintf(
                "tmpl:parse2, line %d (%d): Unknown element after 'for'!",
                el.ln, el.lnum()))
        }
        out_el = el

    default:
        panic(fmt.Sprintf(
            "tmpl:parse2: line ? (%d) Unknown element!", el.lnum()))
    }
    return
}

func parse2(in *[]Element, blk_type int) (out []Element, err os.Error) {
    for len(*in) > 0 {
        switch el := (*in)[0].(type) {
        case *ElifElem:
            if blk_type != IF_BLK {
                err = ParseErr{el.ln, PARSE_UNEXP_ELIF}
            }
            return

        case *ElseElem:
            if blk_type != FOR_BLK && blk_type != IF_BLK {
                err = ParseErr{el.ln, PARSE_UNEXP_ELSE}
            }
            return

        case *EndElem:
            if blk_type == MAIN_BLK {
                err = ParseErr{el.ln, PARSE_UNEXP_END}
            }
            return

        default:
            // Pozostale elementy parsujemy pojedynczo i dodajemy do wyjscia
            el, err = parse2FirstElem(in)
            if err != nil {
                return
            }
            out = append(out, el)
        }
    }
    return
}
