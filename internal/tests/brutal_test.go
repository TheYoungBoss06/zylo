package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zylo-lang/zylo/internal/ast"
	"github.com/zylo-lang/zylo/internal/codegen"
	"github.com/zylo-lang/zylo/internal/lexer"
	"github.com/zylo-lang/zylo/internal/parser"
)

func TestBrutalComprehensive(t *testing.T) {
	// Path to the Zylo source file
	sourcePath := "../../tests/brutal_test.zylo"

	// Check if the source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Skip("Test file brutal_test.zylo not found")
	}

	// Compile the Zylo code to Go
	generatedCode, err := compileZyloToGo(sourcePath)
	if err != nil {
		t.Fatal("Failed to compile Zylo code:", err)
	}

	// Execute the generated Go code
	output, err := executeGeneratedCode(generatedCode)
	if err != nil {
		t.Fatal("Failed to execute generated code:", err)
	}

	// Verify the output contains the expected fragments
	expectedFragments := []string{
		"Hola, soy Wilson",        // From object instantiation
		"Factorial de 5: 120",     // From function call
		"Error atrapado",          // From exception handling
	}

	for _, fragment := range expectedFragments {
		if !strings.Contains(output, fragment) {
			t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", fragment, output)
		}
	}
}

func compileZyloToGo(sourcePath string) (string, error) {
	// Read the Zylo source file
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", err
	}

	// Parse the Zylo code
	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	if errors := p.Errors(); len(errors) > 0 {
		return "", fmt.Errorf("parsing errors: %v", errors)
	}

	// Generate Go code
	cg := codegen.NewCodeGenerator()
	generatedCode, err := cg.Generate(program)
	if err != nil {
		return "", err
	}

	// DEBUGGING: Imprimir el código generado
	fmt.Println("=== CÓDIGO GO GENERADO ===")
	fmt.Println(generatedCode)
	fmt.Println("=== FIN CÓDIGO GENERADO ===")

	// También guardar en archivo para inspección
	os.WriteFile("debug_generated.go", []byte(generatedCode), 0644)

	return generatedCode, nil
}

func executeGeneratedCode(code string) (string, error) {
	// Create a temporary Go file
	tempDir, err := os.MkdirTemp("", "zylo_test")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(tempFile, []byte(code), 0644); err != nil {
		return "", err
	}

	// Run the Go code using 'go run'
	cmd := exec.Command("go", "run", tempFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("execution failed: %v, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

func TestDebugCodeGeneration(t *testing.T) {
	simpleCode := `
func test() {
	 show.log("hello");
}
`

	l := lexer.New(simpleCode)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	cg := codegen.NewCodeGenerator()
	generated, err := cg.Generate(program)
	if err != nil {
		t.Fatalf("Code generation error: %v", err)
	}

	fmt.Println("Input Zylo:")
	fmt.Println(simpleCode)
	fmt.Println("\nGenerated Go:")
	fmt.Println(generated)
}

func TestSpecificCallExpression(t *testing.T) {
	code := `show.log("test");`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Printf("Parser error: %s\n", err)
		}
	}

	fmt.Printf("AST Statements: %d\n", len(program.Statements))
	for i, stmt := range program.Statements {
		fmt.Printf("Statement %d: %T\n", i, stmt)
		if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok {
			fmt.Printf("  Expression: %T\n", exprStmt.Expression)
			if callExpr, ok := exprStmt.Expression.(*ast.CallExpression); ok {
				fmt.Printf("  Function: %T\n", callExpr.Function)
				fmt.Printf("  Arguments: %d\n", len(callExpr.Arguments))
			}
		}
	}

	cg := codegen.NewCodeGenerator()
	generated, err := cg.Generate(program)
	if err != nil {
		t.Fatalf("Code generation error: %v", err)
	}
	fmt.Println("Generated:")
	fmt.Println(generated)
}
