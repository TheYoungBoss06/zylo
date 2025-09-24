package ast

import (
	"fmt"

	"github.com/zylo-lang/zylo/internal/lexer"
)

// Variable representa una variable (usada para parámetros de función).
type Variable struct {
	Token lexer.Token // El token IDENTIFIER.
	Name  string
	Type  string // Anotación de tipo opcional.
}

// expressionNode implementa la interfaz Expression.
func (v *Variable) expressionNode() {}

// TokenLiteral devuelve el literal del token de la variable.
func (v *Variable) TokenLiteral() string { return v.Token.Lexeme }

// String devuelve una representación en string de la variable.
func (v *Variable) String() string {
	if v.Type != "" {
		return fmt.Sprintf("%s: %s", v.Name, v.Type)
	}
	return v.Name
}
