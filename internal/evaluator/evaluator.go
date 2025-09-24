package evaluator

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/zylo-lang/zylo/internal/ast"
)

// ZyloObject representa un objeto en tiempo de ejecución de Zylo
type ZyloObject interface {
	Type() string
	Inspect() string
}

// String representa un objeto string
type String struct {
	Value string
}

func (s *String) Type() string    { return "STRING_OBJ" }
func (s *String) Inspect() string { return s.Value }

// Integer representa un objeto integer
type Integer struct {
	Value int64
}

func (i *Integer) Type() string    { return "INTEGER_OBJ" }
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

// Float representa un objeto float
type Float struct {
	Value float64
}

func (f *Float) Type() string    { return "FLOAT_OBJ" }
func (f *Float) Inspect() string { return fmt.Sprintf("%g", f.Value) }

// List representa un objeto list
type List struct {
	Items []Value
}

func (l *List) Type() string { return "LIST_OBJ" }
func (l *List) Inspect() string {
	parts := make([]string, len(l.Items))
	for i, el := range l.Items {
		if obj, ok := el.(ZyloObject); ok {
			parts[i] = obj.Inspect()
		} else {
			parts[i] = fmt.Sprintf("%v", el)
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
// Hash representa un objeto hash
type Hash struct {
	Pairs map[string]Value
}

func (h *Hash) Type() string { return "HASH_OBJ" }
func (h *Hash) Inspect() string {
	var pairs []string
	for key, value := range h.Pairs {
		if obj, ok := value.(ZyloObject); ok {
			pairs = append(pairs, fmt.Sprintf("%s: %s", key, obj.Inspect()))
		} else {
			pairs = append(pairs, fmt.Sprintf("%s: %v", key, value))
		}
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

// Boolean representa un objeto boolean

// Boolean representa un objeto boolean
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() string    { return "BOOLEAN_OBJ" }
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// Null representa un objeto null
type Null struct{}

func (n *Null) Type() string    { return "NULL_OBJ" }
func (n *Null) Inspect() string { return "null" }

// Value representa un valor en tiempo de ejecución de Zylo
type Value interface{}

// Environment representa el entorno de ejecución con variables
type Environment struct {
	variables map[string]Value
	parent    *Environment
}

// NewEnvironment crea un nuevo entorno
func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		parent:    nil,
	}
}

// NewChildEnvironment crea un entorno hijo
func (e *Environment) NewChildEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		parent:    e,
	}
}

// Get obtiene el valor de una variable
func (e *Environment) Get(name string) (Value, bool) {
	if value, exists := e.variables[name]; exists {
		return value, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set establece el valor de una variable
func (e *Environment) Set(name string, value Value) {
	e.variables[name] = value
}

// Evaluator evalúa expresiones y sentencias de Zylo
type Evaluator struct {
	env    *Environment
	reader *bufio.Reader
}

// NewEvaluator crea un nuevo evaluador
func NewEvaluator() *Evaluator {
	eval := &Evaluator{
		env:    NewEnvironment(),
		reader: bufio.NewReader(os.Stdin),
	}
	eval.InitBuiltins()
	return eval
}

// InitBuiltins inicializa las funciones incorporadas
func (e *Evaluator) InitBuiltins() {
	// Mostrar por consola
	e.env.Set("show.log", &BuiltinFunction{
		Name: "show.log",
		Fn: func(args []Value) (Value, error) {
			for _, arg := range args {
				if obj, ok := arg.(ZyloObject); ok {
					fmt.Print(obj.Inspect(), " ")
				} else {
					fmt.Print(arg, " ")
				}
			}
			fmt.Println()
			os.Stdout.Sync() // Force flush after printing
			return &Null{}, nil
		},
	})

	// Leer una línea como string
	e.env.Set("read.line", &BuiltinFunction{
		Name: "read.line",
		Fn: func(args []Value) (Value, error) {
			fmt.Print("> ")  // Mostrar prompt
			os.Stdout.Sync() // Force flush
			input, err := e.reader.ReadString('\n')
			if err != nil {
				fmt.Println("⚠️  No se pudo leer entrada, usando valor vacío")
				return &String{Value: ""}, nil
			}
			return &String{Value: strings.TrimSpace(input)}, nil
		},
	})

	// Leer un número entero
	e.env.Set("read.int", &BuiltinFunction{
		Name: "read.int",
		Fn: func(args []Value) (Value, error) {
			for { // Loop until valid input is received
				fmt.Print("> ")  // Mostrar prompt
				os.Stdout.Sync() // Force flush
				input, err := e.reader.ReadString('\n')
				if err != nil {
					fmt.Println("⚠️  No se pudo leer entrada, usando 0 por defecto")
					return &Integer{Value: 0}, nil // Still return 0 on read error, as per original logic
				}
				input = strings.TrimSpace(input)
				n, err := strconv.Atoi(input)
				if err != nil {
					fmt.Println("❌ Error: no es un número válido, por favor intenta de nuevo.")
					// Continue the loop to re-prompt
				} else {
					// Valid input received, return the integer
					return &Integer{Value: int64(n)}, nil
				}
			}
		},
	})

	// Convertir número a string
	e.env.Set("string", &BuiltinFunction{
		Name: "string",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("string() espera exactamente 1 argumento, recibió %d", len(args))
			}

			switch arg := args[0].(type) {
			case *Integer:
				return &String{Value: fmt.Sprintf("%d", arg.Value)}, nil
			case *Float:
				return &String{Value: fmt.Sprintf("%g", arg.Value)}, nil
			case *String:
				return arg, nil
			case *Boolean:
				return &String{Value: fmt.Sprintf("%t", arg.Value)}, nil
			default:
				return &String{Value: fmt.Sprintf("%v", arg)}, nil
			}
		},
	})

	// len function
	e.env.Set("len", &BuiltinFunction{
		Name: "len",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("len() expects exactly 1 argument")
			}
			switch arg := args[0].(type) {
			case *List:
				return &Integer{Value: int64(len(arg.Items))}, nil
			case *String:
				return &Integer{Value: int64(len(arg.Value))}, nil
			default:
				return nil, fmt.Errorf("len() not supported for %T", arg)
			}
		},
	})

	// split function
	e.env.Set("split", &BuiltinFunction{
		Name: "split",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("split() expects exactly 2 arguments")
			}
			str, ok := args[0].(*String)
			if !ok {
				return nil, fmt.Errorf("first argument to split() must be string")
			}
			sep, ok := args[1].(*String)
			if !ok {
				return nil, fmt.Errorf("second argument to split() must be string")
			}
			parts := strings.Split(str.Value, sep.Value)
			items := make([]Value, len(parts))
			for i, part := range parts {
				items[i] = &String{Value: part}
			}
			return &List{Items: items}, nil
		},
	})

	// to_number function
	e.env.Set("to_number", &BuiltinFunction{
		Name: "to_number",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("to_number() expects exactly 1 argument")
			}
			str, ok := args[0].(*String)
			if !ok {
				return nil, fmt.Errorf("argument to to_number() must be string")
			}
			if val, err := strconv.ParseFloat(str.Value, 64); err == nil {
				return &Float{Value: val}, nil
			}
			if val, err := strconv.Atoi(str.Value); err == nil {
				return &Integer{Value: int64(val)}, nil
			}
			return &Null{}, nil
		},
	})

	// zyloruntime.Split function
	e.env.Set("zyloruntime.Split", &BuiltinFunction{
		Name: "zyloruntime.Split",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("zyloruntime.Split() expects exactly 2 arguments")
			}
			str, ok := args[0].(*String)
			if !ok {
				return nil, fmt.Errorf("first argument to zyloruntime.Split() must be string")
			}
			sep, ok := args[1].(*String)
			if !ok {
				return nil, fmt.Errorf("second argument to zyloruntime.Split() must be string")
			}
			parts := strings.Split(str.Value, sep.Value)
			items := make([]Value, len(parts))
			for i, part := range parts {
				items[i] = &String{Value: part}
			}
			return &List{Items: items}, nil
		},
	})

	// try function
	e.env.Set("try", &BuiltinFunction{
		Name: "try",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("try() expects exactly 2 arguments")
			}
			funcBlock, ok := args[0].(*ZyloFunction)
			if !ok {
				return nil, fmt.Errorf("first argument to try() must be function")
			}
			catchBlock, ok := args[1].(*ZyloFunction)
			if !ok {
				return nil, fmt.Errorf("second argument to try() must be function")
			}
			// Call funcBlock
			result, err := e.callZyloFunction(funcBlock, []Value{})
			if err != nil {
				// Call catchBlock with error as string
				errorStr := &String{Value: err.Error()}
				_, catchErr := e.callZyloFunction(catchBlock, []Value{errorStr})
				if catchErr != nil {
					return nil, catchErr
				}
				return &Null{}, nil
			}
			return result, nil
		},
	})

	// zyloruntime namespace (dummy object)
	e.env.Set("zyloruntime", &String{Value: "zyloruntime_namespace"})

	// getInput function
	e.env.Set("getInput", &BuiltinFunction{
		Name: "getInput",
		Fn: func(args []Value) (Value, error) {
			fmt.Print("> ")
			os.Stdout.Sync()
			input, err := e.reader.ReadString('\n')
			if err != nil {
				return &String{Value: ""}, nil
			}
			return &String{Value: strings.TrimSpace(input)}, nil
		},
	})

	// null value
	e.env.Set("null", &Null{})

	// Math functions for calculator
	e.env.Set("add", &BuiltinFunction{
		Name: "add",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("add() expects exactly 2 arguments")
			}
			return e.applyOperator("+", args[0], args[1])
		},
	})

	e.env.Set("subtract", &BuiltinFunction{
		Name: "subtract",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("subtract() expects exactly 2 arguments")
			}
			return e.applyOperator("-", args[0], args[1])
		},
	})

	e.env.Set("multiply", &BuiltinFunction{
		Name: "multiply",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("multiply() expects exactly 2 arguments")
			}
			return e.applyOperator("*", args[0], args[1])
		},
	})

	e.env.Set("divide", &BuiltinFunction{
		Name: "divide",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("divide() expects exactly 2 arguments")
			}
			return e.applyOperator("/", args[0], args[1])
		},
	})
}

// EvaluateProgram evalúa un programa completo
func (e *Evaluator) EvaluateProgram(program *ast.Program) error {
	for _, stmt := range program.Statements {
		_, err := e.evaluateStatement(stmt)
		if err != nil {
			return err
		}
	}

	// Execute main function if it exists
	mainFunc, exists := e.env.Get("main")
	if exists {
		if fn, ok := mainFunc.(*ZyloFunction); ok {
			_, err := e.callZyloFunction(fn, []Value{})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// evaluateStatement evalúa una sentencia
func (e *Evaluator) evaluateStatement(stmt ast.Statement) (Value, error) {
	if stmt == nil {
		return nil, fmt.Errorf("nil statement")
	}
	switch s := stmt.(type) {
	case *ast.VarStatement:
		if s == nil {
			return nil, fmt.Errorf("nil var statement")
		}
		err := e.evaluateVarStatement(s)
		if err != nil {
			return nil, err
		}
		return &Null{}, nil // Var statements don't return a value
	case *ast.ExpressionStatement:
		if s == nil {
			return &Null{}, nil // Skip nil expression statements
		}
		if s.Expression == nil {
			return &Null{}, nil // Skip statements with nil expressions
		}
		return e.evaluateExpression(s.Expression)
	case *ast.FuncStatement:
		if s == nil {
			return nil, fmt.Errorf("nil func statement")
		}
		err := e.evaluateFuncStatement(s)
		if err != nil {
			return nil, err
		}
		return &Null{}, nil // Func statements don't return a value
	case *ast.ReturnStatement:
		if s == nil {
			return nil, fmt.Errorf("nil return statement")
		}
		// If a return statement is encountered directly, evaluate its value and return it.
		// This assumes it's handled correctly by the caller (e.g., evaluateBlockStatement).
		if s.ReturnValue != nil {
			return e.evaluateExpression(s.ReturnValue)
		}
		return &Null{}, nil // Return Null if return statement has no value
	case *ast.IfStatement:
		if s == nil {
			return nil, fmt.Errorf("nil if statement")
		}
		return e.evaluateIfStatement(s) // This already returns Value, error (after the fix)
	case *ast.WhileStatement:
		if s == nil {
			return nil, fmt.Errorf("nil while statement")
		}
		return e.evaluateWhileStatement(s)
	case *ast.ForInStatement:
		return e.evaluateForInStatement(s)
	case *ast.BreakStatement:
		return e.evaluateBreakStatement(s)
	case *ast.ContinueStatement:
		return e.evaluateContinueStatement(s)
	case *ast.ClassStatement:
		return e.evaluateClassStatement(s)
	case *ast.TryStatement:
		// TODO: Implement try-catch to return Value, error
		return e.evaluateTryStatement(s) // Assuming this will be fixed to return Value, error
	case *ast.ThrowStatement:
		// TODO: Implement throw to return Value, error
		return e.evaluateThrowStatement(s) // Assuming this will be fixed to return Value, error
	case *ast.ImportStatement:
		return e.evaluateImportStatement(s)
	case *ast.BlockStatement:
		return e.evaluateBlockStatement(s) // This needs to return Value, error
	default:
		return nil, fmt.Errorf("sentencia no soportada: %T", s)
	}
}

// evaluateVarStatement evalúa una declaración de variable
func (e *Evaluator) evaluateVarStatement(stmt *ast.VarStatement) error {
	var value Value
	var err error
	if stmt.Value != nil {
		value, err = e.evaluateExpression(stmt.Value)
		if err != nil {
			return err
		}
	} else {
		value = &Null{}
	}
	e.env.Set(stmt.Name.Value, value)
	return nil
}

// evaluateFuncStatement evalúa una declaración de función
func (e *Evaluator) evaluateFuncStatement(stmt *ast.FuncStatement) error {
	// Crear una función Zylo
	zyloFunc := &ZyloFunction{
		Name:       stmt.Name.Value,
		Parameters: stmt.Parameters,
		Body:       stmt.Body,
		Env:        e.env,
	}
	e.env.Set(stmt.Name.Value, zyloFunc)
	return nil
}

// evaluateIfStatement evalúa una sentencia if
func (e *Evaluator) evaluateIfStatement(stmt *ast.IfStatement) (Value, error) {
	condition, err := e.evaluateExpression(stmt.Condition)
	if err != nil {
		return nil, err
	}

	if e.isTruthy(condition) {
		// Evaluar el bloque 'consequence' (el 'if' principal)
		return e.evaluateBlockStatement(stmt.Consequence)
	} else if stmt.Alternative != nil {
		// Si la condición principal es falsa, verificar el tipo de la alternativa.
		// Based on the compiler error, stmt.Alternative is likely always a *ast.BlockStatement.
		// The original code had a switch that handled both *ast.BlockStatement and *ast.IfStatement.
		// The error message suggests *ast.BlockStatement is not an interface, implying stmt.Alternative
		// is a concrete type. If it's always *ast.BlockStatement, then we only need to handle that.

		// Since Alternative is always *ast.BlockStatement or nil, we can directly call evaluateBlockStatement
		return e.evaluateBlockStatement(stmt.Alternative)
	}

	// Si ninguna condición se cumple y no hay alternativa, no hacer nada.
	return &Null{}, nil
}

// evaluateTryStatement evalúa una sentencia try-catch
func (e *Evaluator) evaluateTryStatement(stmt *ast.TryStatement) (Value, error) {
	// TODO: Implementar try-catch
	if stmt.TryBlock == nil {
		return nil, fmt.Errorf("nil try block")
	}
	return e.evaluateBlockStatement(stmt.TryBlock)
}

// evaluateThrowStatement evalúa una sentencia throw
func (e *Evaluator) evaluateThrowStatement(stmt *ast.ThrowStatement) (Value, error) {
	// TODO: Implementar throw
	return &Null{}, nil
}

// evaluateBlockStatement evalúa un bloque de sentencias
func (e *Evaluator) evaluateBlockStatement(stmt *ast.BlockStatement) (Value, error) {
	// Create a new environment for this block to isolate variable scope
	childEnv := e.env.NewChildEnvironment()
	oldEnv := e.env
	e.env = childEnv
	defer func() { e.env = oldEnv }()

	var lastValue Value = &Null{} // To store the value of the last expression

	for _, bodyStmt := range stmt.Statements {
		// Evaluate statement and capture its value
		value, err := e.evaluateStatement(bodyStmt)
		if err != nil {
			return nil, err // Propagate error
		}

		// Handle break and continue
		if _, ok := value.(*BreakValue); ok {
			return value, nil
		}
		if _, ok := value.(*ContinueValue); ok {
			return value, nil
		}

		// Capture the value for potential return
		lastValue = value
	}

	// Return the value of the last evaluated statement
	return lastValue, nil
}

// evaluateWhileStatement evalúa una sentencia while
func (e *Evaluator) evaluateWhileStatement(stmt *ast.WhileStatement) (Value, error) {
	for {
		condition, err := e.evaluateExpression(stmt.Condition)
		if err != nil {
			return nil, err
		}

		if !e.isTruthy(condition) {
			break
		}

		// Evaluate body statements directly without creating child environment
		for _, bodyStmt := range stmt.Body.Statements {
			value, err := e.evaluateStatement(bodyStmt)
			if err != nil {
				return nil, err
			}

			// Handle break and continue
			if _, ok := value.(*BreakValue); ok {
				return &Null{}, nil // Break out of while
			}
			if _, ok := value.(*ContinueValue); ok {
				break // Continue to next iteration
			}
		}
	}

	return &Null{}, nil
}

// evaluateForInStatement evalúa una sentencia for in
func (e *Evaluator) evaluateForInStatement(stmt *ast.ForInStatement) (Value, error) {
	iterable, err := e.evaluateExpression(stmt.Iterable)
	if err != nil {
		return nil, err
	}

	switch iter := iterable.(type) {
	case *List:
		for _, element := range iter.Items {
			e.env.Set(stmt.Identifier.Value, element)

			result, err := e.evaluateBlockStatement(stmt.Body)
			if err != nil {
				return nil, err
			}

			// Handle break and continue
			if _, ok := result.(*BreakValue); ok {
				break
			}
			if _, ok := result.(*ContinueValue); ok {
				continue
			}
		}
	case *String:
		for _, char := range iter.Value {
			e.env.Set(stmt.Identifier.Value, &String{Value: string(char)})

			result, err := e.evaluateBlockStatement(stmt.Body)
			if err != nil {
				return nil, err
			}

			// Handle break and continue
			if _, ok := result.(*BreakValue); ok {
				break
			}
			if _, ok := result.(*ContinueValue); ok {
				continue
			}
		}
	default:
		return nil, fmt.Errorf("cannot iterate over %T", iterable)
	}

	return &Null{}, nil
}

// evaluateImportStatement evalúa una declaración de import
func (e *Evaluator) evaluateImportStatement(stmt *ast.ImportStatement) (Value, error) {
	if stmt.ModuleName == nil {
		return nil, fmt.Errorf("import statement has nil module name")
	}

	moduleName := stmt.ModuleName.Value

	// Check if module is already imported - if so, ignore silently
	if _, exists := e.env.Get(moduleName); exists {
		return &Null{}, nil
	}

	// For now, we only support "zyloruntime" module
	if moduleName == "zyloruntime" {
		// Create a module object with available functions
		moduleObj := &String{Value: "zyloruntime_module"}
		e.env.Set(moduleName, moduleObj)
		return &Null{}, nil
	}

	return nil, fmt.Errorf("Module '%s' not found", moduleName)
}

// evaluateBreakStatement evalúa una sentencia break
func (e *Evaluator) evaluateBreakStatement(stmt *ast.BreakStatement) (Value, error) {
	return &BreakValue{}, nil
}

// evaluateContinueStatement evalúa una sentencia continue
func (e *Evaluator) evaluateContinueStatement(stmt *ast.ContinueStatement) (Value, error) {
	return &ContinueValue{}, nil
}

// evaluateClassStatement evalúa una declaración de clase
func (e *Evaluator) evaluateClassStatement(stmt *ast.ClassStatement) (Value, error) {
	// Create a class object
	classObj := &ZyloClass{
		Name:       stmt.Name.Value,
		Attributes: make(map[string]Value),
		Methods:    make(map[string]*ZyloFunction),
		InitMethod: nil,
	}

	// Set attributes
	for _, attr := range stmt.Attributes {
		if attr.Value != nil {
			value, err := e.evaluateExpression(attr.Value)
			if err != nil {
				return nil, err
			}
			classObj.Attributes[attr.Name.Value] = value
		} else {
			classObj.Attributes[attr.Name.Value] = &Null{}
		}
	}

	// Set methods
	for _, method := range stmt.Methods {
		zyloFunc := &ZyloFunction{
			Name:       method.Name.Value,
			Parameters: method.Parameters,
			Body:       method.Body,
			Env:        e.env,
		}
		classObj.Methods[method.Name.Value] = zyloFunc

		// Check if it's the init method
		if method.Name.Value == "init" {
			classObj.InitMethod = zyloFunc
		}
	}

	e.env.Set(stmt.Name.Value, classObj)
	return &Null{}, nil
}

// evaluateExpression evalúa una expresión
func (e *Evaluator) evaluateExpression(exp ast.Expression) (Value, error) {
	if exp == nil {
		return nil, fmt.Errorf("nil expression passed to evaluator")
	}

	switch ex := exp.(type) {
	case *ast.Identifier:
		if ex == nil {
			return nil, fmt.Errorf("nil identifier")
		}
		return e.evaluateIdentifier(ex)
	case *ast.StringLiteral:
		if ex == nil {
			return nil, fmt.Errorf("nil string literal")
		}
		return &String{Value: ex.Value}, nil
	case *ast.NumberLiteral:
		if ex == nil {
			return nil, fmt.Errorf("nil number literal")
		}
		if val, ok := ex.Value.(float64); ok {
			return &Float{Value: val}, nil
		} else if val, ok := ex.Value.(int64); ok {
			return &Integer{Value: val}, nil
		}
		return &Integer{Value: 0}, nil
	case *ast.BooleanLiteral:
		if ex == nil {
			return nil, fmt.Errorf("nil boolean literal")
		}
		return &Boolean{Value: ex.Value}, nil
	case *ast.NullLiteral:
		if ex == nil {
			return nil, fmt.Errorf("nil null literal")
		}
		return &Null{}, nil
	case *ast.CallExpression:
		if ex == nil {
			return nil, fmt.Errorf("nil call expression")
		}
		return e.evaluateCallExpression(ex)
	case *ast.MemberExpression:
		if ex == nil {
			return nil, fmt.Errorf("nil member expression")
		}
		return e.evaluateMemberExpression(ex)
	case *ast.ListLiteral:
		if ex == nil {
			return nil, fmt.Errorf("nil list literal")
		}
		elements := make([]Value, len(ex.Elements))
		for i, el := range ex.Elements {
			var err error
			elements[i], err = e.evaluateExpression(el)
			if err != nil {
				return nil, err
			}
		}
		return &List{Items: elements}, nil
	case *ast.HashLiteral:
		if ex == nil {
			return nil, fmt.Errorf("nil hash literal")
		}
		pairs := make(map[string]Value)
		for key, value := range ex.Pairs {
			keyVal, err := e.evaluateExpression(key)
			if err != nil {
				return nil, err
			}
			// For now, assume keys are strings
			keyStr, ok := keyVal.(*String)
			if !ok {
				return nil, fmt.Errorf("hash key must be string")
			}
			val, err := e.evaluateExpression(value)
			if err != nil {
				return nil, err
			}
			pairs[keyStr.Value] = val
		}
		return &Hash{Pairs: pairs}, nil
	case *ast.IndexExpression:
		if ex == nil {
			return nil, fmt.Errorf("nil index expression")
		}
		left, err := e.evaluateExpression(ex.Left)
		if err != nil {
			return nil, err
		}
		index, err := e.evaluateExpression(ex.Index)
		if err != nil {
			return nil, err
		}
		return e.indexValue(left, index)
	case *ast.InfixExpression:
		if ex == nil {
			return nil, fmt.Errorf("nil infix expression")
		}
		return e.evaluateInfixExpression(ex)
	case *ast.PrefixExpression:
		if ex == nil {
			return nil, fmt.Errorf("nil prefix expression")
		}
		return e.evaluatePrefixExpression(ex)
	case *ast.ThisExpression:
		if ex == nil {
			return nil, fmt.Errorf("nil this expression")
		}
		return e.evaluateThisExpression(ex)
	case *ast.ImportStatement:
		if ex == nil {
			return nil, fmt.Errorf("nil import statement")
		}
		return e.evaluateImportStatement(ex)
	default:
		return nil, fmt.Errorf("expresión no soportada: %T", ex)
	}
}

// evaluateIdentifier evalúa un identificador
func (e *Evaluator) evaluateIdentifier(exp *ast.Identifier) (Value, error) {
	value, exists := e.env.Get(exp.Value)
	if !exists {
		return nil, fmt.Errorf("variable no definida: %s", exp.Value)
	}
	return value, nil
}

// evaluateMemberExpression evalúa una expresión de acceso a miembro
func (e *Evaluator) evaluateMemberExpression(exp *ast.MemberExpression) (Value, error) {
	if exp.Object == nil {
		return nil, fmt.Errorf("nil object in member expression")
	}
	if exp.Property == nil {
		return nil, fmt.Errorf("nil property in member expression")
	}

	// Obtener el nombre de la propiedad/método
	propName := exp.Property.Value

	// Manejar casos específicos como 'show.log' y 'read.line' sin evaluar el objeto primero
	if identifier, ok := exp.Object.(*ast.Identifier); ok {
		objName := identifier.Value

		// Crear el nombre completo de la función built-in
		fullName := objName + "." + propName

		// Buscar en el entorno si existe una función built-in
		if builtin, exists := e.env.Get(fullName); exists {
			return builtin, nil
		}
	}

	// Para otros casos, evaluar el objeto primero
		obj, err := e.evaluateExpression(exp.Object)
		if err != nil {
			return nil, err
		}
		if obj == nil {
			return nil, fmt.Errorf("cannot access member on nil object")
		}

	// Handle zyloruntime namespace
	if identifier, ok := obj.(*ast.Identifier); ok && identifier.Value == "zyloruntime" {
		switch propName {
		case "Split":
			return &BuiltinFunction{
				Name: "zyloruntime.Split",
				Fn: func(args []Value) (Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("zyloruntime.Split() expects exactly 2 arguments")
					}
					str, ok := args[0].(*String)
					if !ok {
						return nil, fmt.Errorf("first argument to zyloruntime.Split() must be string")
					}
					sep, ok := args[1].(*String)
					if !ok {
						return nil, fmt.Errorf("second argument to zyloruntime.Split() must be string")
					}
					parts := strings.Split(str.Value, sep.Value)
					items := make([]Value, len(parts))
					for i, part := range parts {
						items[i] = &String{Value: part}
					}
					return &List{Items: items}, nil
				},
			}, nil
		default:
			return nil, fmt.Errorf("property '%s' not found on zyloruntime", propName)
		}
	}

	// Handle List methods
	if list, ok := obj.(*List); ok {
		switch propName {
		case "Get":
			return &BuiltinFunction{
				Name: "List.Get",
				Fn: func(args []Value) (Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("List.Get() expects exactly 1 argument")
					}
					idx, ok := args[0].(*Integer)
					if !ok {
						return nil, fmt.Errorf("List.Get() index must be integer")
					}
					if idx.Value < 0 || int(idx.Value) >= len(list.Items) {
						return nil, fmt.Errorf("index out of bounds")
					}
					return list.Items[idx.Value], nil
				},
			}, nil
		case "Append":
			return &BuiltinFunction{
				Name: "List.Append",
				Fn: func(args []Value) (Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("List.Append() expects exactly 1 argument")
					}
					list.Items = append(list.Items, args[0])
					return &Null{}, nil
				},
			}, nil
		case "Len":
			return &BuiltinFunction{
				Name: "List.Len",
				Fn: func(args []Value) (Value, error) {
					return &Integer{Value: int64(len(list.Items))}, nil
				},
			}, nil
		default:
			return nil, fmt.Errorf("method '%s' not found on List", propName)
		}
	}

	// Handle instance member access
	if instance, ok := obj.(*ZyloInstance); ok {
		if field, exists := instance.Fields[propName]; exists {
			return field, nil
		}
		if method, exists := instance.Class.Methods[propName]; exists {
			// Return a bound method
			return &BoundMethod{
				Instance: instance,
				Method:   method,
			}, nil
		}
		return nil, fmt.Errorf("property '%s' not found on instance of %s", propName, instance.Class.Name)
	}

	return nil, fmt.Errorf("cannot access property '%s' on %T", propName, obj)
}

// evaluateCallExpression evalúa una llamada a función
func (e *Evaluator) evaluateCallExpression(exp *ast.CallExpression) (Value, error) {
	if exp.Function == nil {
		return nil, fmt.Errorf("nil function in call expression")
	}

	// Evaluar la función
	fn, err := e.evaluateExpression(exp.Function)
	if err != nil {
		return nil, err
	}

	// Check if it's a class instantiation
	if class, ok := fn.(*ZyloClass); ok {
		return e.instantiateClass(class, exp.Arguments)
	}

	// Evaluar argumentos
	args := make([]Value, len(exp.Arguments))
	for i, arg := range exp.Arguments {
		if arg == nil {
			return nil, fmt.Errorf("nil argument at position %d in call expression", i)
		}
		args[i], err = e.evaluateExpression(arg)
		if err != nil {
			return nil, err
		}
	}

	// Llamar a la función
	return e.callFunction(fn, args)
}

// evaluateInfixExpression evalúa una expresión infija
func (e *Evaluator) evaluateInfixExpression(exp *ast.InfixExpression) (Value, error) {
	if exp.Left == nil {
		return nil, fmt.Errorf("nil left operand in infix expression")
	}
	if exp.Right == nil {
		return nil, fmt.Errorf("nil right operand in infix expression")
	}

	// Handle assignment operator
	if exp.Operator == "=" {
		// Check if left is an identifier
		leftIdent, ok := exp.Left.(*ast.Identifier)
		if !ok {
			return nil, fmt.Errorf("left side of assignment must be an identifier")
		}

		// Evaluate the right side
		right, err := e.evaluateExpression(exp.Right)
		if err != nil {
			return nil, err
		}

		// Check if the variable is declared (exists in environment)
		_, exists := e.env.Get(leftIdent.Value)
		if !exists {
			return nil, fmt.Errorf("variable no definida: %s", leftIdent.Value)
		}

		// Assign the value
		e.env.Set(leftIdent.Value, right)
		return right, nil
	}

	left, err := e.evaluateExpression(exp.Left)
	if err != nil {
		return nil, err
	}

	right, err := e.evaluateExpression(exp.Right)
	if err != nil {
		return nil, err
	}

	return e.applyOperator(exp.Operator, left, right)
}

// evaluatePrefixExpression evalúa una expresión prefija
func (e *Evaluator) evaluatePrefixExpression(exp *ast.PrefixExpression) (Value, error) {
	if exp.Right == nil {
		return nil, fmt.Errorf("nil right operand in prefix expression")
	}

	right, err := e.evaluateExpression(exp.Right)
	if err != nil {
		return nil, err
	}

	switch exp.Operator {
	case "!":
		return &Boolean{Value: !e.isTruthy(right)}, nil
	case "-":
		if num, ok := right.(*Integer); ok {
			return &Integer{Value: -num.Value}, nil
		}
		return nil, fmt.Errorf("operador '-' no soportado para tipo %T", right)
	default:
		return nil, fmt.Errorf("operador prefijo no soportado: %s", exp.Operator)
	}
}

// callFunction llama a una función
func (e *Evaluator) callFunction(fn Value, args []Value) (Value, error) {
	if fn == nil {
		return nil, fmt.Errorf("cannot call nil function")
	}
	switch f := fn.(type) {
	case *ZyloFunction:
		return e.callZyloFunction(f, args)
	case *BuiltinFunction:
		return f.Fn(args)
	case *BoundMethod:
		return e.callBoundMethod(f, args)
	default:
		// Intentar funciones built-in
		if ident, ok := fn.(*ast.Identifier); ok {
			return e.callBuiltinFunction(ident.Value, args)
		}
		return nil, fmt.Errorf("no se puede llamar a: %T", fn)
	}
}

// instantiateClass crea una instancia de una clase
func (e *Evaluator) instantiateClass(class *ZyloClass, args []ast.Expression) (Value, error) {
	instance := &ZyloInstance{
		Class:  class,
		Fields: make(map[string]Value),
	}

	// Copy class attributes to instance
	for name, value := range class.Attributes {
		instance.Fields[name] = value
	}

	// Call init method if it exists
	if class.InitMethod != nil {
		// Evaluate arguments
		evalArgs := make([]Value, len(args))
		for i, arg := range args {
			var err error
			evalArgs[i], err = e.evaluateExpression(arg)
			if err != nil {
				return nil, err
			}
		}

		// Create instance environment
		funcEnv := class.InitMethod.Env.NewChildEnvironment()
		funcEnv.Set("this", instance)

		// Set parameters
		for i, param := range class.InitMethod.Parameters {
			if i < len(evalArgs) {
				funcEnv.Set(param.Value, evalArgs[i])
			}
		}

		// Execute init method
		oldEnv := e.env
		e.env = funcEnv
		defer func() { e.env = oldEnv }()

		_, err := e.evaluateBlockStatement(class.InitMethod.Body)
		if err != nil {
			return nil, err
		}
	}

	return instance, nil
}

// callZyloFunction llama a una función Zylo
func (e *Evaluator) callZyloFunction(fn *ZyloFunction, args []Value) (Value, error) {
	// Crear entorno de función
	funcEnv := fn.Env.NewChildEnvironment()

	// Establecer parámetros
	for i, param := range fn.Parameters {
		if i < len(args) {
			funcEnv.Set(param.Value, args[i]) // Usar param.Value
		}
	}

	// Ejecutar cuerpo de la función
	oldEnv := e.env
	e.env = funcEnv
	defer func() { e.env = oldEnv }()

	result, err := e.evaluateBlockStatement(fn.Body) // Capture the result
	if err != nil {
		return nil, err // Propagate error
	}
	return result, nil // Return the captured result
}

// callBoundMethod llama a un método ligado
func (e *Evaluator) callBoundMethod(boundMethod *BoundMethod, args []Value) (Value, error) {
	// Crear entorno de método
	funcEnv := boundMethod.Method.Env.NewChildEnvironment()
	funcEnv.Set("this", boundMethod.Instance)

	// Establecer parámetros
	for i, param := range boundMethod.Method.Parameters {
		if i < len(args) {
			funcEnv.Set(param.Value, args[i])
		}
	}

	// Ejecutar cuerpo del método
	oldEnv := e.env
	e.env = funcEnv
	defer func() { e.env = oldEnv }()

	result, err := e.evaluateBlockStatement(boundMethod.Method.Body)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// evaluateThisExpression evalúa una expresión 'this'
func (e *Evaluator) evaluateThisExpression(exp *ast.ThisExpression) (Value, error) {
	// Buscar 'this' en el entorno actual
	value, exists := e.env.Get("this")
	if !exists {
		return nil, fmt.Errorf("'this' is not available in this context")
	}
	return value, nil
}

// callBuiltinFunction llama a una función built-in
func (e *Evaluator) callBuiltinFunction(name string, args []Value) (Value, error) {
	// Buscar la función en el entorno
	if builtin, exists := e.env.Get(name); exists {
		if fn, ok := builtin.(*BuiltinFunction); ok {
			return fn.Fn(args)
		}
	}
	return nil, fmt.Errorf("función no definida: %s", name)
}

// applyOperator aplica un operador binario
func (e *Evaluator) applyOperator(operator string, left, right Value) (Value, error) {
	// Check for nil operands
	if left == nil || right == nil {
		return nil, fmt.Errorf("cannot apply operator '%s' to nil values", operator)
	}

	switch operator {
	case "+":
		// Manejar concatenación de strings
		if leftStr, ok := left.(*String); ok {
			if rightStr, ok := right.(*String); ok {
				return &String{Value: leftStr.Value + rightStr.Value}, nil
			}
			// Convertir números a string para concatenación
			if rightNum, ok := right.(*Integer); ok {
				return &String{Value: leftStr.Value + fmt.Sprintf("%d", rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &String{Value: leftStr.Value + fmt.Sprintf("%g", rightFloat.Value)}, nil
			}
		}
		// Manejar concatenación string + número (orden inverso)
		if rightStr, ok := right.(*String); ok {
			if leftNum, ok := left.(*Integer); ok {
				return &String{Value: fmt.Sprintf("%d", leftNum.Value) + rightStr.Value}, nil
			}
			if leftFloat, ok := left.(*Float); ok {
				return &String{Value: fmt.Sprintf("%g", leftFloat.Value) + rightStr.Value}, nil
			}
		}
		// Operaciones numéricas
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Integer{Value: leftNum.Value + rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Float{Value: float64(leftNum.Value) + rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Float{Value: leftFloat.Value + float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Float{Value: leftFloat.Value + rightFloat.Value}, nil
			}
		}
	case "-":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Integer{Value: leftNum.Value - rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Float{Value: float64(leftNum.Value) - rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Float{Value: leftFloat.Value - float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Float{Value: leftFloat.Value - rightFloat.Value}, nil
			}
		}
	case "*":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Integer{Value: leftNum.Value * rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Float{Value: float64(leftNum.Value) * rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Float{Value: leftFloat.Value * float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Float{Value: leftFloat.Value * rightFloat.Value}, nil
			}
		}
	case "/":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				if rightNum.Value == 0 {
					return nil, fmt.Errorf("división por cero")
				}
				return &Integer{Value: leftNum.Value / rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				if rightFloat.Value == 0 {
					return nil, fmt.Errorf("división por cero")
				}
				return &Float{Value: float64(leftNum.Value) / rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				if rightNum.Value == 0 {
					return nil, fmt.Errorf("división por cero")
				}
				return &Float{Value: leftFloat.Value / float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				if rightFloat.Value == 0 {
					return nil, fmt.Errorf("división por cero")
				}
				return &Float{Value: leftFloat.Value / rightFloat.Value}, nil
			}
		}
	case "==":
		// Handle string comparison specifically
		if leftStr, ok := left.(*String); ok {
			if rightStr, ok := right.(*String); ok {
				return &Boolean{Value: leftStr.Value == rightStr.Value}, nil
			}
		}
		// Handle number comparison
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftNum.Value == rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: float64(leftNum.Value) == rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftFloat.Value == float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: leftFloat.Value == rightFloat.Value}, nil
			}
		}
		// Handle boolean comparison
		if leftBool, ok := left.(*Boolean); ok {
			if rightBool, ok := right.(*Boolean); ok {
				return &Boolean{Value: leftBool.Value == rightBool.Value}, nil
			}
		}
		return &Boolean{Value: false}, nil
	case "!=":
		// Handle string comparison specifically
		if leftStr, ok := left.(*String); ok {
			if rightStr, ok := right.(*String); ok {
				return &Boolean{Value: leftStr.Value != rightStr.Value}, nil
			}
		}
		// Handle number comparison
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftNum.Value != rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: float64(leftNum.Value) != rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftFloat.Value != float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: leftFloat.Value != rightFloat.Value}, nil
			}
		}
		// Handle boolean comparison
		if leftBool, ok := left.(*Boolean); ok {
			if rightBool, ok := right.(*Boolean); ok {
				return &Boolean{Value: leftBool.Value != rightBool.Value}, nil
			}
		}
		return &Boolean{Value: true}, nil
	case "<":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftNum.Value < rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: float64(leftNum.Value) < rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftFloat.Value < float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: leftFloat.Value < rightFloat.Value}, nil
			}
		}
	case ">":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftNum.Value > rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: float64(leftNum.Value) > rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftFloat.Value > float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: leftFloat.Value > rightFloat.Value}, nil
			}
		}
	case "<=":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftNum.Value <= rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: float64(leftNum.Value) <= rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftFloat.Value <= float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: leftFloat.Value <= rightFloat.Value}, nil
			}
		}
	case ">=":
		if leftNum, ok := left.(*Integer); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftNum.Value >= rightNum.Value}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: float64(leftNum.Value) >= rightFloat.Value}, nil
			}
		}
		if leftFloat, ok := left.(*Float); ok {
			if rightNum, ok := right.(*Integer); ok {
				return &Boolean{Value: leftFloat.Value >= float64(rightNum.Value)}, nil
			}
			if rightFloat, ok := right.(*Float); ok {
				return &Boolean{Value: leftFloat.Value >= rightFloat.Value}, nil
			}
		}
	case "&&":
		leftBool := e.isTruthy(left)
		if !leftBool {
			return &Boolean{Value: false}, nil
		}
		rightBool := e.isTruthy(right)
		return &Boolean{Value: rightBool}, nil
	case "||":
		leftBool := e.isTruthy(left)
		if leftBool {
			return &Boolean{Value: true}, nil
		}
		rightBool := e.isTruthy(right)
		return &Boolean{Value: rightBool}, nil
	}

	return nil, fmt.Errorf("operador '%s' no soportado para tipos %T y %T", operator, left, right)
}

// isTruthy determina si un valor es "verdadero"
func (e *Evaluator) isTruthy(value Value) bool {
	if value == nil {
		return false
	}
	if boolVal, ok := value.(*Boolean); ok {
		return boolVal.Value
	}
	if intVal, ok := value.(*Integer); ok {
		return intVal.Value != 0
	}
	if strVal, ok := value.(*String); ok {
		return len(strVal.Value) > 0
	}
	return true
}

// indexValue handles indexing for arrays and strings
func (e *Evaluator) indexValue(left, index Value) (Value, error) {
	if left == nil {
		return nil, fmt.Errorf("cannot index nil value")
	}
	switch l := left.(type) {
	case *List:
		idx, ok := index.(*Integer)
		if !ok {
			return nil, fmt.Errorf("list index must be integer")
		}
		if idx.Value < 0 || int(idx.Value) >= len(l.Items) {
			return nil, fmt.Errorf("index out of bounds")
		}
		return l.Items[idx.Value], nil
	case *String:
		idx, ok := index.(*Integer)
		if !ok {
			return nil, fmt.Errorf("string index must be integer")
		}
		if idx.Value < 0 || int(idx.Value) >= len(l.Value) {
			return nil, fmt.Errorf("index out of bounds")
		}
		return &String{Value: string(l.Value[idx.Value])}, nil
	default:
		return nil, fmt.Errorf("cannot index %T", left)
	}
}

// ZyloFunction representa una función definida en Zylo
type ZyloFunction struct {
	Name       string
	Parameters []*ast.Identifier // Cambiado de []*ast.Variable a []*ast.Identifier
	ReturnType string          // Nuevo campo para el tipo de retorno
	Body       *ast.BlockStatement
	Env        *Environment
}

// BuiltinFunction representa una función built-in
type BuiltinFunction struct {
	Name string
	Fn   func([]Value) (Value, error)
}

// ZyloClass representa una clase definida en Zylo
type ZyloClass struct {
	Name       string
	Attributes map[string]Value
	Methods    map[string]*ZyloFunction
	InitMethod *ZyloFunction
}

func (c *ZyloClass) Type() string { return "CLASS_OBJ" }
func (c *ZyloClass) Inspect() string {
	return fmt.Sprintf("class %s", c.Name)
}

// ZyloInstance representa una instancia de una clase Zylo
type ZyloInstance struct {
	Class  *ZyloClass
	Fields map[string]Value
}

func (i *ZyloInstance) Type() string { return "INSTANCE_OBJ" }
func (i *ZyloInstance) Inspect() string {
	return fmt.Sprintf("instance of %s", i.Class.Name)
}

// BoundMethod representa un método ligado a una instancia
type BoundMethod struct {
	Instance *ZyloInstance
	Method   *ZyloFunction
}

func (b *BoundMethod) Type() string { return "BOUND_METHOD_OBJ" }
func (b *BoundMethod) Inspect() string {
	return fmt.Sprintf("bound method %s", b.Method.Name)
}

// Control flow types for break and continue
type BreakValue struct{}
type ContinueValue struct{}

func (b *BreakValue) Type() string    { return "BREAK_OBJ" }
func (b *BreakValue) Inspect() string { return "break" }

func (c *ContinueValue) Type() string    { return "CONTINUE_OBJ" }
func (c *ContinueValue) Inspect() string { return "continue" }
