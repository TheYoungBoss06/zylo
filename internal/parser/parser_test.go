package parser

import (
	"testing"
	"github.com/zylo-lang/zylo/internal/ast"
	"github.com/zylo-lang/zylo/internal/lexer"
)

func TestVarStatements(t *testing.T) {
	input := `
var x = 5;
var y = 10;
var foobar = 838383;
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testVarStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch expected.(type) {
	case int64:
		intVal, ok := exp.(*ast.NumberLiteral)
		if !ok {
			t.Errorf("exp not *ast.NumberLiteral. got=%T", exp)
			return false
		}
		if intVal.Value != expected {
			t.Errorf("intVal.Value not %d. got=%d", expected, intVal.Value)
			return false
		}
	case float64:
		// Nota: El AST actual solo soporta int64 para NumberLiteral.
		// Esto requerirá una refactorización futura para manejar floats correctamente.
		// Por ahora, solo verificamos que el tipo sea NumberLiteral.
		_, ok := exp.(*ast.NumberLiteral)
		if !ok {
			t.Errorf("exp not *ast.NumberLiteral. got=%T", exp)
			return false
		}
	case string:
		strVal, ok := exp.(*ast.StringLiteral)
		if !ok {
			t.Errorf("exp not *ast.StringLiteral. got=%T", exp)
			return false
		}
		if strVal.Value != expected {
			t.Errorf("strVal.Value not %q. got=%q", expected, strVal.Value)
			return false
		}
	case bool:
		boolVal, ok := exp.(*ast.BooleanLiteral)
		if !ok {
			t.Errorf("exp not *ast.BooleanLiteral. got=%T", exp)
			return false
		}
		if boolVal.Value != expected {
			t.Errorf("boolVal.Value not %t. got=%t", expected, boolVal.Value)
			return false
		}
	default:
		t.Errorf("unhandled type for literal test: %T", expected)
		return false
	}
	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func testVarStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "var" {
		t.Errorf("s.TokenLiteral not 'var'. got=%q", s.TokenLiteral())
		return false
	}

	varStmt, ok := s.(*ast.VarStatement)
	if !ok {
		t.Errorf("s not *ast.VarStatement. got=%T", s)
		return false
	}

	if varStmt.Name.Value != name {
		t.Errorf("varStmt.Name.Value not '%s'. got=%s", name, varStmt.Name.Value)
		return false
	}

	if varStmt.Name.TokenLiteral() != name {
		t.Errorf("varStmt.Name.TokenLiteral() not '%s'. got=%s", name, varStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func TestIfStatement(t *testing.T) {
	input := `if (x < y) { x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Consequence.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.Identifier).Value != "x" {
		t.Errorf("consequence is not %s. got=%s", "x", stmt.Consequence.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.Identifier).Value)
	}

	if stmt.Alternative != nil {
		t.Errorf("stmt.Alternative.Statements was not nil. got=%+v", stmt.Alternative)
	}
}

func TestIfElseStatement(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfStatement. got=%T",
			program.Statements[0])
	}

	if stmt.Consequence.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.Identifier).Value != "x" {
		t.Errorf("consequence is not %s. got=%s", "x", stmt.Consequence.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.Identifier).Value)
	}

	if stmt.Alternative == nil {
		t.Errorf("stmt.Alternative.Statements was nil")
	}
}

func TestExpressionStatement(t *testing.T) {
	input := `show("Hola");`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	callExp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
			stmt.Expression)
	}

	ident, ok := callExp.Function.(*ast.Identifier)
	if !ok {
		t.Fatalf("callExp.Function is not ast.Identifier. got=%T",
			callExp.Function)
	}

	if ident.Value != "show" {
		t.Errorf("function name is not 'show'. got=%s", ident.Value)
	}

	if len(callExp.Arguments) != 1 {
		t.Fatalf("wrong number of arguments. got=%d", len(callExp.Arguments))
	}

	testLiteralExpression(t, callExp.Arguments[0], "Hola")
}
