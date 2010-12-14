package kasia

import (
    "fmt"
    "reflect"
)

var cmp_map map[int]string

func init() {
    cmp_map = map[int]string{
        if_eq:    "==",
        if_ne:    "!=",
        if_lt:    "<",
        if_le:    "<=",
        if_gt:    ">",
        if_ge:    ">=",
    }
}

func revParse1Param(par interface{}) (ret string) {
    switch pv := par.(type) {
    case reflect.Value:
        ret += fmt.Sprint(pv.Interface())

    case []Element:
        ret += "\"" + revParse1(pv) + "\""

    case *VarFunElem:
        ret += revParse1VarFun(pv)

    default:
        ret += "!UNKNOWN_ARG!"
    }
    return
}

func revParse1VarFun(vf *VarFunElem) (ret string) {
    show_dot := false
    for vf != nil {
        switch pe := vf.name.(type) {
        case *reflect.StringValue:
            if show_dot {
                ret += "."
            }
            ret += pe.Get()

        case reflect.Value:
            ret += fmt.Sprintf("[%v]", pe.Interface())

        case nil:
            // Samowywolanie

        case []Element:
            ret += "[\"" + revParse1(pe) + "\"]"

        case *VarFunElem:
            ret += fmt.Sprintf("[%s]", revParse1VarFun(pe))

        default:
            ret += "!UNKNOWN_NAME!"
        }
        show_dot = true

        if vf.fun {
            ret += "("
            for ii := range vf.args {
                if ii != 0 {
                    ret += ", "
                }
                ret += revParse1Param(vf.args[ii])
            }
            ret += ")"
        }

        vf = vf.next
    }
    return
}

func revParse1(elems []Element) (ret string){
    for _, va := range elems {
        switch el := va.(type) {
        case *TxtElem:
            ret += el.txt

        case *VarFunElem:
            ret += "$"
            if !el.filt {
                ret += ":"
            }
            ret += "{" + revParse1VarFun(el)  + "}"

        case *IfElem:
            ret += "$if " + revParse1Param(el.arg1)
            if el.cmp != if_nocmp {
                ret += " "
                cmp, ok := cmp_map[el.cmp]
                if ok {
                    ret += cmp
                } else {
                    ret += "!UNKNOWN_CMP!"
                }
                ret += " " + revParse1Param(el.arg2)
            }
            ret +=  ":"

        case *ElifElem:
            ret += "$elif " + revParse1Param(el.arg1)
            if el.cmp != if_nocmp {
                ret += " "
                cmp, ok := cmp_map[el.cmp]
                if ok {
                    ret += cmp
                } else {
                    ret += "!UNKNOWN_CMP!"
                }
                ret += " " + revParse1Param(el.arg2)
            }
            ret +=  ":"

        case *ElseElem:
            ret += "$else:"

        case *EndElem:
            ret += "${end}"

        case *ForElem:
            inc := ""
            if el.iter_inc == 1 {
                inc = "+"
            }
            ret += fmt.Sprintf("$for %s%s, %s in %s:", el.iter, inc, el.val,
                revParse1VarFun(el.arg))

        default:
            ret += "!UNKNOWN ELEMENT!"
        }
    }
    return
}

func revParse2Param(par interface{}) (ret string) {
    switch pv := par.(type) {
    case reflect.Value:
        ret += fmt.Sprint(pv.Interface())

    case []Element:
        ret += "\"" + revParse2(pv) + "\""

    case *VarFunElem:
        ret += revParse2VarFun(pv)

    default:
        ret += "!UNKNOWN_ARG!"
    }
    return
}


func revParse2VarFun(vf *VarFunElem) (ret string) {
    show_dot := false
    for vf != nil {
        switch pe := vf.name.(type) {
        case *reflect.StringValue:
            if show_dot {
                ret += "."
            }
            ret += pe.Get()

        case reflect.Value:
            ret += fmt.Sprintf("[%v]", pe.Interface())

        case nil:
            // Samowywolanie

        case []Element:
            ret += "[\"" + revParse2(pe) + "\"]"

        case *VarFunElem:
            ret += fmt.Sprintf("[%s]", revParse2VarFun(pe))

        default:
            ret += "!UNKNOWN_NAME!"
        }
        show_dot = true

        if vf.fun {
            ret += "("
            for ii := range vf.args {
                if ii != 0 {
                    ret += ", "
                }
                ret += revParse2Param(vf.args[ii])
            }
            ret += ")"
        }

        vf = vf.next
    }
    return
}

func revParse2(elems []Element) (ret string) {
    for _, va := range elems {
        switch el := va.(type) {
        case *TxtElem:
            ret += el.txt

        case *VarFunElem:
            ret += "$"
            if !el.filt {
                ret += ":"
            }
            ret += "{" + revParse2VarFun(el)  + "}"

        case *IfElem:
            ret += "$if " + revParse2Param(el.arg1)
            if el.cmp != if_nocmp {
                ret += " "
                cmp, ok := cmp_map[el.cmp]
                if ok {
                    ret += cmp
                } else {
                    ret += "!UNKNOWN_CMP!"
                }
                ret += " " + revParse2Param(el.arg2)
            }
            ret +=  ":"
            ret += revParse2(el.true_block)
            if el.false_block != nil {
                ret += "$else:" + revParse2(el.false_block)
            }
            ret += "${end}"

        case *ForElem:
            inc := ""
            if el.iter_inc == 1 {
                inc = "+"
            }
            ret += fmt.Sprintf("$for %s%s, %s in %s:", el.iter, inc, el.val,
                revParse2VarFun(el.arg))
            ret += revParse2(el.iter_block)
            if el.else_block != nil {
                ret += "$else:" + revParse2(el.else_block)
            }
            ret += "${end}"

        default:
            ret += "!UNKNOWN ELEMENT!"
        }
    }
    return
}
