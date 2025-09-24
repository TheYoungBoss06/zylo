package sema

import (
	"testing"
	"github.com/zylo-lang/zylo/internal/ast"
	"github.com/zylo-lang/zylo/internal/lexer"
	"github.com/zylo-lang/zylo/internal/parser"
)

func TestSemanticAnalysis(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedErrors int
		expectedSymbols map[string]string // Nombre -> Tipo (simplificado a "any" por ahora)
	}{
		{
			name: "Simple variable declaration",
			input: `
var x = 5;
`,
			expectedErrors: 0,
			expectedSymbols: map[string]string{
				"x": "any",
			},
		},
		{
			name: "Multiple variable declarations",
			input: `
var a = 1;
var b = "hello";
var c = true;
`,
			expectedErrors: 0,
			expectedSymbols: map[string]string{
				"a": "any",
				"b": "any",
				"c": "any",
			},
		},
		{
			name: "Undeclared identifier usage",
			input: `
var x = y; // y is not declared
`,
			expectedErrors: 1, // Esperamos un error de "identifier not found"
			expectedSymbols: map[string]string{
				"x": "any", // 'x' debería ser definido, 'y' debería causar un error.
			},
		},
		{
			name: "Shadowing (optional, current implementation allows it)",
			input: `
var z = 10;
var z = 20; // This might be allowed or an error depending on language rules.
`,
			expectedErrors: 0, // Asumiendo que el shadowing simple es permitido por ahora.
			expectedSymbols: map[string]string{
				"z": "any", // El último 'z' define el símbolo.
			},
		},
		{
			name: "Function declaration and usage",
			input: `
func greet(name) {
  print("Hello, " + name);
}
greet("World");
`,
			expectedErrors: 0,
			expectedSymbols: map[string]string{
				"greet": "func", // La función 'greet' debe ser registrada.
			},
		},
		{
			name: "Undeclared function call",
			input: `
unknownFunc("test");
`,
			expectedErrors: 1, // Esperamos un error de "identifier not found" para unknownFunc.
			expectedSymbols: map[string]string{}, // No se esperan símbolos definidos aquí.
		},
		{
			name: "Function with parameters and body",
			input: `
func add(a, b) {
	return a + b;
}
var result = add(5, 10);
`,
			expectedErrors: 0,
			expectedSymbols: map[string]string{
				"add": "func",
				"result": "any",
			},
		},
		{
			name: "Nested scopes",
			input: `
var globalVar = 10;
func outer() {
	var outerVar = 20;
	func inner() {
		var innerVar = 30;
		print(globalVar);
		print(outerVar);
		print(innerVar);
	}
	inner();
	print(outerVar);
}
outer();
`,
			expectedErrors: 0,
			expectedSymbols: map[string]string{
				"globalVar": "any",
				"outer": "func",
				"outerVar": "any",
				"inner": "func",
				"innerVar": "any",
			},
		},
		{
			name: "Undeclared variable in nested scope",
			input: `
func outer() {
	print(undeclaredVar);
}
outer();
`,
			expectedErrors: 1, // Esperamos un error de "identifier not found" para undeclaredVar.
			expectedSymbols: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			sa := NewSemanticAnalyzer()
			sa.Analyze(program)

			if len(sa.Errors()) != tt.expectedErrors {
				t.Fatalf("Semantic analysis failed: expected %d errors, got %d. Errors: %v",
					tt.expectedErrors, len(sa.Errors()), sa.Errors())
			}

			// Verificar los símbolos definidos (simplificado)
			for name, symType := range tt.expectedSymbols {
				sym, ok := sa.symbolTable.Resolve(name)
				if !ok {
					t.Errorf("Symbol '%s' not found in symbol table", name)
				} else if sym.Type != symType {
					t.Errorf("Symbol '%s' has wrong type. Expected '%s', got '%s'", name, symType, sym.Type)
				}
			}
		})
	}
}

// Helper function to create a simple AST program for testing.
// This is a simplified approach; a real test would use the parser.
func createTestProgram(statements ...ast.Statement) *ast.Program {
	return &ast.Program{Statements: statements}
}

// Helper function to create a VarStatement for testing.
// This helper is not used by the current test suite but can be useful for future tests.
func createVarStatement(name string, value ast.Expression) *ast.VarStatement {
	return &ast.VarStatement{
		Token: lexer.Token{Type: lexer.VAR, Lexeme: "var"},
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Lexeme: name},
			Value: name,
		},
		Value: value,
	}
}

// Helper function to create an Identifier for testing.
// This helper is not used by the current test suite but can be useful for future tests.
func createIdentifier(name string) *ast.Identifier {
	return &ast.Identifier{
		Token: lexer.Token{Type: lexer.IDENTIFIER, Lexeme: name},
		Value: name,
	}
}
