package kasia

import (
    "reflect"
    "fmt"
    "os"
)

const (
    RUN_OK = iota
    RUN_NOT_FOUND
    RUN_WRONG_ARG_NUM
    RUN_WRONG_ARG_TYP
    RUN_INDEX_OOR
    RUN_UNK_TYPE
    RUN_NOT_FUNC
    RUN_NIL_CTX
    RUN_NESTED
    CMP_NMATCH
    CMP_ARRSLIC
    CMP_BOOL
    CMP_CHANN
    CMP_CPLX
    CMP_FUNC
    CMP_MAP
    CMP_UNKNOWN
)

type RunErr struct {
    Lnum   int
    Enum   int
    Nested os.Error
}

func (re RunErr) String() (txt string) {
    errStr := [...]string {
        "no errors",
        "variable not found",
        "wrong number of arguments",
        "wrong argument type",
        "index out of range",
        "unknown type",
        "not a function",
        "nil context",
        "nested template error",
        "can't compare, values don't match",
        "can't compare logical values",
        "can't compare arrays or slices",
        "can't compare channels",
        "can't compare complex numbers",
        "can't compare functions",
        "can't compare maps",
        "can't compare values of unknown type",
    }
    txt = fmt.Sprintf("Line %d: Runtime error: %s.", re.Lnum, errStr[re.Enum])
    if re.Nested != nil {
        txt += "\n  " + re.Nested.String()
    }
    return
}

// Pelna dereferencja wskaznikow i interfejsow
func dereference(val *reflect.Value) {
loop:
    for {
        switch vt := (*val).(type) {
        case *reflect.PtrValue:
            *val = vt.Elem()

        case *reflect.InterfaceValue:
            *val = vt.Elem()

        default:
            break loop
        }
    }
}

// Funkcja sprawdza zgodnosc typow argumentow. Jesli to konieczne konwertuje
// argumenty do typu interface{}. Jesli to potrzebna, funkcja odpowiednio
// dostosowuje args dla funkcji.
func argsMatch(ft *reflect.FuncType, args *[]reflect.Value, method int) int {
    // Liczba arguemntow akceptowanych przez funkcje/metode
    num_in := ft.NumIn() - method

    // Sprawdzamy zgodnosc liczby argumentow i obecnosc funkcji dotdotdot
    var head_args, tail_args []reflect.Value
    if ft.DotDotDot() {
        num_in--
        if len(*args) < num_in {
            return RUN_WRONG_ARG_NUM
        }
        head_args = (*args)[0:num_in]
        tail_args = (*args)[num_in:]
    } else {
        if num_in != len(*args) {
            return RUN_WRONG_ARG_NUM
        }
        head_args = *args
    }

    // Sprawdzamy zgodnosc typow poczatkowych argumentow funkcji
    for kk, av := range head_args {
        at := ft.In(kk + method)
        //fmt.Printf("DEB: ctx: %s, fun: %s\n", av.Type(), at)
        if at != av.Type() {
            // Sprawdzamy czy arguentem funkcji jest typ interface{}
            if it, ok := at.(*reflect.InterfaceType); ok && it.NumMethod()==0 {
                // Zmieniamy typ argumentu przekazywanego do funkcji
                vi := av.Interface()
                head_args[kk] = reflect.NewValue(&vi).(*reflect.PtrValue).Elem()
            } else {
                return RUN_WRONG_ARG_TYP
            }
        }
    }

    if !ft.DotDotDot() {
        return RUN_OK
    }

    // Okreslamy typ argumentów zawartych w dotdotdot
    st := ft.In(ft.NumIn() - 1).(*reflect.SliceType)
    at := st.Elem()
    it, ok := at.(*reflect.InterfaceType)
    at_is_interface := (ok && it.NumMethod() == 0)
    // Przygotowujemy miejsce na argumenty
    ddd := reflect.MakeSlice(st, len(tail_args), cap(tail_args))
    for kk, av := range tail_args {
        if at != av.Type() {
            // Sprawdzamy czy arguentem funkcji jest typ interface{}
            if at_is_interface {
                // Zmieniamy typ argumentu przekazywanego do funkcji
                vi := av.Interface()
                tail_args[kk] = reflect.NewValue(&vi).(*reflect.PtrValue).Elem()
            } else {
                return RUN_WRONG_ARG_TYP
            }
        }
        // Umieszczamy argument w ddd
        ddd.Elem(kk).SetValue(av)
    }

    // Zwracamy zmodyfikowana tablice argumentow
    *args = append(head_args, ddd)

    return RUN_OK
}


// Funkcja zwraca wartosc zmiennej o podanej nazwie lub indeksie. Jesli zmienna
// jest funkcja, wczesniej wywoluje ja z podanymi argumentami. Funkcja
// przeprowadza niezbedne dereferencje tak aby zwrocic zmienna, ktora nie jest
// ani wskaznikiem, ani interfejsem lub zwrocic wskaznik lub interface do takiej
// zmiennej. Uwaga! Funkcja moze modyfikowac wartosci args!
func getVarFun(ctx, name reflect.Value, args []reflect.Value, fun bool) (
        reflect.Value, int) {

    if ctx == nil {
        return nil, RUN_NIL_CTX
    }

    // Dereferencja nazwy
    dereference(&name)

    // Dereferencja jesli kontekst jest interfejsem
    if vi, ok := ctx.(*reflect.InterfaceValue); ok {
        ctx = vi.Elem()
    }

    // Jesli nazwa jest stringiem probujemy znalezc metode o tej nazwie
    if id, ok := name.(*reflect.StringValue); ok {
        tt := ctx.Type()
        for ii := 0; ii < tt.NumMethod(); ii++ {
            method := tt.Method(ii)
            // Sprawdzamy zgodnosc nazwy metody i typu receiver'a
            if method.Name == id.Get() && tt == method.Type.In(0) {
                stat := argsMatch(method.Type, &args, 1)
                if stat != RUN_OK {
                    return nil, stat
                }
                // Zwracamy pierwsza wartosc zwrocona przez metode
                ctx = ctx.Method(ii).Call(args)[0]
                name = nil
                args = nil
                fun = false
                break
            }
        }
    }

    // Jesli name != nil operujemy na nazwanej zmiennej z kontekstu,
    // w przeciwnym razie operujemy na samym kontekscie jako zmiennej.
    if name != nil {
        // Pelna dereferencja kontekstu
        dereference(&ctx)
        // Pobieramy wartosc
        switch vt := ctx.(type) {
        case *reflect.StructValue:
            switch id := name.(type) {
            case *reflect.StringValue:
                // Zwracamy pole struktury o podanej nazwie
                ctx = vt.FieldByName(id.Get())
                if ctx == nil {
                    return nil, RUN_NOT_FOUND
                }
            case *reflect.IntValue:
                // Zwracamy pole sruktury o podanym indeksie
                fi := int(id.Get())
                if fi < 0 || fi >= vt.NumField() {
                    return nil, RUN_INDEX_OOR
                }
                ctx = vt.Field(fi)
            default:
                return nil, RUN_NOT_FOUND
            }

        case reflect.ArrayOrSliceValue:
            switch id := name.(type) {
            case *reflect.IntValue:
                // Zwracamy element tablicy o podanym indeksie
                ei := int(id.Get())
                if ei < 0 || ei >= vt.Len() {
                    return nil, RUN_INDEX_OOR
                }
                ctx = vt.Elem(ei)
            default:
                return nil, RUN_NOT_FOUND
            }

        case *reflect.MapValue:
            kt := vt.Type().(*reflect.MapType).Key()
            if name.Type() != kt {
                if it,ok := kt.(*reflect.InterfaceType);ok && it.NumMethod()==0{
                    // Jesli mapa posiada klucz typu interface{} to
                    // przeksztalcamy indeks na typ interface{}
                    vi := name.Interface()
                    name = reflect.NewValue(&vi).(*reflect.PtrValue).Elem()
                } else {
                    return nil, RUN_NOT_FOUND
                }
            }
            ctx = vt.Elem(name)
            if ctx == nil {
                return nil, RUN_NOT_FOUND
            }

        case nil:
            return nil, RUN_NIL_CTX

        default:
            return nil, RUN_UNK_TYPE
        }
    }

    // Jesli mamy wywolanie funkcji to robimy pelna dereferencje.
    if fun {
        dereference(&ctx)
    }

    // Sprawdzenie czy ctx odnosi sie do funkcji.
    if vt, ok := ctx.(*reflect.FuncValue); ok {
        ft := vt.Type().(*reflect.FuncType)
        // Sprawdzamy zgodnosc liczby argumentow
        stat := argsMatch(ft, &args, 0)
        if stat != RUN_OK {
            return nil, stat
        }
        // Zwracamy pierwsza wrtosc zwrocaona przez funkcje.
        ctx = vt.Call(args)[0]
    } else if fun {
        return nil, RUN_NOT_FUNC
    }

    //fmt.Println("DEB getvar.ret:", reflect.Typeof(ctx.Interface()))
    return ctx, RUN_OK
}

// Funkcja zwraca wartosc logiczna argumentu
func getBool(val reflect.Value) bool {
    switch av := val.(type) {
    case nil:
        return false
    case reflect.ArrayOrSliceValue:
        return av.Len() != 0
    case *reflect.BoolValue:
        return av.Get()
    case *reflect.ChanValue:
        return !av.IsNil()
    case *reflect.ComplexValue:
        return av.Get() != cmplx(0, 0)
    case *reflect.FloatValue:
        return av.Get() != 0.0
    case *reflect.FuncValue:
        return !av.IsNil()
    case *reflect.IntValue:
        return av.Get() != 0
    case *reflect.InterfaceValue:
        return !av.IsNil()
    case *reflect.MapValue:
        return av.Len() != 0
    case *reflect.PtrValue:
        return !av.IsNil()
    case *reflect.StringValue:
        return len(av.Get()) != 0
    case *reflect.UintValue:
        return av.Get() != 0
    }
    return true
}

// Porownuje argumenty Wczesniej przeprowadza ich pelna dereferencje.
func getCmp(arg1, arg2 reflect.Value, cmp int) (tf bool, stat int) {
    dereference(&arg1)
    dereference(&arg2)
    if arg1 == nil && arg2 == nil {
        switch cmp {
        case if_eq, if_ge, if_le:
            tf = true
        default:
            tf = false
        }
    } else {
        // Sprawdzamy czy typy pasuja do siebie.
        typdif := false
        if arg1 == nil || arg2 == nil {
            typdif = true
        } else if arg1.Type() != arg2.Type() {
            typdif = true
        }
        if typdif {
            switch cmp {
            case if_eq:
                tf = false
            case if_ne:
                tf = true
            default:
                stat = CMP_NMATCH
            }
        } else {
            unk_oper := "tmpl:getCmp: Unknown comparasion operator!"
            switch av1 := arg1.(type) {
            case reflect.ArrayOrSliceValue:
                stat = CMP_ARRSLIC

            case *reflect.BoolValue:
                a1 := av1.Get()
                a2 := arg2.(*reflect.BoolValue).Get()
                switch cmp {
                case if_eq:
                    tf = (a1 == a2)
                case if_ne:
                    tf = (a1 != a2)
                default:
                    stat = CMP_BOOL
                }

            case *reflect.ChanValue:
                stat = CMP_CHANN

            case *reflect.ComplexValue:
                a1 := av1.Get()
                a2 := arg2.(*reflect.ComplexValue).Get()
                switch cmp {
                case if_eq:
                    tf = (a1 == a2)
                case if_ne:
                    tf = (a1 != a2)
                default:
                    stat = CMP_CPLX
                }

            case *reflect.FloatValue:
                a1 := av1.Get()
                a2 := arg2.(*reflect.FloatValue).Get()
                switch cmp {
                case if_eq:
                    tf = (a1 == a2)
                case if_ne:
                    tf = (a1 != a2)
                case if_lt:
                    tf = (a1 < a2)
                case if_le:
                    tf = (a1 <= a2)
                case if_gt:
                    tf = (a1 > a2)
                case if_ge:
                    tf = (a1 >= a2)
                default:
                    panic(unk_oper)
                }

            case *reflect.FuncValue:
                stat = CMP_FUNC

            case *reflect.IntValue:
                a1 := av1.Get()
                a2 := arg2.(*reflect.IntValue).Get()
                switch cmp {
                case if_eq:
                    tf = (a1 == a2)
                case if_ne:
                    tf = (a1 != a2)
                case if_lt:
                    tf = (a1 < a2)
                case if_le:
                    tf = (a1 <= a2)
                case if_gt:
                    tf = (a1 > a2)
                case if_ge:
                    tf = (a1 >= a2)
                default:
                    panic(unk_oper)
                }

            case *reflect.MapValue:
                stat = CMP_MAP

            case *reflect.StringValue:
                a1 := av1.Get()
                a2 := arg2.(*reflect.StringValue).Get()
                switch cmp {
                case if_eq:
                    tf = (a1 == a2)
                case if_ne:
                    tf = (a1 != a2)
                case if_lt:
                    tf = (a1 < a2)
                case if_le:
                    tf = (a1 <= a2)
                case if_gt:
                    tf = (a1 > a2)
                case if_ge:
                    tf = (a1 >= a2)
                default:
                    panic(unk_oper)
                }

            case *reflect.UintValue:
                a1 := av1.Get()
                a2 := arg2.(*reflect.UintValue).Get()
                switch cmp {
                case if_eq:
                    tf = (a1 == a2)
                case if_ne:
                    tf = (a1 != a2)
                case if_lt:
                    tf = (a1 < a2)
                case if_le:
                    tf = (a1 <= a2)
                case if_gt:
                    tf = (a1 > a2)
                case if_ge:
                    tf = (a1 >= a2)
                default:
                    panic(unk_oper)
                }
            default:
                stat = CMP_UNKNOWN
            }
        }
    }
    return
}
