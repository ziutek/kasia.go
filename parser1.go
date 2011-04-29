package kasia

import (
    "strings"
    "unicode"
    "strconv"
    "os"
    "reflect"
)

func findChar(txt []byte, chars string, lnum *int) int {
    for ii, cc := range txt {
        if strings.IndexRune(chars, int(cc)) != -1 {
            return ii
        }
        if cc == '\n' {
            *lnum++
        }
    }
    return -1
}

func findNotChar(txt []byte, chars string, lnum *int) int {
    for ii, cc := range txt {
        if strings.IndexRune(chars, int(cc)) == -1 {
            return ii
        }
        if cc == '\n' {
            *lnum++
        }
    }
    return -1
}

func findCharEsc(txt []byte, chars string, lnum *int) int {
    index := 0
    for {
        id := findChar(txt, chars, lnum)
        if id == -1 {
            return -1
        }
        index += id
        // Liczymy znaki $ ktore bezposrednio poprzedzaja znaleziony znak.
        dnum := 0
        for ; dnum < id; dnum++ {
            if txt[id-dnum-1] != '$' {
                break
            }
        }
        if dnum & 1 == 0 {
            // Jesli liczba znakow $ jest parzysta to znalezlismy znak.
            break
        }
        txt = txt[id+1:]
        index++
    }
    return index
}

const (
    seps_newline = "\n\r"
    seps_white   = seps_newline + " \t"
    seps_nnum    = seps_white + "~!@#$%^&*={}|\\,;<>?/()[]:`\"'"
    seps_num     = "+-."
    seps         = seps_nnum + seps_num
    digits       = "1234567890"
    not_letters  = seps + digits
)

func skip(frag *[]byte, lnum *int, seps string) os.Error {
    ii := findNotChar(*frag, seps, lnum)
    if ii == -1 {
        return ParseErr{*lnum, PARSE_UNEXP_EOF}
    }
    *frag = (*frag)[ii:]
    return nil
}

func skipWhite(frag *[]byte, lnum *int) os.Error {
    return skip(frag, lnum, seps_white)
}

func skipNewline(frag *[]byte, lnum *int) {
    nl := false
    for len(*frag) != 0 {
        switch (*frag)[0] {
        case '\n':
            if nl {
                return
            }
            *lnum++
            nl = true
        case '\r':
        default:
            return
        }
        *frag = (*frag)[1:]
    }
}

// Funkcja pozwala na uniknac przechowywanie w sparsowanym szabloni
// powtarzajacych sie symboli. Jesli symbol pojawil sie juz wczenieniej w
// trakcie parsowania to jest zwracany przez ta funckje.
func getSymbol(sym string, stab map[string]string) string {
    if s, ok := stab[sym]; ok {
        sym = s
    } else {
        stab[sym] = sym
    }
    return sym
}

// Parsuje parametr znajdujacy sie na poczatku frag. Pomija biale znaki przed
// parametrem i usuwa za nim (po wywolaniu frag[0] nie jest bialym znakiem).
func parse1Param(txt *[]byte, lnum *int, stab map[string]string) (
        par interface{}, err os.Error) {
    frag := *txt

    // Pomijamy biale znaki, ktore moga wystapic przed parametrem.
    err = skipWhite(&frag, lnum)
    if err != nil {
        return
    }

    if frag[0] == '"' || frag[0] == '\'' || frag[0] == '`' {
        // Parametrem jest tekst.
        txt_sep := string(frag[0:1])
        frag = frag[1:]
        if len(frag) == 0 {
            err = ParseErr{*lnum, PARSE_UNEXP_EOF}
            return
        }
        // Parsujemy tekst parametru. 
        par, err = parse1(&frag, lnum, stab, txt_sep)
        if err != nil {
            return
        }
    } else if unicode.IsDigit(int(frag[0])) ||
            strings.IndexRune(seps_num, int(frag[0])) != -1 {
        // Parametrem jest liczba
        ii := findChar(frag, seps_nnum, lnum)
        if ii == -1 {
            err = ParseErr{*lnum, PARSE_UNEXP_EOF}
            return
        }
        var iv int
        sn := string(frag[0:ii])
        iv, err = strconv.Atoi(sn)
        if err != nil {
            var fv float64
            fv, err = strconv.Atof64(sn)
            if err != nil {
                err = ParseErr{*lnum, PARSE_BAD_FLOINT}
                return
            }
            par = reflect.ValueOf(fv)
        } else {
            par = reflect.ValueOf(iv)
        }
        frag = frag[ii:]
    } else {
        par, err = parse1VarFun("", &frag, lnum, stab, false)
        if err != nil {
            return
        }
    }

    // Pomijamy biale znaki, ktore moga wystapic po parametrze.
    err = skipWhite(&frag, lnum)

    *txt = frag
    return
}

func parse1VarFun(symbol string, txt *[]byte, lnum *int, stab map[string]string,
        filt bool) (el *VarFunElem, err os.Error) {

    el = &VarFunElem{}
    el.ln = *lnum
    el.filt = filt
    frag := *txt

    if len(symbol) == 0 {
        // Poszukujemy nazwy zmiennej/funkcji
        if frag[0] == '[' {
            // Dynamiczna nazwa w postaci indeksu
            frag = frag[1:]
            el.name, err = parse1Param(&frag, lnum, stab)
            if err != nil {
                return
            }
            // parse1Param() usunela biale znaki, wiec musi teraz wystapic ']'.
            if frag[0] != ']' {
                err = ParseErr{*lnum, PARSE_BAD_NAME}
                return
            }
            frag = frag[1:]
        } else if frag[0] == '@' {
            // Bezposrednie odwolanie do samego kontekstu
            el.name = nil
            frag = frag[1:]
        } else if frag[0] == '(' {
            // Bezposrednie wywolanie funkcji na samym kontekscie
            el.name = nil
        } else if strings.IndexRune(not_letters, int(frag[0])) == -1 {
            // Nazwa statyczna
            ii := findChar(frag, seps, lnum)
            if ii == -1 {
                ii = len(frag)
            }
            el.name = reflect.ValueOf(getSymbol(string(frag[0:ii]), stab))
            frag = frag[ii:]
        } else {
            err = ParseErr{*lnum, PARSE_BAD_NAME}
            return
        }
    } else {
        el.name = reflect.ValueOf(getSymbol(symbol, stab))
    }

    // Sprawdzamy czy sa argumenty
    if len(frag) != 0 && frag[0] == '(' {
        el.fun = true
        var arg interface{}
        for {
            // Pomijamy '(' lub ','
            frag = frag[1:]
            err = skipWhite(&frag, lnum)
            if err != nil {
                return
            }
            // Sprawdzamy kolejny znak.
            if frag[0] == ')' {
                break
            }

            // Pobieramy parametr usuwajac otaczajace go biale znaki.
            arg, err = parse1Param(&frag, lnum, stab)
            if err != nil {
                return
            }
            // Dodajemy argument do tabeli argumentow.
            el.args = append(el.args, arg)

            // Sprawdzamy kolejny znak po argumencie i bialych znakach za nim.
            if frag[0] == ')' {
                break
            }
            if frag[0] != ',' {
                err = ParseErr{*lnum, PARSE_FUN_NEND}
                return
            }
        }
        frag = frag[1:] // Pomijamy ')'
    }

    // Jesli to nie jest koniec sciezki dodajemy kolejna zmienna
    if len(frag) > 1 && frag[0] == '.' &&
            strings.IndexRune(not_letters, int(frag[1])) == -1 {
        // Kropka po ktorej nastepuje litera oznacza kolejny element sciezki.
        frag = frag[1:]
        el.next = el // Oznaczamy ze jest nastepna zmienna
    } else if len(frag) > 2 && (frag[0] == '[' || frag[0] == '(') {
        // Na nastepnej pozycji jest indeks lub samowywolanie.
        el.next = el // Oznaczamy ze jest nastepna zmienna
    }
    if el.next != nil {
        // Ustawiamy znacznik filtrowania we wszystkich elementach sciezki,
        // choc potrzebny bedzie zapewne tylko w pierwszym (moze w ostatnim?).
        el.next, err = parse1VarFun("", &frag, lnum, stab, filt)
        if err != nil {
            return
        }
    }
    *txt = frag
    return
}

func parse1(txt *[]byte, lnum *int, stab map[string]string, txt_sep string) (
        elems []Element, err os.Error) {

    hot_chars := "$" + txt_sep
    frag := *txt
    for {
        ii := findChar(frag, hot_chars, lnum)
        if ii == -1 {
            // Nie znaleziono znacznika
            if len(txt_sep) != 0 {
                // Powinien byc choc znacznik z txt_sep
                err = ParseErr{*lnum, PARSE_TXT_NEND}
                return
            }
            ii = len(frag)
        }

        if ii != 0 {
            elems = append(elems, &TxtElem{*lnum, frag[0:ii]})
        }

        if ii == len(frag) {
            // Koniect tekstu
            break
        }
        if frag[ii] != '$' {
            // Koniec tekstu ze wzgledu na txt_sep
            frag = frag[ii+1:]
            break
        }

        frag = frag[ii+1:]
        if len(frag) == 0 {
            err = ParseErr{*lnum, PARSE_UNEXP_EOF}
            return
        }

        // Wyluskujemy symbol znajdujacy sie po znaku $

        filtered := true

        if frag[0] == ':' {
            // Znacznik braku filtracji.
            filtered = false
            frag = frag[1:]
            if len(frag) == 0 {
                err = ParseErr{*lnum, PARSE_UNEXP_EOF}
                return
            }
        }

        bracketed := false

        if frag[0] == '{' {
            // Symbol zamkniety w nawiasy klamrowe
            bracketed = true
            frag = frag[1:]
            if len(frag) == 0 {
                err = ParseErr{*lnum, PARSE_UNEXP_EOF}
                return
            }
        }

        ii = findChar(frag, seps, lnum)
        if ii == -1 {
            ii = len(frag)
        }

        symbol := string(frag[0:ii])
        frag = frag[ii:]

        // W tym momencie symbol zawiera nazwe symbolu,
        // frag zawiera tekst bezposrednio za sybolem.

        switch symbol {
        case "if", "elif":
            // Parsujemy zmienna lub funkcje.
            var arg1, arg2 interface{}
            arg1, err = parse1Param(&frag, lnum, stab)
            if err != nil {
                return
            }
            cmp := if_nocmp
            if frag[0] != ':' {
                //If z porownaniem argumentow.
                ii = findNotChar(frag, "=!<>", lnum)
                if ii == -1 {
                    err = ParseErr{*lnum, PARSE_UNEXP_EOF}
                    return
                }
                switch string(frag[0:ii]) {
                case "==":
                    cmp = if_eq
                    frag = frag[2:]
                case "!=":
                    cmp = if_ne
                    frag = frag[2:]
                case "<":
                    cmp = if_lt
                    frag = frag[1:]
                case "<=":
                    cmp = if_le
                    frag = frag[2:]
                case ">":
                    cmp = if_gt
                    frag = frag[1:]
                case ">=":
                    cmp = if_ge
                    frag = frag[2:]
                default:
                    err = ParseErr{*lnum, PARSE_IF_ERR}
                    return
                }
                arg2, err = parse1Param(&frag, lnum, stab)
                if err != nil {
                    return
                }
            }

            // Wyrazenie konczy sie dwukropkiem.
            if frag[0] != ':' {
                err = ParseErr{*lnum, PARSE_IF_ERR}
                return
            }
            frag = frag[1:]
            skipNewline(&frag, lnum)

            if symbol == "if" {
                elems = append(elems, &IfElem{*lnum, cmp, arg1, arg2, nil, nil})
            } else {
                elems = append(elems, &ElifElem{*lnum, cmp, arg1, arg2})
            }


        case "for":
            err = skipWhite(&frag, lnum)
            if err != nil {
                return
            }
            // Okreslamy zmienna iteracyjna
            ii = findChar(frag, seps, lnum)
            if ii == -1 {
                err = ParseErr{*lnum, PARSE_UNEXP_EOF}
                return
            }
            iter := getSymbol(string(frag[0:ii]), stab)
            if len(iter) == 0 {
                err = ParseErr{*lnum, PARSE_FOR_ERR}
                return
            }
            frag = frag[ii:]
            inc := 0
            if frag[0] == '+' {
                // Zmiena iteracyjna moze byc powiekszona o 1
                inc = 1
                frag = frag[1:]
            }
            // Pomijamy ewentualne biale znaki za zmienna iteracyjna
            err = skipWhite(&frag, lnum)
            if err != nil {
                return
            }
            if frag[0] != ',' {
                err = ParseErr{*lnum, PARSE_FOR_ERR}
                return
            }
            frag = frag[1:]
            // Pomijamy ewentualne biale znaki za przecinkiem
            err = skipWhite(&frag, lnum)
            if err != nil {
                return
            }
            // Okreslamy zmienna zawierajaca wartosc dladanej iteracji
            ii = findChar(frag, seps, lnum)
            if ii == -1 {
                err = ParseErr{*lnum, PARSE_UNEXP_EOF}
                return
            }
            val := getSymbol(string(frag[0:ii]), stab)

            frag = frag[ii:]
            // Pomijamy biale znaki przed "in"
            err = skipWhite(&frag, lnum)
            if err != nil {
                return
            }
            ii = findChar(frag, seps, lnum)
            if ii == -1 {
                err = ParseErr{*lnum, PARSE_UNEXP_EOF}
                return
            }
            if string(frag[0:ii]) != "in" {
                err = ParseErr{*lnum, PARSE_FOR_ERR}
                return
            }
            frag = frag[ii:]
            // Pomijamy biale znaki za "in"
            err = skipWhite(&frag, lnum)
            if err != nil {
                return
            }
            // Parsujemy zmienna lub funkcje tablicowa.
            var vf *VarFunElem
            vf, err = parse1VarFun("", &frag, lnum, stab, false)
            if err != nil {
                return
            }
            // Pomijamy ewentualne biale znaki za zmienna tablicowa.
            err = skipWhite(&frag, lnum)
            if err != nil {
                return
            }
            // Wyrazenie konczy sie dwukropkiem.
            if frag[0] != ':' {
                err = ParseErr{*lnum, PARSE_FOR_ERR}
                return
            }
            frag = frag[1:]
            skipNewline(&frag, lnum)

            elems = append(elems, &ForElem{*lnum, inc, iter, val, vf, nil, nil})

        case "else":
            // Pomijamy ewentualne biale znaki
            err = skipWhite(&frag, lnum)
            if err != nil {
                return
            }
            if frag[0] != ':' {
                err = ParseErr{*lnum, PARSE_ELSE_ERR}
                return
            }
            frag = frag[1:]
            skipNewline(&frag, lnum)

            elems = append(elems, &ElseElem{*lnum})

        case "end":
            elems = append(elems, &EndElem{*lnum})
            skipNewline(&frag, lnum)

        case "":
            if len(frag) == 0 {
                err = ParseErr{*lnum, PARSE_UNEXP_EOF}
                return
            }
            switch frag[0] {
            case '@', '[', '(':
                // Bezposrednie odwolanie do samego kontekstu, zmienna
                // o dynamicznej lub liczbowej nazwie lub samowywolanie.
                var vf *VarFunElem
                vf, err = parse1VarFun("", &frag, lnum, stab, filtered)
                if err != nil {
                    return
                }
                elems = append(elems, vf)

            case '$', '"', '\'', '`':
                // Escape sequence
                elems = append(elems, &TxtElem{*lnum, frag[0:1]})
                frag = frag[1:]

            case '#':
                // Komentarz.
                frag = frag[1:]
                for {
                    ii = findChar(frag, "#", lnum)
                    if ii == -1 || ii == len(frag)-1 {
                        err = ParseErr{*lnum, PARSE_UNEXP_EOF}
                        return
                    }
                    frag = frag[ii+1:]
                    if frag[0] == '$' {
                        break
                    }
                }
                frag = frag[1:]
                skipNewline(&frag, lnum)

            default:
                err = ParseErr{*lnum, PARSE_BAD_NAME}
                return
            }

        default: // Zmienna lub funkcja
            var vf *VarFunElem
            vf, err = parse1VarFun(symbol, &frag, lnum, stab, filtered)
            if err != nil {
                return
            }
            elems = append(elems, vf)
        }

        // Jesli symbol rozpoczynał się od klamry, zajmujemy sie jejzamknieciem
        if bracketed {
            if len(frag) == 0 || frag[0] != '}' {
                err = ParseErr{*lnum, PARSE_SYM_NEND}
                return
            }
            frag = frag[1:]
        }
    }
    *txt = frag
    return
}
