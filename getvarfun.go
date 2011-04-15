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
    RUN_NOT_RET
    RUN_INDEX_OOR
    RUN_UNK_TYPE
    RUN_NOT_FUNC
    RUN_NIL_CTX
    RUN_INC_MAP_KEY
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
        "there isn't return value",
        "index out of range",
        "unknown type",
        "not a function",
        "nil context",
        "can't increment map key",
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
    for {
        switch val.Kind() {
        case reflect.Ptr:
            *val = val.Elem()

        case reflect.Interface:
            *val = val.Elem()

        default:
            return
        }
    }
}

// Funkcja sprawdza zgodnosc typow argumentow. Jesli to konieczne konwertuje
// argumenty do typu interface{}. Jesli to potrzebne, funkcja odpowiednio
// dostosowuje args dla funkcji.
func argsMatch(ft reflect.Type, args *[]reflect.Value, method int) int {
    if ft.NumOut() == 0 {
        return RUN_NOT_RET
    }
    // Liczba arguemntow akceptowanych przez funkcje/metode
    num_in := ft.NumIn() - method
    // Sprawdzamy zgodnosc liczby argumentow i obecnosc funkcji dotdotdot
    var head_args, tail_args []reflect.Value
    if ft.IsVariadic() {
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
            if at.Kind() == reflect.Interface && at.NumMethod() == 0 {
                // Zmieniamy typ argumentu przekazywanego do funkcji
                vi := av.Interface()
                // TODO: Sprawdzic czy da sie bez posrednictwa wskaznika
                head_args[kk] = reflect.NewValue(&vi).Elem()
            } else {
                return RUN_WRONG_ARG_TYP
            }
        }
    }

    if !ft.IsVariadic() {
        return RUN_OK
    }

    // Okreslamy typ argumentów zawartych w dotdotdot
    st := ft.In(ft.NumIn() - 1) // slice
    at := st.Elem()
    at_is_interface := (at.Kind() == reflect.Interface && at.NumMethod() == 0)
    // Przygotowujemy miejsce na argumenty
    ddd := reflect.MakeSlice(st, len(tail_args), len(tail_args))
    for kk, av := range tail_args {
        if at != av.Type() {
            // Sprawdzamy czy arguentem funkcji jest typ interface{}
            if at_is_interface {
                // Zmieniamy typ argumentu przekazywanego do funkcji
                vi := av.Interface()
                // TODO: Sprawdzic czy da sie bez posrednictwa wskaznika
                tail_args[kk] = reflect.NewValue(&vi).Elem()
            } else {
                return RUN_WRONG_ARG_TYP
            }
        }
        // Umieszczamy argument w ddd
        ddd.Index(kk).Set(av)
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
        ret reflect.Value, stat int) {

    if !ctx.IsValid() {
        stat = RUN_NIL_CTX
        return
    }

    // Dereferencja nazwy
    dereference(&name)

    // Dereferencja jesli kontekst jest interfejsem
    if ctx.Kind() == reflect.Interface {
        ctx = ctx.Elem()
    }

    // Jesli nazwa jest stringiem probujemy znalezc metode o tej nazwie
    if name.Kind() == reflect.String {
        tt := ctx.Type()
        //nm := tt.NumMethod()
        for ii := 0; ii < tt.NumMethod(); ii++ {
            method := tt.Method(ii)
            // Sprawdzamy zgodnosc nazwy metody oraz typu receiver'a
            if method.Name == name.String() && tt == method.Type.In(0) {
                stat = argsMatch(method.Type, &args, 1)
                if stat != RUN_OK {
                    return
                }
                // Zwracamy pierwsza wartosc zwrocona przez metode
                ctx = ctx.Method(ii).Call(args)[0]
                // Nie pozwalamy na dalsza analize nazwy
                name = reflect.Value{}
                // Nie pozwalamy na traktowanie zwroconej zmiennej jak funkcji
                args = nil
                fun = false
                break
            }
        }
    }

    // Jesli name zawiera wartosc operujemy na nazwanej zmiennej z kontekstu,
    // w przeciwnym razie operujemy na samym kontekscie jako zmiennej.
    if name.IsValid() {
        // Pelna dereferencja kontekstu
        dereference(&ctx)
        // Pobieramy wartosc
        switch ctx.Kind() {
        case reflect.Struct:
            switch name.Kind() {
            case reflect.String:
                // Zwracamy pole struktury o podanej nazwie
                ctx = ctx.FieldByName(name.String())
                if !ctx.IsValid() {
                    stat = RUN_NOT_FOUND
                    return
                }
            case reflect.Int:
                // Zwracamy pole sruktury o podanym indeksie
                fi := int(name.Int())
                if fi < 0 || fi >= ctx.NumField() {
                    stat = RUN_INDEX_OOR
                    return
                }
                ctx = ctx.Field(fi)
            default:
                stat = RUN_NOT_FOUND
                return
            }

        case reflect.Map:
            kt := ctx.Type().Key()
            if name.Type() != kt {
                if kt.Kind() == reflect.Interface && kt.NumMethod() == 0 {
                    // Jesli mapa posiada klucz typu interface{} to
                    // przeksztalcamy indeks na typ interface{}
                    vi := name.Interface()
                    // TODO: Sprawdzic czy da sie bez posrednictwa wskaznika
                    name = reflect.NewValue(&vi).Elem()
                } else {
                    stat = RUN_NOT_FOUND
                    return
                }
            }
            ctx = ctx.MapIndex(name)
            if !ctx.IsValid() {
                stat = RUN_NOT_FOUND
                return
            }

        case reflect.Array, reflect.Slice:
            switch name.Kind() {
            case reflect.Int:
                // Zwracamy element tablicy o podanym indeksie
                ei := int(name.Int())
                if ei < 0 || ei >= ctx.Len() {
                    stat = RUN_INDEX_OOR
                    return
                }
                ctx = ctx.Index(ei)
            default:
                stat = RUN_NOT_FOUND
                return
            }

        case reflect.Invalid:
            stat = RUN_NIL_CTX
            return

        default:
            stat = RUN_UNK_TYPE
            return
        }
    }

    // Jesli mamy wywolanie funkcji to robimy pelna dereferencje.
    if fun {
        dereference(&ctx)
    }

    // Sprawdzenie czy ctx odnosi sie do funkcji.
    if ctx.Kind() == reflect.Func {
        ft := ctx.Type()
        if fun || ft.NumIn() == 0 {
            // Sprawdzamy zgodnosc liczby argumentow
            stat = argsMatch(ft, &args, 0)
            if stat != RUN_OK {
                return
            }
            // Zwracamy pierwsza wrtosc zwrocaona przez funkcje.
            ctx = ctx.Call(args)[0]
        }
    } else if fun {
        stat = RUN_NOT_FUNC
        return
    }

    //fmt.Println("DEB getvar.ret:", reflect.Typeof(ctx.Interface()))
    return ctx, RUN_OK
}

// Funkcja zwraca wartosc logiczna argumentu
func getBool(val reflect.Value) bool {
    switch val.Kind() {
    case reflect.Invalid:
        return false
    case reflect.Array, reflect.Slice:
        return val.Len() != 0
    case reflect.Bool:
        return val.Bool()
    case reflect.Chan:
        return !val.IsNil()
    case reflect.Complex64, reflect.Complex128:
        return val.Complex() != complex(0, 0)
    case reflect.Float32, reflect.Float64:
        return val.Float() != 0.0
    case reflect.Func:
        return !val.IsNil()
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
            reflect.Int64:
        return val.Int() != 0
    case reflect.Interface:
        return !val.IsNil()
    case reflect.Map:
        return val.Len() != 0
    case reflect.Ptr:
        return !val.IsNil()
    case reflect.String:
        return len(val.String()) != 0
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
            reflect.Uint64:
        return val.Uint() != 0
    }
    return true
}

// Porownuje argumenty. Wczesniej przeprowadza ich pelna dereferencje.
func getCmp(arg1, arg2 reflect.Value, cmp int) (tf bool, stat int) {
    dereference(&arg1)
    dereference(&arg2)
    if !arg1.IsValid() && !arg2.IsValid() {
        switch cmp {
        case if_eq, if_ge, if_le:
            tf = true
        default:
            tf = false
        }
    } else {
        // Sprawdzamy czy typy pasuja do siebie.
        typdif := false
        if !arg1.IsValid() || !arg2.IsValid() {
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
            switch arg1.Kind() {
            case reflect.Array, reflect.Slice:
                stat = CMP_ARRSLIC

            case reflect.Bool:
                a1 := arg1.Bool()
                a2 := arg2.Bool()
                switch cmp {
                case if_eq:
                    tf = (a1 == a2)
                case if_ne:
                    tf = (a1 != a2)
                default:
                    stat = CMP_BOOL
                }

            case reflect.Chan:
                stat = CMP_CHANN

            case reflect.Complex64, reflect.Complex128:
                a1 := arg1.Complex()
                a2 := arg2.Complex()
                switch cmp {
                case if_eq:
                    tf = (a1 == a2)
                case if_ne:
                    tf = (a1 != a2)
                default:
                    stat = CMP_CPLX
                }

            case reflect.Float32, reflect.Float64:
                a1 := arg1.Float()
                a2 := arg2.Float()
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

            case reflect.Func:
                stat = CMP_FUNC

            case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
                    reflect.Int64:
                a1 := arg1.Int()
                a2 := arg2.Int()
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

            case reflect.Map:
                stat = CMP_MAP

            case reflect.String:
                a1 := arg1.String()
                a2 := arg2.String()
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

            case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
                    reflect.Uint64:
                a1 := arg1.Uint()
                a2 := arg2.Uint()
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
