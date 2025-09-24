package main

import (
	"os"
	"testing"

	"github.com/zylo-lang/zylo/internal/evaluator"
	"github.com/zylo-lang/zylo/internal/lexer"
	"github.com/zylo-lang/zylo/internal/parser"
)

func TestCalculator(t *testing.T) {
	// Leer archivo calculator.zylo
	content, err := os.ReadFile("examples/calculator.zylo")
	if err != nil {
		t.Fatalf("Error leyendo archivo: %v", err)
	}

	t.Log("🚀 Ejecutando calculadora Zylo...")
	t.Log("=================================")

	// Parsear
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Log("Errores de parsing:")
		for _, err := range p.Errors() {
			t.Errorf("  %s", err)
		}
		t.Fatal("Errores de parsing")
	}

	// Ejecutar con el evaluador
	eval := evaluator.NewEvaluator()
	err = eval.EvaluateProgram(program)
	if err != nil {
		t.Fatalf("Error de ejecución: %v", err)
	}

	t.Log("\n✅ ¡Calculadora ejecutada exitosamente!")
}
