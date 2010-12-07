Sorry for my poor English :(

# FAQ

## What is Kasia.go?

Kasia.go is Go implementation of Kasia templating system.

Kasia is primarily designed for HTML, but you can use it for anything you want.
Its syntax is somewhat similar to web.py templating system - Templator, but much
much simpler.

## What means "Kasia"?

Kasia is my daughter's name (Polish equivalent of Katie or Kathy).

## How to use it?

Kasia's native interface contains one function and three methods:

    func New() *Template
    func (*Template) Parse(str string) os.Error
    func (*Template) Run(wr io.Writer, ctx ...interface{}) os.Error
    func (*Template) Nested(ctx interface{}) *NestedTemplate 

The simplest example is:

    package main

    import (
        "kasia"
        "os"
        "fmt"
    )

    type Ctx struct {
        h, w string
    }

    func main() {
        ctx := &Ctx{"Hello", "world"}

        tpl := kasia.New()
        err := tpl.Parse("$h $w!\n")
        if err != nil {
            fmt.Println(err)
            return
        }

        err = tpl.Run(os.Stdout, ctx)
        if err != nil {
            fmt.Println(err)
        }
    }

More examples are in *examples* directory

There are next few methods/functions for Go's template compatibility:

    func (*Template) Execute(data interface{}, wr io.Writer) os.Error
    func (*Template) ParseFile(filename string) os.Error
    func Parse(str string) (*Template, os.Error)
    func ParseFile(filename string) (*Template, os.Error)

and one method for mustache.go compatibility:

    func (*Template) Render(ctx ...interface{}) string

Example:

    tpl_filename := "template.kt"

    // Template loading
    tpl, err := ParseFile(tpl_filename)
    if err != nil {
        fmt.Println(tpl_filename, "-", err)
        return
    }

    // Strict mode rendering
    tpl.Strict = true

    // Render to stdout - template.go way
    err = tpl.Execute(data, os.Stdout)

    // One more time to stdout - mustache.go way
    err = fmt.Println(tpl.Render(context_data))

# Kasia's syntax in examples

## Getting data from a array/slice

If the context is a slice or an array you can get variable
from it like this:

    $[0] $[2]

You can use negative indexes in Python way: `a[-n] = a[len(a)-n]`.

If `$[0]` is int variable, you can use it as an index:

    $[[0]](1.1)

In this example `$[[0]]` is a function, so we calling it with float argument.
Note the lack of '$' sign for index variable.


## Geting data from a struct and function calling

If the context is a struct like this:

    type ContextData struct {
        A   int
        B   string
        C   bool
        D   []interface{}
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

you can get values from it using field's name or field's index:

    $A    ==  $[0]
    $B    ==  $[1]
    $C    ==  $[2]
    $D[A] ==  $[3][A]   // $A is integer used as index
    $E.a  ==  $[4].a
    $G.A  ==  $[6][0]
    $G.A  ==  $[-1][0]  // Index can be negative.

You can use variable or string to specify struct field. If $A == 2 and $B == "A"
then:

    $[A]  ==  $[2]  ==  $C
    $[B]  ==  $["$B"]  ==  $["A"]  ==  $A

Strings must be enclosed in quotes or apostrophes. If you want to use quote,
apostrophe or dolar sign in the string, use $", $' or $$ insted. In the fact,
strings are subtemplates with the same context as main template:

    $F(0)("$$$A $"and$" $$$["A"]", 1) == $['F'](0)('$$$A "and" $$$['A']', 1)

In previous example you can also see how pass arguments to a function (also
where the function returning other function). In the same way you can call
methods:

    $:M1       // M1 has no arguments
    $:M1()     // M1 has no arguments, force function/method
    $M2("A")   // Doesn't work (potiner method - see note)
    $G.M2("A") // Works, becouse $$G is pointer (see note)

[Note about pointer methods](http://groups.google.com/group/golang-nuts/browse_thread/thread/ec6b27e332ed7f77).

Types of arguments passed to a function/method must match the function
definition. There is one exception: interface{} type in function definition
match any type. This will be improved in the furure if the reflect package will
provide Type.Implement(TypeInterface) method.

If F value is a function, `$F` or `$F()` calls this function without parameters
and returns its first return value.

If F value is a reference (interface/pointer) to the function, `$F` returns this
reference, `$F()` calls the function without parameters and returns its first
return value.

If F value is a function or a reference to the function, `$F(a, b)` calls the
function with a, b parameters and returns its first return value.

If data context is a function, you can call it like this:

    $(8)

You can explicitly specify variable or function boundaries using braces:

    eee${A}eee uuu$:{A}uuu

By default, if variable doesn't exist you get empty string. You can change this
behavior by setting `Template.Strict` to true and now get error code if variable
doesn't exist. If variable is used as function's argument or index it is always
executed in strict mode.

`fmt.Fprint()` and `fmt.Sprint()` functions are used to render variables values.

## Getting data from a map

If the data context is a map you can use it in way smilar to struct context.
Context map can be string or int indexed. If your map is string indexed like
this:

    map[string]interface{} {
        "a":    1,
        "b":    "'&text in map&'\"",
        "c":    false,
        "d":    data_slice[1:],
        "e":    data_map,
        "f":    &data_func,
        "g":    &data_struct,
    }

you can get values from it like this:

    $a          ==  $["a"]
    $:b         ==  $:["b"]
    $c          ==  $['c']
    $:d[0](2.1) ==  $:['d'][0](2.1)
    $f(1)       ==  $["f"](1)

If map is int indexed

    map[int]string {
        -1:     "minus jeden",
        101:    "sto jeden",
    }

you can use only index notation:

    $$[-1]
    $$[101]

## Escaping

If you create template using `New()` function, all values are rendered HTML
escaped by default. If you want unescaped text use ':' just behind '$' sign:

    $:B

If you want other escape function, define it your own as:

    func(io.Writer, string) os.Error

and assign it to Template.EscapeFunc field before rendering. Variables used as
parameters or indexes are always unescaped.

If you don't use `New()` to create template, EscapeFunction is nil and there is
no escaping.

## Control statements

### 'If' statement

    $if 0.0:
        Text 1
    $elif A:
        Text 2
        $if A:TTTT${end}
    $elif "$if B: Txt$B $end":
        Text 3
    $else:
        Else text
    $end

Value is false when:
1. Is bool and false.
2. Is int and eqal 0.
3. Is float and eqal 0.0.
4. Is complex and eqal cplx(0, 0)
5. Is zero length string, slice, array or map.
6. Is nil interface or pointer.

This form of *if* doesn't dereference interfaces nor pointers before evaluate.

You can use simple comparasion in if/elif statement:

    $if "sss" == 1:
        It's false
    $elif 2 == 2.0:
        Note! It's false!
    $elif A > 2:
        Text $A
    $elif B != C:
        B=$B, C=$C
    $end

If compared values have diferent types you should use only '!=' and '=='
operators. If the types match, Go's comparasion rules are applied. This form of
*if* dereference interfaces and pointers before comparasion.

You can use braces for all statements (useful for `$end` statement):

    Start$if 1 < 0:AAAA$else:BBBB${end}stop.

### 'For' statement

    $for i, v in D:
        $i: $if i == 1:
            $v(1.1)
        $else:
            D[$i] = $v
        $end
    $else:
        D == nil or len(D) == 0
    $end

Value of D can be any type. If it's an array or a slice, the first block is
executed for each of its element, v = element value, i = element index. If value
is a scalar, v = value, i = 0. If you add '+' to index name:

    $for i+, v in D:
        $i...
    $end

index will be increased by one.

If there is newline just behind *if*, *for*, *else*, *end* statement, this new
line isn't printed. If you want print this new line, insert this statement into
braces:

    ${for i, v in D:}
    $i,
    $end

## Coments

All the text between $# and #$ are ignored: not interpreted and not printed to
the output.

    $# This is my comment #$

    This is interpreted code $a, $b.

    $#
    This code are ignored

    $a $b
    #$

## Subtemplates (nested/embded templates)

If you have modularize your template you can do this like int this example:

    type Ctx struct {
        a int
        b string
        c *SubCtx
        tpl1 *Template
        tpl2 *Template
    }

    type SubCtx struct {
        s int
        t float
    } 

Main template:

    Variables from main context: $a, $b

    First subtemplate: $tpl1

    Second subtemplate: $tpl2.Nested(c)

Subtemplate tpl1:

    This template uses main context: $a, $b

    You can lead to a loop if uncoment this: $# $tpl1 #$

Subtemplate tpl2:

    This template operates on context passed to it via Nested method: $s, $t

How you can see in the example above, you can use subtemplates in two ways:
1. Render subtemplate on main context by typing `$tpl1`.
2. Render subtemplate with custom context: `$tpl2.Nested(custom_context)`.
