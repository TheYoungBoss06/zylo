package main

import (
	"fmt"
	"os"

	"github.com/zylo-lang/zylo/internal/codegen"
	"github.com/zylo-lang/zylo/internal/evaluator"
	"github.com/zylo-lang/zylo/internal/lexer"
	"github.com/zylo-lang/zylo/internal/parser"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "build":
		if len(os.Args) < 3 {
			fmt.Println("Error: Debes especificar un archivo .zylo")
			os.Exit(1)
		}
		buildFile(os.Args[2])
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Error: Debes especificar un archivo .zylo")
			os.Exit(1)
		}
		runFile(os.Args[2])
	case "help":
		printUsage()
	default:
		fmt.Printf("Comando desconocido: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Uso: zylo <comando> [archivo]")
	fmt.Println("Comandos:")
	fmt.Println("  build <archivo.zylo>  - Compila un archivo Zylo a Go")
	fmt.Println("  run <archivo.zylo>    - Ejecuta un archivo Zylo directamente")
	fmt.Println("  help                  - Muestra esta ayuda")
}

func buildFile(filename string) {
	// Verificar que el archivo existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("Error: El archivo '%s' no existe\n", filename)
		os.Exit(1)
	}

	// Verificar extensión
	if len(filename) < 5 || filename[len(filename)-5:] != ".zylo" {
		fmt.Printf("Error: El archivo debe tener extensión .zylo\n")
		os.Exit(1)
	}

	fmt.Printf("Compilando %s...\n", filename)

	// Leer archivo
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error leyendo archivo: %v\n", err)
		os.Exit(1)
	}

	// Parsear
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Printf("Errores de parsing:\n")
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		os.Exit(1)
	}

	// Generar código Go
	cg := codegen.NewCodeGenerator()
	goCode, err := cg.Generate(program)
	if err != nil {
		fmt.Printf("Error generando código: %v\n", err)
		os.Exit(1)
	}

	// Crear archivo de salida
	outputFile := filename[:len(filename)-5] + ".go"
	err = os.WriteFile(outputFile, []byte(goCode), 0644)
	if err != nil {
		fmt.Printf("Error escribiendo archivo: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Código generado en: %s\n", outputFile)
	fmt.Printf("Para ejecutar: go run %s\n", outputFile)
}

func runFile(filename string) {
	// Verificar que el archivo existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("Error: El archivo '%s' no existe\n", filename)
		os.Exit(1)
	}

	// Verificar extensión
	if len(filename) < 5 || filename[len(filename)-5:] != ".zylo" {
		fmt.Printf("Error: El archivo debe tener extensión .zylo\n")
		os.Exit(1)
	}

	fmt.Printf("Ejecutando %s...\n", filename)

	// Leer archivo
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error leyendo archivo: %v\n", err)
		os.Exit(1)
	}

	// Parsear
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Printf("Errores de parsing:\n")
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		os.Exit(1)
	}

	// Debug: Imprimir AST generado
	if len(os.Args) > 3 && os.Args[3] == "--debug" {
		fmt.Printf("AST generado: %+v\n", program)
		fmt.Printf("Número de statements: %d\n", len(program.Statements))
	}

	// Ejecutar directamente con el evaluador
	eval := evaluator.NewEvaluator()
	// InitBuiltins ya se llama en NewEvaluator()
	err = eval.EvaluateProgram(program)
	if err != nil {
		fmt.Printf("Error de ejecución: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Programa ejecutado exitosamente!")
}
