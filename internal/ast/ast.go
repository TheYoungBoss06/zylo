package ast

import (
	"fmt"

	"github.com/zylo-lang/zylo/internal/lexer"
)

// Node es la interfaz base para todos los nodos del AST.
type Node interface {
	TokenLiteral() string // Devuelve el literal del token asociado al nodo.
	String() string       // Devuelve una representación en string del nodo para debugging.
}

// Statement es una interfaz para todos los nodos de sentencia.
type Statement interface {
	Node
	statementNode() // Método marcador para identificar nodos de sentencia.
}

// Expression es una interfaz para todos los nodos de expresión.
type Expression interface {
	Node
	expressionNode() // Método marcador para identificar nodos de expresión.
}

// Program es el nodo raíz de todo AST de un programa Zylo.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out string
	for _, s := range p.Statements {
		out += s.String()
	}
	return out
}

// ImportStatement representa una declaración de import (e.g., import zyloruntime).
type ImportStatement struct {
	Token      lexer.Token // El token 'import'.
	ModuleName *Identifier // El nombre del módulo a importar.
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) expressionNode()      {} // Also implement Expression interface
func (is *ImportStatement) TokenLiteral() string { return is.Token.Lexeme }
func (is *ImportStatement) String() string {
	out := "import "
	if is.ModuleName != nil {
		out += is.ModuleName.String()
	}
	out += ";"
	return out
}

// VarStatement representa una declaración de variable (e.g., var x = 5;).
type VarStatement struct {
	Token lexer.Token // El token 'var'.
	Name  *Identifier
	Value Expression
}

func (vs *VarStatement) statementNode()       {}
func (vs *VarStatement) TokenLiteral() string { return vs.Token.Lexeme }
func (vs *VarStatement) String() string {
	var out string
	out += vs.TokenLiteral() + " "
	if vs.Name != nil {
		out += vs.Name.String()
	}
	out += " = "
	if vs.Value != nil {
		out += vs.Value.String()
	}
	out += ";"
	return out
}

// Identifier representa un identificador en el código.
type Identifier struct {
	Token lexer.Token // El token IDENTIFIER.
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Lexeme }
func (i *Identifier) String() string       { return i.Value }

// ExpressionStatement es una sentencia que consiste en una sola expresión.
type ExpressionStatement struct {
	Token      lexer.Token // El primer token de la expresión.
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Lexeme }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// FuncStatement representa una declaración de función.
type FuncStatement struct {
	Token      lexer.Token // El token 'func'.
	Name       *Identifier
	Parameters []*Identifier // Cambiado de []*Variable a []*Identifier
	ReturnType string      // Nuevo campo para el tipo de retorno
	Body       *BlockStatement
}
func (fs *FuncStatement) statementNode()       {}
func (fs *FuncStatement) TokenLiteral() string { return fs.Token.Lexeme }
func (fs *FuncStatement) String() string {
	params := []string{}
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}
	returnType := ""
	if fs.ReturnType != "" {
		returnType = fmt.Sprintf(": %s", fs.ReturnType)
	}
	return fmt.Sprintf("%s %s(%s)%s %s", fs.TokenLiteral(), fs.Name.String(), formatStrings(params), returnType, fs.Body.String())
}

// ReturnStatement representa una sentencia de retorno.
type ReturnStatement struct {
	Token       lexer.Token // El token 'return'.
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Lexeme }
func (rs *ReturnStatement) String() string {
	var out string
	out += rs.TokenLiteral() + " "
	if rs.ReturnValue != nil {
		out += rs.ReturnValue.String()
	}
	out += ";"
	return out
}

// BlockStatement representa un bloque de código entre llaves.
type BlockStatement struct {
	Token      lexer.Token // El token '{'.
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Lexeme }
func (bs *BlockStatement) String() string {
	var out string
	for _, s := range bs.Statements {
		out += s.String()
	}
	return out
}

// ForInStatement representa una sentencia 'for' con iteración sobre rangos o listas.
type ForInStatement struct {
	Token      lexer.Token // El token 'for'.
	Identifier *Identifier // El identificador de la variable de iteración (e.g., 'x' in 'for x in ...').
	Iterable   Expression  // La expresión que evalúa a la lista o rango sobre el que iterar.
	Body       *BlockStatement // El cuerpo del bucle.
}

func (fs *ForInStatement) statementNode()       {}
func (fs *ForInStatement) TokenLiteral() string { return fs.Token.Lexeme }
func (fs *ForInStatement) String() string {
	out := "for "
	if fs.Identifier != nil {
		out += fs.Identifier.String()
	}
	out += " in "
	if fs.Iterable != nil {
		out += fs.Iterable.String()
	}
	out += " "
	if fs.Body != nil {
		out += fs.Body.String()
	}
	return out
}

// TryStatement representa una sentencia 'try-catch'.
type TryStatement struct {
	Token       lexer.Token // El token 'try'.
	TryBlock    *BlockStatement
	CatchClause *CatchClause // Puede ser nil si solo hay finally.
	FinallyBlock *BlockStatement // Puede ser nil.
}

func (ts *TryStatement) statementNode()       {}
func (ts *TryStatement) TokenLiteral() string { return ts.Token.Lexeme }
func (ts *TryStatement) String() string {
	out := "try "
	if ts.TryBlock != nil {
		out += ts.TryBlock.String()
	}
	if ts.CatchClause != nil {
		out += " " + ts.CatchClause.String()
	}
	if ts.FinallyBlock != nil {
		out += " finally " + ts.FinallyBlock.String()
	}
	return out
}

// CatchClause representa una cláusula 'catch'.
type CatchClause struct {
	Token      lexer.Token // El token 'catch'.
	Parameter  *Identifier // El identificador para la excepción capturada.
	CatchBlock *BlockStatement
}

func (cc *CatchClause) statementNode()       {} // CatchClause es parte de TryStatement, no una sentencia independiente.
func (cc *CatchClause) TokenLiteral() string { return cc.Token.Lexeme }
func (cc *CatchClause) String() string {
	if cc.Parameter == nil || cc.CatchBlock == nil {
		return "catch (invalid) { invalid }"
	}
	return fmt.Sprintf("catch (%s) %s", cc.Parameter.String(), cc.CatchBlock.String())
}

// ThrowStatement representa una sentencia 'throw'.
type ThrowStatement struct {
	Token     lexer.Token // El token 'throw'.
	Exception Expression
}

func (ths *ThrowStatement) statementNode()       {}
func (ths *ThrowStatement) TokenLiteral() string { return ths.Token.Lexeme }
func (ths *ThrowStatement) String() string {
	var out string
	out += ths.TokenLiteral() + " "
	if ths.Exception != nil {
		out += ths.Exception.String()
	}
	out += ";"
	return out
}

// NumberLiteral representa un literal numérico.
type NumberLiteral struct {
	Token lexer.Token
	Value interface{} // int64 or float64
}

func (nl *NumberLiteral) expressionNode()      {}
func (nl *NumberLiteral) TokenLiteral() string { return nl.Token.Lexeme }
func (nl *NumberLiteral) String() string       { return nl.Token.Lexeme }

// StringLiteral representa un literal de cadena.
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Lexeme }
func (sl *StringLiteral) String() string       { return sl.Token.Lexeme }

// BooleanLiteral representa un literal booleano.
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Lexeme }
func (bl *BooleanLiteral) String() string       { return bl.Token.Lexeme }

// NullLiteral representa un literal null.
type NullLiteral struct {
	Token lexer.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Lexeme }
func (nl *NullLiteral) String() string       { return "null" }

// PrefixExpression representa una expresión con un operador prefijo.
type PrefixExpression struct {
	Token    lexer.Token // El operador prefijo.
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Lexeme }
func (pe *PrefixExpression) String() string {
	if pe.Right == nil {
		return fmt.Sprintf("(%sINVALID)", pe.Operator)
	}
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right.String())
}

// InfixExpression representa una expresión con un operador infijo.
type InfixExpression struct {
	Token    lexer.Token // El operador infijo.
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Lexeme }
func (ie *InfixExpression) String() string {
	if ie.Left == nil || ie.Right == nil {
		return fmt.Sprintf("(INVALID %s INVALID)", ie.Operator)
	}
	return fmt.Sprintf("(%s %s %s)", ie.Left.String(), ie.Operator, ie.Right.String())
}

// CallExpression representa una llamada a función.
type CallExpression struct {
	Token     lexer.Token // El token '(' o el identificador de la función.
	Function  Expression  // La expresión que evalúa a la función.
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Lexeme }
func (ce *CallExpression) String() string {
	if ce.Function == nil {
		return "INVALID()"
	}
	return fmt.Sprintf("%s(%s)", ce.Function.String(), formatExpressions(ce.Arguments))
}

// IndexExpression representa el acceso a un índice (ej. array[index]).
type IndexExpression struct {
	Token lexer.Token // El token '['
	Left  Expression  // La expresión que evalúa al objeto indexable.
	Index Expression  // La expresión que evalúa al índice.
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Lexeme }
func (ie *IndexExpression) String() string {
	if ie.Left == nil || ie.Index == nil {
		return "(INVALID[INVALID])"
	}
	return fmt.Sprintf("(%s[%s])", ie.Left.String(), ie.Index.String())
}

// MemberExpression representa el acceso a un miembro (ej. object.property).
type MemberExpression struct {
	Token    lexer.Token // El token del identificador de la propiedad.
	Object   Expression  // La expresión que evalúa al objeto.
	Property *Identifier // El identificador de la propiedad.
}

func (me *MemberExpression) expressionNode()      {}
func (me *MemberExpression) TokenLiteral() string { return me.Token.Lexeme }
func (me *MemberExpression) String() string {
	if me.Object == nil || me.Property == nil {
		return "(INVALID.INVALID)"
	}
	return fmt.Sprintf("(%s.%s)", me.Object.String(), me.Property.String())
}

// BlockExpression representa un bloque de código como una expresión.
type BlockExpression struct {
	Token lexer.Token // El token '{'.
	Block *BlockStatement
}

func (be *BlockExpression) expressionNode()      {}
func (be *BlockExpression) TokenLiteral() string { return be.Token.Lexeme }
func (be *BlockExpression) String() string {
	if be.Block == nil {
		return "{INVALID}"
	}
	return be.Block.String()
}

// IfStatement representa una sentencia 'if'.
type IfStatement struct {
	Token       lexer.Token     // El token 'if'.
	Condition   Expression      // La condición del if.
	Consequence *BlockStatement // El bloque del if.
	Alternative *BlockStatement // El bloque del else/elif (si existe).
	// Note: Elif se maneja como IfStatement dentro del Alternative.
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Lexeme }
func (is *IfStatement) String() string {
	out := "if "
	if is.Condition != nil {
		out += is.Condition.String()
	}
	out += " "
	if is.Consequence != nil {
		out += is.Consequence.String()
	}
	if is.Alternative != nil {
		out += " else " + is.Alternative.String()
	}
	return out
}

// BreakStatement representa una sentencia 'break'.
type BreakStatement struct {
	Token lexer.Token // El token 'break'.
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Lexeme }
func (bs *BreakStatement) String() string       { return bs.Token.Lexeme + ";" }

// ContinueStatement representa una sentencia 'continue'.
type ContinueStatement struct {
	Token lexer.Token // El token 'continue'.
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Lexeme }
func (cs *ContinueStatement) String() string       { return cs.Token.Lexeme + ";" }

// WhileStatement representa una sentencia 'while'.
type WhileStatement struct {
	Token     lexer.Token // El token 'while'.
	Condition Expression  // La condición del bucle.
	Body      *BlockStatement // El cuerpo del bucle.
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Lexeme }
func (ws *WhileStatement) String() string {
	out := "while "
	if ws.Condition != nil {
		out += ws.Condition.String()
	}
	out += " "
	if ws.Body != nil {
		out += ws.Body.String()
	}
	return out
}

// ClassStatement representa una declaración de clase.
type ClassStatement struct {
	Token      lexer.Token // El token 'class'.
	Name       *Identifier
	Attributes []*VarStatement // Atributos de la clase
	Methods    []*FuncStatement // Métodos de la clase
	InitMethod *FuncStatement // Método constructor (init)
}

func (cs *ClassStatement) statementNode()       {}
func (cs *ClassStatement) TokenLiteral() string { return cs.Token.Lexeme }
func (cs *ClassStatement) String() string {
	out := "class "
	if cs.Name != nil {
		out += cs.Name.String()
	}
	out += " {\n"
	for _, attr := range cs.Attributes {
		out += "    " + attr.String() + "\n"
	}
	for _, method := range cs.Methods {
		out += "    " + method.String() + "\n"
	}
	out += "}"
	return out
}

// ListLiteral representa un literal de lista (e.g., [1, 2, 3]).
type ListLiteral struct {
	Token    lexer.Token // El token '['.
	Elements []Expression
}

func (ll *ListLiteral) expressionNode()      {}
func (ll *ListLiteral) TokenLiteral() string { return ll.Token.Lexeme }
func (ll *ListLiteral) String() string {
	if ll.Elements == nil {
		return "[]"
	}
	return fmt.Sprintf("[%s]", formatExpressions(ll.Elements))
}
// HashLiteral representa un literal de hash (e.g., {key: value}).
type HashLiteral struct {
	Token lexer.Token // El token '{'.
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Lexeme }
func (hl *HashLiteral) String() string {
	if hl.Pairs == nil {
		return "{}"
	}
	var pairs []string
	for key, value := range hl.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", key.String(), value.String()))
	}
	return fmt.Sprintf("{%s}", formatStrings(pairs))
}

// ClassInstantiation representa la instanciación de una clase (e.g., Persona("Wilson", 25)).
type ClassInstantiation struct {
	Token     lexer.Token // El token de la clase.
	ClassName *Identifier
	Arguments []Expression
}

func (ci *ClassInstantiation) expressionNode()      {}
func (ci *ClassInstantiation) TokenLiteral() string { return ci.Token.Lexeme }
func (ci *ClassInstantiation) String() string {
	if ci.ClassName == nil {
		return "INVALID()"
	}
	return fmt.Sprintf("%s(%s)", ci.ClassName.String(), formatExpressions(ci.Arguments))
}

// ThisExpression representa la expresión 'this'
type ThisExpression struct {
	Token lexer.Token // El token 'this'.
}

func (te *ThisExpression) expressionNode()      {}
func (te *ThisExpression) TokenLiteral() string { return te.Token.Lexeme }
func (te *ThisExpression) String() string       { return "this" }

// Helper para formatear listas de expresiones en strings.
func formatExpressions(exps []Expression) string {
	var parts []string
	for _, exp := range exps {
		parts = append(parts, exp.String())
	}
	return formatStrings(parts)
}

func formatStrings(strs []string) string {
	var result string
	for i, s := range strs {
		result += s
		if i < len(strs)-1 {
			result += ", "
		}
	}
	return result
}
