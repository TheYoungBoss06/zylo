package codegen

import (
	"fmt"
	"strings"

	"github.com/zylo-lang/zylo/internal/ast"
)

// CodeGenerator es el struct principal para la generación de código Go.
type CodeGenerator struct {
	output      strings.Builder
	indentation int
	classNames  []string
}

// NewCodeGenerator crea un nuevo CodeGenerator.
func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{classNames: make([]string, 0)}
}

// Generate genera código Go a partir de un AST.
func (cg *CodeGenerator) Generate(program *ast.Program) (string, error) {
	if program == nil {
		return "", fmt.Errorf("program is nil")
	}

	cg.writeString("package main\n\n")
	cg.writeString("import (\n")
	cg.writeString("    \"fmt\"\n")
	cg.writeString(")\n\n")

	// First pass: generate all function and class declarations
	for _, stmt := range program.Statements {
		if stmt != nil {
			if funcStmt, ok := stmt.(*ast.FuncStatement); ok {
				if funcStmt.Name.Value != "main" {
					cg.generateStatement(stmt)
				}
			} else if classStmt, ok := stmt.(*ast.ClassStatement); ok {
				cg.classNames = append(cg.classNames, classStmt.Name.Value)
				cg.generateStatement(stmt)
			}
		}
	}

	// Generate main function with executable statements
	cg.writeString("func main() {\n")
	cg.indent()

	for _, stmt := range program.Statements {
		// Skip function and class declarations in main
		if stmt != nil {
			if _, ok := stmt.(*ast.FuncStatement); ok {
				continue
			} else if _, ok := stmt.(*ast.ClassStatement); ok {
				continue
			}
			cg.generateStatement(stmt)
		}
	}

	cg.dedent()
	cg.writeString("}\n")

	return cg.output.String(), nil
}

// generateBreakStatement genera código Go para una sentencia 'break'.
func (cg *CodeGenerator) generateBreakStatement(stmt *ast.BreakStatement) {
	cg.writeString("break\n")
}

// generateContinueStatement genera código Go para una sentencia 'continue'.
func (cg *CodeGenerator) generateContinueStatement(stmt *ast.ContinueStatement) {
	cg.writeString("continue\n")
}

// generateForInStatement genera código Go para una sentencia 'for in'.
func (cg *CodeGenerator) generateForInStatement(stmt *ast.ForInStatement) {
	cg.writeString(fmt.Sprintf("for _, %s := range ", stmt.Identifier.Value))
	cg.generateExpression(stmt.Iterable)
	cg.writeString(" {\n")
	cg.indent()

	if stmt.Body != nil {
		for _, bodyStmt := range stmt.Body.Statements {
			cg.generateStatement(bodyStmt)
		}
	}

	cg.dedent()
	cg.writeString("}\n")
}

// generateStatement genera código Go para una sentencia del AST.
func (cg *CodeGenerator) generateStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.VarStatement:
		if s != nil {
			cg.generateVarStatement(s)
		}
	case *ast.ExpressionStatement:
		if s != nil {
			cg.generateExpressionStatement(s)
		}
	case *ast.FuncStatement:
		if s != nil {
			cg.generateFuncStatement(s)
		}
	case *ast.ReturnStatement:
		if s != nil {
			cg.generateReturnStatement(s)
		}
	case *ast.IfStatement:
		if s != nil {
			cg.generateIfStatement(s)
		}
	case *ast.WhileStatement:
		if s != nil {
			cg.generateWhileStatement(s)
		}
	case *ast.ForInStatement:
		if s != nil {
			cg.generateForInStatement(s)
		}
	case *ast.TryStatement:
		if s != nil {
			cg.generateTryStatement(s)
		}
	case *ast.ThrowStatement:
		if s != nil {
			cg.generateThrowStatement(s)
		}
	case *ast.BlockStatement:
		if s != nil {
			cg.generateBlockStatement(s)
		}
	case *ast.BreakStatement:
		if s != nil {
			cg.generateBreakStatement(s)
		}
	case *ast.ContinueStatement:
		if s != nil {
			cg.generateContinueStatement(s)
		}
	case *ast.ClassStatement:
		if s != nil {
			cg.generateClassStatement(s)
		}
	default:
		// TODO: Manejar otros tipos de sentencias.
		cg.writeString(fmt.Sprintf("// TODO: Sentencia no soportada: %T\n", s))
	}
}

// generateWhileStatement genera código Go para una sentencia 'while'.
func (cg *CodeGenerator) generateWhileStatement(stmt *ast.WhileStatement) {
	cg.writeString("for ")
	cg.generateExpression(stmt.Condition)
	cg.writeString(" {\n")
	cg.indent()

	if stmt.Body != nil {
		for _, bodyStmt := range stmt.Body.Statements {
			cg.generateStatement(bodyStmt)
		}
	}

	cg.dedent()
	cg.writeString("}\n")
}

// generateVarStatement genera código Go para una declaración de variable.
func (cg *CodeGenerator) generateVarStatement(stmt *ast.VarStatement) {
	cg.writeString(fmt.Sprintf("var %s ", stmt.Name.Value))
	if stmt.Value != nil {
		cg.writeString("= ")
		cg.generateExpression(stmt.Value)
	} else {
		// Si no hay valor, declarar con el tipo cero (implícito en Go).
		// Para Zylo, podríamos necesitar un tipo explícito o inferir "any".
		// Por ahora, dejamos que Go maneje el tipo cero.
	}
	cg.writeString("\n")
}

// generateExpressionStatement genera código Go para una sentencia de expresión.
func (cg *CodeGenerator) generateExpressionStatement(stmt *ast.ExpressionStatement) {
	// CRÍTICO: Verificar que stmt y stmt.Expression no sean nil
	if stmt == nil || stmt.Expression == nil {
		return // No generar nada si no hay expresión
	}
	if _, ok := stmt.Expression.(*ast.NumberLiteral); ok {
		return // Skip standalone number literals
	}
	cg.generateExpression(stmt.Expression)
	cg.writeString("\n")
}

// generateFuncStatement genera código Go para una declaración de función.
func (cg *CodeGenerator) generateFuncStatement(stmt *ast.FuncStatement) {
	if stmt == nil || stmt.Name == nil {
		return
	}

	cg.writeString(fmt.Sprintf("func %s(", stmt.Name.Value))

	// Generar parámetros
	for i, param := range stmt.Parameters {
		if i > 0 {
			cg.writeString(", ")
		}
		if param != nil {
			cg.writeString(fmt.Sprintf("%s int", param.Value))
		}
	}

	// Generar tipo de retorno
		returnType := " int"
		if stmt.ReturnType != "" {
			returnType = fmt.Sprintf(" %s", stmt.ReturnType)
		}

	cg.writeString(fmt.Sprintf(")%s {\n", returnType))
	cg.indent()

	// Generar cuerpo de la función
	if stmt.Body != nil {
		for _, bodyStmt := range stmt.Body.Statements {
			if bodyStmt != nil {
				cg.generateStatement(bodyStmt)
			}
		}
	}

	cg.dedent()
	cg.writeString("}\n")
}

// generateExpression genera código Go para una expresión del AST.
func (cg *CodeGenerator) generateExpression(exp ast.Expression) {
	if exp == nil {
		cg.writeString("// TODO: Expresión no soportada: <nil>")
		return
	}

	switch e := exp.(type) {
	case *ast.Identifier:
		if e.Value == "HASH_LITERAL" {
			cg.writeString("make(map[string]interface{})")
		} else {
			cg.writeString(e.Value)
		}
	case *ast.StringLiteral:
		cg.writeString(fmt.Sprintf("%q", e.Value))
	case *ast.NumberLiteral:
		cg.writeString(fmt.Sprintf("%d", e.Value)) // Asumiendo int64 por ahora
	case *ast.BooleanLiteral:
		if e.Value {
			cg.writeString("true")
		} else {
			cg.writeString("false")
		}
	case *ast.ListLiteral:
		cg.writeString("zyloruntime.NewList()")
		for _, element := range e.Elements {
			cg.writeString(".Append(")
			cg.generateExpression(element)
			cg.writeString(")")
		}
	case *ast.HashLiteral:
		cg.writeString("make(map[string]interface{})")
	case *ast.ClassInstantiation:
		// For now, treat as a function call to the class constructor
		if e.ClassName != nil {
			cg.writeString(fmt.Sprintf("New%s(", e.ClassName.Value))
			for i, arg := range e.Arguments {
				cg.generateExpression(arg)
				if i < len(e.Arguments)-1 {
					cg.writeString(", ")
				}
			}
			cg.writeString(")")
		}
	case *ast.CallExpression:
		// Manejar funciones especiales del runtime
		if ident, ok := e.Function.(*ast.Identifier); ok {
			switch ident.Value {
			case "show.log":
				cg.writeString("fmt.Println(") // Usar fmt.Println
				for i, arg := range e.Arguments {
					if arg != nil {
						cg.generateExpression(arg)
					}
					if i < len(e.Arguments)-1 {
						cg.writeString(", ")
					}
				}
				cg.writeString(")")
			case "read.line":
				cg.writeString("fmt.Scanln()")
			case "read.int":
				cg.writeString("fmt.Scanf(\"%d\")")
			default:
				// Si es una llamada a una función definida por el usuario, simplemente llamarla.
				oldIndent := cg.indentation
				cg.indentation = 0
				cg.generateExpression(e.Function)
				cg.writeString("(")
				for i, arg := range e.Arguments {
					if arg != nil {
						cg.generateExpression(arg)
					}
					if i < len(e.Arguments)-1 {
						cg.writeString(", ")
					}
				}
				cg.writeString(")")
				cg.indentation = oldIndent
			}
		} else if member, ok := e.Function.(*ast.MemberExpression); ok {
			// Handle member access calls like obj.method(...)
			if member.Object != nil && member.Property != nil {
				// Generate as obj.method(...)
				oldIndent := cg.indentation
				cg.indentation = 0
				cg.generateExpression(e.Function)
				cg.writeString("(")
				for i, arg := range e.Arguments {
					if arg != nil {
						cg.generateExpression(arg)
					}
					if i < len(e.Arguments)-1 {
						cg.writeString(", ")
					}
				}
				cg.writeString(")")
				cg.indentation = oldIndent
			}
		} else {
			// Check if it's a class instantiation
			if ident, ok := e.Function.(*ast.Identifier); ok {
				for _, class := range cg.classNames {
					if ident.Value == class {
						cg.writeString(fmt.Sprintf("New%s(", class))
						for i, arg := range e.Arguments {
							if arg != nil {
								cg.generateExpression(arg)
							}
							if i < len(e.Arguments)-1 {
								cg.writeString(", ")
							}
						}
						cg.writeString(")")
						return
					}
				}
			}

			// Handle other function calls
			cg.generateExpression(e.Function)
			cg.writeString("(")
			for i, arg := range e.Arguments {
				if arg != nil {
					cg.generateExpression(arg)
				}
				if i < len(e.Arguments)-1 {
					cg.writeString(", ")
				}
			}
			cg.writeString(")")
		}
	case *ast.InfixExpression:
		cg.generateInfixExpression(e)
	case *ast.MemberExpression:
		// Handle special cases like show.log()
		if e.Object != nil && e.Property != nil {
			if objId, ok := e.Object.(*ast.Identifier); ok && objId.Value == "show" {
				if e.Property.Value == "log" {
					cg.writeString("fmt.Println")
					return
				}
			}
		}

		// Generate member expression without intermediate indentation
		oldIndent := cg.indentation
		cg.indentation = 0
		if e.Object != nil {
			cg.generateExpression(e.Object)
		}
		cg.writeString(".")
		if e.Property != nil {
			cg.writeString(e.Property.Value)
		}
		cg.indentation = oldIndent
	case *ast.ThisExpression:
		cg.generateThisExpression(e)
	default:
		// TODO: Manejar otros tipos de expresiones.
		cg.writeString(fmt.Sprintf("// TODO: Expresión no soportada: %T", e))
	}
}

// generateReturnStatement genera código Go para una sentencia de retorno.
func (cg *CodeGenerator) generateReturnStatement(stmt *ast.ReturnStatement) {
	cg.writeString("return")
	if stmt.ReturnValue != nil {
		cg.writeString(" ")
		cg.generateExpression(stmt.ReturnValue)
	}
	cg.writeString("\n")
}

// generateIfStatement genera código Go para una sentencia 'if'.
func (cg *CodeGenerator) generateIfStatement(stmt *ast.IfStatement) {
	cg.writeString("if ")
	cg.generateExpression(stmt.Condition)
	cg.writeString(" {\n")
	cg.indent()

	if stmt.Consequence != nil {
		for _, bodyStmt := range stmt.Consequence.Statements {
			cg.generateStatement(bodyStmt)
		}
	}

	cg.dedent()
	cg.writeString("}")

	if stmt.Alternative != nil {
		cg.writeString(" else {\n")
		cg.indent()

		for _, bodyStmt := range stmt.Alternative.Statements {
			cg.generateStatement(bodyStmt)
		}

		cg.dedent()
		cg.writeString("}")
	}
	cg.writeString("\n")
}

// generateTryStatement genera código Go para una sentencia 'try-catch'.
func (cg *CodeGenerator) generateTryStatement(stmt *ast.TryStatement) {
	cg.writeString("zyloruntime.Try(func() {\n")
	cg.indent()

	if stmt.TryBlock != nil {
		for _, bodyStmt := range stmt.TryBlock.Statements {
			cg.generateStatement(bodyStmt)
		}
	}

	cg.dedent()
	cg.writeString("}, func(err error) {\n")
	cg.indent()

	if stmt.CatchClause != nil && stmt.CatchClause.CatchBlock != nil {
		// Declarar la variable de error si hay un parámetro
		if stmt.CatchClause.Parameter != nil {
			cg.writeString(fmt.Sprintf("var %s = err\n", stmt.CatchClause.Parameter.Value)) // Usar param.Value
		}

		for _, bodyStmt := range stmt.CatchClause.CatchBlock.Statements {
			cg.generateStatement(bodyStmt)
		}
	}

	cg.dedent()
	cg.writeString("})\n")

	// Manejar finally block si existe
	if stmt.FinallyBlock != nil {
		cg.writeString("defer func() {\n")
		cg.indent()

		for _, bodyStmt := range stmt.FinallyBlock.Statements {
			cg.generateStatement(bodyStmt)
		}

		cg.dedent()
		cg.writeString("}()\n")
	}
}

// generateThrowStatement genera código Go para una sentencia 'throw'.
func (cg *CodeGenerator) generateThrowStatement(stmt *ast.ThrowStatement) {
	cg.writeString("zyloruntime.Throw(")
	if stmt.Exception != nil {
		cg.generateExpression(stmt.Exception)
	} else {
		cg.writeString("\"\"")
	}
	cg.writeString(")\n")
}

// generateBlockStatement genera código Go para un bloque de sentencias.
func (cg *CodeGenerator) generateBlockStatement(stmt *ast.BlockStatement) {
	cg.writeString("{\n")
	cg.indent()

	for _, bodyStmt := range stmt.Statements {
		cg.generateStatement(bodyStmt)
	}

	cg.dedent()
	cg.writeString("}\n")
}

// writeString escribe una cadena en la salida con la indentación actual.
func (cg *CodeGenerator) writeString(s string) {
	if len(s) > 0 && s[0] != '\n' { // No indentar si la línea comienza con un salto de línea
		for i := 0; i < cg.indentation; i++ {
			cg.output.WriteString("    ") // 4 espacios por nivel de indentación
		}
	}
	cg.output.WriteString(s)
}

// indent aumenta el nivel de indentación.
func (cg *CodeGenerator) indent() {
	cg.indentation++
}

// dedent disminuye el nivel de indentación.
func (cg *CodeGenerator) dedent() {
	if cg.indentation > 0 {
		cg.indentation--
	}
}

// generateClassStatement genera código Go para una declaración de clase.
func (cg *CodeGenerator) generateClassStatement(stmt *ast.ClassStatement) {
	if stmt.Name == nil {
		return
	}

	className := stmt.Name.Value

	// Generate struct definition
	cg.writeString(fmt.Sprintf("type %s struct {\n", className))
	cg.indent()

	// Generate attributes
	for _, attr := range stmt.Attributes {
		if attr.Name != nil {
			cg.writeString(fmt.Sprintf("%s interface{}\n", attr.Name.Value))
		}
	}

	cg.dedent()
	cg.writeString("}\n\n")

	// Generate constructor function
		cg.writeString(fmt.Sprintf("func New%s(", className))
		// Add parameters for init method if it exists
		if stmt.InitMethod != nil && len(stmt.InitMethod.Parameters) > 0 {
			for i, param := range stmt.InitMethod.Parameters {
				if i > 0 {
					cg.writeString(", ")
				}
				cg.writeString(fmt.Sprintf("%s int", param.Value))
			}
		}
	cg.writeString(fmt.Sprintf(") *%s {\n", className))
	cg.indent()

	cg.writeString(fmt.Sprintf("obj := &%s{}\n", className))

	// Call init method if it exists
	if stmt.InitMethod != nil {
		cg.writeString(fmt.Sprintf("obj.%s(", stmt.InitMethod.Name.Value))
		if len(stmt.InitMethod.Parameters) > 0 {
			for i, param := range stmt.InitMethod.Parameters {
				if i > 0 {
					cg.writeString(", ")
				}
				cg.writeString(param.Value)
			}
		}
		cg.writeString(")\n")
	}

	cg.writeString("return obj\n")
	cg.dedent()
	cg.writeString("}\n\n")

	// Generate methods
	for _, method := range stmt.Methods {
		if method.Name == nil {
			continue
		}

		cg.writeString(fmt.Sprintf("func (obj *%s) %s(", className, method.Name.Value))

		// Add parameters
				for i, param := range method.Parameters {
					if i > 0 {
						cg.writeString(", ")
					}
					cg.writeString(fmt.Sprintf("%s int", param.Value))
				}

		cg.writeString(")")

		// Add return type if specified
		if method.ReturnType != "" {
			cg.writeString(fmt.Sprintf(" %s", method.ReturnType))
		}

		cg.writeString(" {\n")
		cg.indent()

		// Generate method body
		if method.Body != nil {
			for _, bodyStmt := range method.Body.Statements {
				cg.generateStatement(bodyStmt)
			}
		}

		cg.dedent()
		cg.writeString("}\n\n")
	}
}

// generateInfixExpression genera código Go para una expresión infija
func (cg *CodeGenerator) generateInfixExpression(exp *ast.InfixExpression) {
	// Manejar expresiones infijas con operadores correctos
	if exp == nil {
		return
	}

	if exp.Operator == "=" {
		if exp.Left != nil {
			cg.generateExpression(exp.Left)
		}
		cg.writeString(" = ")
		if exp.Right != nil {
			cg.generateExpression(exp.Right)
		}
	} else {
		if exp.Operator == "+" {
			cg.writeString("fmt.Sprintf(\"%v%v\", ")
			if exp.Left != nil {
				cg.generateExpression(exp.Left)
			} else {
				cg.writeString("nil")
			}
			cg.writeString(", ")
			if exp.Right != nil {
				cg.generateExpression(exp.Right)
			} else {
				cg.writeString("nil")
			}
			cg.writeString(")")
		} else {
			if exp.Left != nil {
				cg.generateExpression(exp.Left)
			} else {
				cg.writeString("nil")
			}

			// Convertir operadores de Zylo a Go
			switch exp.Operator {
			case "==":
				cg.writeString(" == ")
			case "!=":
				cg.writeString(" != ")
			case "-":
				cg.writeString(" - ")
			case "*":
				cg.writeString(" * ")
			case "/":
				cg.writeString(" / ")
			case "<":
				cg.writeString(" < ")
			case ">":
				cg.writeString(" > ")
			case "<=":
				cg.writeString(" <= ")
			case ">=":
				cg.writeString(" >= ")
			case "&&":
				cg.writeString(" && ")
			case "||":
				cg.writeString(" || ")
			default:
				cg.writeString(" " + exp.Operator + " ")
			}

			if exp.Right != nil {
				cg.generateExpression(exp.Right)
			} else {
				cg.writeString("nil")
			}
		}
	}
}

// generateThisExpression genera código Go para una expresión 'this'
func (cg *CodeGenerator) generateThisExpression(exp *ast.ThisExpression) {
	cg.writeString("obj")
}
