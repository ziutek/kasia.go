package kasia

import "fmt"

const (
    PARSE_OK = iota
    PARSE_UNEXP_EOF
    PARSE_SYM_NEND
    PARSE_TXT_NEND
    PARSE_FUN_NEND
    PARSE_IF_ERR
    PARSE_BAD_NAME
    PARSE_ELSE_ERR
    PARSE_FOR_ERR
    PARSE_UNEXP_ELIF
    PARSE_UNEXP_ELSE
    PARSE_UNEXP_END
    PARSE_BAD_FLOINT
    PARSE_IF_NEND
    PARSE_FOR_NEND
)

type ParseErr struct {
    Lnum, Enum int
}

func (pe ParseErr) String() string {
    errStr := [...]string{
        "no errors",
        "unexpected end of template",
        "symbol has not close bracket",
        "there is not close `\"` or `'` for text parameter",
        "not ended function call",
        "syntax error in 'if/elif' statement",
        "bad name for variable/function",
        "syntax error in 'else' statement",
        "syntax error in 'for' statement",
        "unexpected 'elif' statement",
        "unexpected 'else' statement",
        "unexpected 'end' statement",
        "bad float/int constant",
        "not closed if/elif",
        "not closed for",
    }
    return fmt.Sprintf("Line %d: Parse error: %s.",  pe.Lnum, errStr[pe.Enum])
}
