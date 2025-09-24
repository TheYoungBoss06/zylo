package codegen

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/zylo-lang/zylo/internal/lexer"
	"github.com/zylo-lang/zylo/internal/parser"
)

func TestHelloWorld(t *testing.T) {
	input := `
var message = "Hola desde Zylo!";
print(message);
`
	expectedOutput := "Hola desde Zylo!\n"

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	cg := NewCodeGenerator()
	goCode, err := cg.Generate(program)
	if err != nil {
		t.Fatalf("Code generation failed: %v", err)
	}

	// Crear un directorio temporal para el código Go generado.
	tempDir, err := os.MkdirTemp("", "zylo_codegen_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Limpiar al final.

	goFilePath := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(goFilePath, []byte(goCode), 0644)
	if err != nil {
		t.Fatalf("Failed to write Go code to file: %v", err)
	}

	// Ejecutar gofmt para formatear el código generado (opcional, pero buena práctica).
	cmdFmt := exec.Command("go", "fmt", goFilePath)
	if err := cmdFmt.Run(); err != nil {
		t.Logf("gofmt failed (non-fatal): %v", err)
	}

	// Compilar el código Go generado.
	outputBinaryPath := filepath.Join(tempDir, "output")
	if runtime.GOOS == "windows" {
		outputBinaryPath += ".exe"
	}
	cmdBuild := exec.Command("go", "build", "-o", outputBinaryPath, goFilePath)
	var buildErr bytes.Buffer
	cmdBuild.Stderr = &buildErr
	if err := cmdBuild.Run(); err != nil {
		t.Fatalf("Go build failed: %v\nOutput:\n%s", err, buildErr.String())
	}

	// Ejecutar el binario generado.
	cmdRun := exec.Command(outputBinaryPath)
	var runOutput bytes.Buffer
	cmdRun.Stdout = &runOutput
	cmdRun.Stderr = &runOutput // Capturar stderr también
	if err := cmdRun.Run(); err != nil {
		t.Fatalf("Generated binary execution failed: %v\nOutput:\n%s", err, runOutput.String())
	}

	// Verificar la salida.
	if runOutput.String() != expectedOutput {
		t.Errorf("Unexpected output.\nExpected: %q\nGot: %q", expectedOutput, runOutput.String())
	}
}
