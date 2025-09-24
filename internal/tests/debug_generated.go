package main

import (
    "fmt"
)

type Persona struct {
    nombre interface{}
    edad interface{}
}

func NewPersona(n interface{}, e interface{}) *Persona {
    obj := &Persona{}
    obj.init(    n    ,     e    )
    return obj
}

func (obj *Persona) init(n interface{}, e interface{}) {
obj.nombre     =     n
    n
obj.edad     =     e
    e
}

func (obj *Persona) saludar() {
fmt.Println((((("Hola, soy " + obj.nombre) + " y tengo ") + obj.edad) + " años."))
}

func factorial(n interface{}) {
    if     (    n     <=     1    )     {
        return                 1
        1
    }     else {
        if         (        n         ==         2        )         {
            return                         2
            2
        }
    }
}
func main() {
    HASH_LITERAL
    return         (    n     * factorial((n - 1))    )
p.saludar()
fmt.Println(("Factorial de 5: " + factorial(5)))
    var i     =     0
    0
    for     (    i     <     5    )     {
        if         (        i         ==         2        )         {
fmt.Println("Encontré el número 2, saltando…")
        }
    }
    (    i     =     i     +     1    )
    1
    continue
fmt.Println(("i = " + i))
    (    i     =     i     +     1    )
    1
    for _, x := range     zyloruntime.NewList()    .Append(    1    )    .Append(    2    )    .Append(    3    )    .Append(    4    )    .Append(    5    )     {
fmt.Println(("Iterando en lista: " + x))
    }
    zyloruntime.Try(func() {
        var z         =         (        10         /         0        )
        0
    }, func(err error) {
        var err = err
fmt.Println(("Error atrapado: " + err))
    })
}
