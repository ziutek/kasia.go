package main

import (
    "os"
    "kasia"
)

type Person struct {
    Name *string
    Age  *int
}

func main() {
    tpl, _ := kasia.Parse("$Name: $Age\n")

    name := "Oliver"
    age := 37
    p := &Person{&name, &age}
    tpl.Execute(p, os.Stdout)
}
