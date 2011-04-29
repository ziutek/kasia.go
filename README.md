# FAQ

## What is Kasia.go?

Kasia.go is a Go implementation of the Kasia templating system.

Kasia is primarily designed for HTML, but you can use it for anything you want.
Its syntax is somewhat similar to the web.py templating system - Templator, but much
much simpler.

## What does "Kasia" mean?

Kasia is my daughter's name (Polish equivalent of Katie or Kathy).

## Installing Kasia.go

### Using *goinstall* - prefered way:

    $ goinstall github.com/ziutek/kasia.go

After this command *kasia.go* package is ready to use. You may find source in

    $GOROOT/src/pkg/github.com/ziutek/kasia.go

directory. You can use `goinstall -u -a` for update all installed packages.

### Using *git clone* command:

    $ git clone git://github.com/ziutek/kasia.go
    $ cd kasia.go && make install

### Version for Go weekly releases

If master branch can't be compiled with Go weekly release, try clone
Kasia.go weekly branch:

    $ git clone -b weekly git://github.com/ziutek/kasia.go
    $ cd kasia.go && make install

## Using "Kasia.go"

Kasia's native interface contains one function and three methods:

    func New() *Template
    func (*Template) Parse(str string) os.Error
    func (*Template) Run(wr io.Writer, ctx ...interface{}) os.Error
    func (*Template) Nested(ctx ...interface{}) *NestedTemplate 

The simplest example is:

    package main

    import (
        "os"
        "fmt"
        "github.com/ziutek/kasia.go"
    )

    type Ctx struct {
        H, W string
    }

    func main() {
        ctx := &Ctx{"Hello", "world"}

        tpl, err := kasia.Parse("$H $W!\n")
        if err != nil {
            fmt.Println(err)
            return
        }

        err = tpl.Run(os.Stdout, ctx)
        if err != nil {
            fmt.Println(err)
        }
    }

More examples are in *examples* directory. You need to install *web.go* for
some of them:

    $ goinstall github.com/hoisie/web.go

See [simple_go_wiki tutorial](https://github.com/ziutek/simple_go_wiki) for
example usage of Kasia.go in simple Wiki.


There are a few methods/functions for Go template compatibility:

    func (*Template) ParseFile(filename string) os.Error
    func Parse(txt string) (*Template, os.Error)
    func MustParse(txt string) *Template
    func ParseFile(filename string) (*Template, os.Error)
    func MustParseFile(filename string) (*Template)
    func (*Template) Execute(wr io.Writer, data interface{}) os.Error

one method and one function for mustache.go compatibility (they panics when
error occurs):

    func (*Template) Render(ctx ...interface{}) string
    func Render(txt string, ctx ...interface{}) string

and their counterparts with error reporting (without panic):

    func (*Template) RenderString(ctx ...interface{}) (string, os.Error)
    func RenderString(txt string, strict bool, ctx ...interface{}) (string, os.Error)

    func (*Template) RenderBytes(ctx ...interface{}) ([]byte, os.Error)
    func RenderBytes(txt string, strict bool, ctx ...interface{}) ([]byte, os.Error)

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

    // Render to stdout, template.go way
    err = tpl.Execute(os.Stdout, data)

    // One more time, mustache.go way
    err = fmt.Println(tpl.Render(data))

# Introduction to Kasia template syntax

## Getting data from an array/slice

If the context is a slice or an array you can get an element
from it like this:

    $[0] $[2]

If `$[0]` is an int variable, you can use it as an index:

    $[[0]](1.1)

In this example `$[[0]]` is a function, so we are calling it with a float argument (1.1).
Note the lack of a '$' sign for variable indexes.


## Geting data from a struct and calling functions

If the context is a struct, for example:

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

you can get values from it using the name or index of a field:

    $A    ==  $[0]
    $B    ==  $[1]
    $C    ==  $[2]
    $D[A] ==  $[3][A]   // $A is an integer used as index
    $E.a  ==  $[4].a
    $G.A  ==  $[6][0]

You can get values only from exported fields. 

You can use variables or strings to specify struct fields. If $A == 2 and $B == "A"
then:

    $[A]  ==  $[2]  ==  $C
    $[B]  ==  $["$B"]  ==  $["A"]  ==  $A

Strings must be enclosed in single or double quotes. If you want to use quotes
or dollar signs in the string, use $", $' or $$ instead. In fact, strings are
subtemplates with the same context as the main template:

    $F(0)("$$$A $"and$" $$$["A"]", 1) == $['F'](0)('$$$A "and" $$$['A']', 1)

In the previous example you can also see how to pass arguments to a function (even
when the function returns another function). In the same way you can call
methods:

    $:M1       // M1 has no arguments
    $:M1()     // M1 has no arguments, force function/method
    $M2("A")   // Doesn't work (potiner method - see note)
    $G.M2("A") // Works, becouse $G is a pointer (see note)

[Note about pointer methods](http://groups.google.com/group/golang-nuts/browse_thread/thread/ec6b27e332ed7f77).

Types of arguments passed to a function/method must be assignable to the
function arguments.

If variable F is a function, `$F` or `$F()` calls this function without parameters
and returns its first return value.

If variable F is a reference (interface/pointer) to a function, `$F` returns the
reference, `$F()` calls the function without parameters and returns its first
return value.

If variable F is a function or a reference to a function, `$F(a, b)` calls the
function with arguments a, b and returns its first return value.

If the context is a function, for example:

    func ctx(i int) int {
        i++
        return i
    }

you can call it like this:

    $(8) == 9
    $((((8)))) == 12

You can explicitly specify variables or function boundaries using braces:

    eee${A}eee uuu$:{A}uuu

By default, if a variable doesn't exist you get an empty string. You can change
this behavior by setting `Template.Strict` to true. Undefined variables now
return an error code.  If a variable is used as argument to a function or an
index it is always executed in strict mode.

If variable is []byte slice its content is treated as text. It is better use
[]byte than string if you don't use escaping, because it doesn't need conversion
before write to io.Writer.

`fmt.Fprint()` and `fmt.Sprint()` are used to render a different type
of variables.

## Getting data from a map

If the context is a map you can use it in a way similar to a struct context.
If the map is has string key like this:

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

If the map hes int key:

    map[int]string {
        -1:     "minus jeden",
        101:    "sto jeden",
    }

you can only use index notation:

    $[-1]
    $[101]

Map may have a key of any type but, you can use directly (without using a variable) only string, int or float key.

## Escaping

If you create template using the `New()` function, all values are rendered HTML
escaped by default. If you want unescaped text use ':' after the '$' sign:

    $:B

If you want to use a different escape function, define your own as:

    func(io.Writer, []byte) os.Error

and assign it to Template.EscapeFunc before rendering. Variables used as
parameters or indexes are always unescaped.

If you don't use `New()` to create templates, EscapeFunction is nil and there is
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

Value is false when it:

1. Is bool and false.
2. Is int and equal to 0.
3. Is float and equal to 0.0.
4. Is complex and equal to cplx(0, 0).
5. Is a zero length string, slice, array or map.
6. Is nil.
7. Doesn't exist (even in strict mode)

This form of *if* doesn't dereference interfaces or pointers before evaluation.

You can use simple comparison in if/elif statement:

    $if "sss" == 1:
        It's false
    $elif 2 == 2.0:
        Note! It's false!
    $elif A > 2:
        Text $A
    $elif B != C:
        B=$B, C=$C
    $end

If the compared values have different types you should only use the '!=' and '=='
operators. If the types match, Go comparison rules are applied. This form of
*if* dereference interfaces and pointers before comparison.

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

D can be any type. If it's an array, a slice, a map or a channel, the first
block is executed for each element of it and v = element value, i = element
number or key. If value is a scalar, v = value, i = nil. If you add '+' to the
index name:

    $for i+, v in D:
        $i...
    $end

index will be increased by one. You can't increment the map key if you iterate
over map even if map key is integer.

If there is newline after *if*, *for*, *else*, *end* statement, this new
line isn't printed. If you want print this new line, insert this statement in
braces:

    ${for i, v in D:}
    $i,
    $end

## Comments

All the text between $# and #$ is ignored: not interpreted and not printed.

    $# This is my comment #$

    This is interpreted text/code: $a, $b.

    $#
    This is ignored  text/code

    $a $b
    #$

## Subtemplates (nested/embedded templates)

If you want modularize your templates, you can do like in folowing example:

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

    You can lead to a loop if you uncoment this: $# $tpl1 #$

Subtemplate tpl2:

    This template operates on context passed to it via Nested method: $s, $t

As you can see, you can use subtemplates in two ways:

1. Render subtemplate with main context by typing `$tpl1`.
2. Render subtemplate with custom context: `$tpl2.Nested(custom_context)`.

## Context stack

You can divide the data context into two (or more) parts. For example, the first
part of the context may be global data:

    type Ctx struct {
        a, b  string
    }

the second part may be local data:

    type LocalCtx struct {
        b string
        c int
    }

You can pass them together to the *Run* method:

    // Global data
    var data = Ctx{//...} 

    func hello(web_ctx *web.Context, val string) {
        // Local data
        var ld *LocalCtx

        if len(val) > 0 {
            ld = &LocalCtx{val, len(val)}
        }

        // Rendering the data
        err := tpl.Run(web_ctx, data, ld)
        if err != nil {
            fmt.Fprintln(web_ctx, "Error:", err)
        }
    }

When `$b` occurs in a template, *Run* first looks for it in *ld*, and only if it
doesn't find it, looks for it in *data*.

You may set the global data once and use they together wit local data created
each *hello()* call. Additionally, if *ld* isn't nil, the *b* field in *ld*
will be rendered, not the *b* field in *data*.

You could create a context stack using the values without internal structure:

    tpl.Run(os.Stdout, 2, "Ala", 3.14159)

To render such values you should use *@* symbol:

    $@[0]  $@[1]  $@[2]

Output:

    2  Ala  3.14159

*@* means context stack itself and behaves like ordinary slice.
