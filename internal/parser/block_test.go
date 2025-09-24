package parser

import (
	"testing"
	"github.com/zylo-lang/zylo/internal/lexer"
	"github.com/zylo-lang/zylo/internal/ast"
)

func TestBlockParsing(t *testing.T) {
	input := `func main() {
    var message = "Hola desde Zylo!";
    print(message);
}`
	
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}
	
	// Debug: print what we got
	t.Logf("Program statements: %d", len(program.Statements))
	for i, stmt := range program.Statements {
		t.Logf("Statement %d: %T", i, stmt)
	}
	
	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}
	
	funcStmt, ok := program.Statements[0].(*ast.FuncStatement)
	if !ok {
		t.Fatalf("Expected FuncStatement, got %T", program.Statements[0])
	}
	
	if funcStmt.Name.Value != "main" {
		t.Fatalf("Expected function name 'main', got '%s'", funcStmt.Name.Value)
	}
	
	if len(funcStmt.Body.Statements) != 2 {
		t.Fatalf("Expected 2 statements in function body, got %d", len(funcStmt.Body.Statements))
	}
	
	// Debug: print what we got in the function body
	for i, stmt := range funcStmt.Body.Statements {
		t.Logf("Function body statement %d: %T", i, stmt)
	}
	
	// Check first statement (var declaration)
	varStmt, ok := funcStmt.Body.Statements[0].(*ast.VarStatement)
	if !ok {
		t.Fatalf("Expected VarStatement, got %T", funcStmt.Body.Statements[0])
	}
	
	if varStmt.Name.Value != "message" {
		t.Fatalf("Expected variable name 'message', got '%s'", varStmt.Name.Value)
	}
	
	// Check second statement (print call)
	exprStmt, ok := funcStmt.Body.Statements[1].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", funcStmt.Body.Statements[1])
	}
	
	callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %T", exprStmt.Expression)
	}
	
	ident, ok := callExpr.Function.(*ast.Identifier)
	if !ok {
		t.Fatalf("Expected Identifier, got %T", callExpr.Function)
	}
	
	if ident.Value != "print" {
		t.Fatalf("Expected function name 'print', got '%s'", ident.Value)
	}
}
