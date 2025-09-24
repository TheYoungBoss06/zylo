package sema

import (
	"fmt"

	"github.com/zylo-lang/zylo/internal/ast"
)

// SymbolTable representa una tabla de símbolos para un ámbito específico.
type SymbolTable struct {
	parent     *SymbolTable
	symbols    map[string]*Symbol // Mapea nombres de identificadores a sus símbolos.
	scopeName  string
	scopeLevel int
}

// Symbol representa una entrada en la tabla de símbolos.
type Symbol struct {
	Name  string
	Type  string // Tipo del identificador (e.g., "int", "string", "any").
	Scope string // Ámbito en el que se definió.
}

// NewSymbolTable crea una nueva tabla de símbolos.
func NewSymbolTable(scopeName string, level int, parent *SymbolTable) *SymbolTable {
	return &SymbolTable{
		parent:     parent,
		symbols:    make(map[string]*Symbol),
		scopeName:  scopeName,
		scopeLevel: level,
	}
}

// Define añade un nuevo símbolo a la tabla de símbolos.
func (st *SymbolTable) Define(name string, symType string) *Symbol {
	symbol := &Symbol{
		Name:  name,
		Type:  symType,
		Scope: fmt.Sprintf("%s (Level %d)", st.scopeName, st.scopeLevel),
	}
	st.symbols[name] = symbol
	return symbol
}

// Resolve busca un símbolo en la tabla de símbolos actual y en sus padres.
func (st *SymbolTable) Resolve(name string) (*Symbol, bool) {
	if sym, ok := st.symbols[name]; ok {
		return sym, true
	}
	if st.parent != nil {
		return st.parent.Resolve(name)
	}
	return nil, false
}

// SemanticAnalyzer es el struct principal para el análisis semántico.
type SemanticAnalyzer struct {
	symbolTable *SymbolTable
	errors      []string
}

// NewSemanticAnalyzer crea un nuevo analizador semántico.
func NewSemanticAnalyzer() *SemanticAnalyzer {
	// Inicializar con una tabla de símbolos global.
	globalScope := NewSymbolTable("global", 0, nil)
	return &SemanticAnalyzer{
		symbolTable: globalScope,
		errors:      []string{},
	}
}

// Analyze ejecuta el análisis semántico sobre un AST.
func (sa *SemanticAnalyzer) Analyze(node ast.Node) {
	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			sa.Analyze(stmt) // Recursivamente analizar cada sentencia.
		}
	case *ast.VarStatement:
		// Definir la variable en la tabla de símbolos.
		// Por ahora, asumimos tipo "any" por defecto.
		sa.symbolTable.Define(n.Name.Value, "any")
		// Analizar el valor de la expresión y su tipo.
		if n.Value != nil {
			sa.Analyze(n.Value)
		}
	case *ast.ExpressionStatement:
		sa.Analyze(n.Expression)
	case *ast.Identifier:
		// Al encontrar un identificador, verificar si está definido.
		if _, ok := sa.symbolTable.Resolve(n.Value); !ok {
			sa.addError(fmt.Sprintf("identifier not found: %s", n.Value))
		}
	case *ast.FuncStatement:
		// Registrar la función en la tabla de símbolos.
		// Por ahora, el tipo de la función es genérico "func".
		sa.symbolTable.Define(n.Name.Value, "func")
		// Analizar el cuerpo de la función en un nuevo scope.
		sa.enterScope(n.Name.Value)
		// Registrar los parámetros de la función en el nuevo scope
		for _, param := range n.Parameters {
			sa.symbolTable.Define(param.Value, "any") // Usar param.Value ya que es *ast.Identifier
		}
		sa.Analyze(n.Body)
		sa.exitScope()
	case *ast.BlockStatement:
		// Analizar cada sentencia dentro del bloque.
		for _, stmt := range n.Statements {
			sa.Analyze(stmt)
		}
	case *ast.CallExpression:
		// Analizar la función y los argumentos
		sa.Analyze(n.Function)
		for _, arg := range n.Arguments {
			sa.Analyze(arg)
		}
	case *ast.InfixExpression:
		// Analizar las expresiones izquierda y derecha
		sa.Analyze(n.Left)
		sa.Analyze(n.Right)
	case *ast.PrefixExpression:
		// Analizar la expresión derecha
		sa.Analyze(n.Right)
	case *ast.NumberLiteral, *ast.StringLiteral, *ast.BooleanLiteral:
		// Los literales no necesitan análisis semántico adicional
		// TODO: Añadir manejo para otros tipos de nodos (ClassDecl, IfStmt, WhileStmt, etc.)
	}
}

// enterScope crea un nuevo ámbito y lo establece como el ámbito actual.
func (sa *SemanticAnalyzer) enterScope(name string) {
	newScope := NewSymbolTable(name, sa.symbolTable.scopeLevel+1, sa.symbolTable)
	sa.symbolTable = newScope
}

// exitScope sale del ámbito actual y vuelve al ámbito padre.
func (sa *SemanticAnalyzer) exitScope() {
	if sa.symbolTable.parent != nil {
		sa.symbolTable = sa.symbolTable.parent
	}
}

// addError añade un error de análisis semántico.
func (sa *SemanticAnalyzer) addError(msg string) {
	sa.errors = append(sa.errors, msg)
}

// Errors devuelve los errores encontrados.
func (sa *SemanticAnalyzer) Errors() []string {
	return sa.errors
}
