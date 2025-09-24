	package parser

	import (
		"context"
		"fmt"
		"time"

		"github.com/zylo-lang/zylo/internal/ast"
		"github.com/zylo-lang/zylo/internal/lexer"
	)

	// Parser toma una secuencia de tokens y construye un AST.
	type Parser struct {
		l      *lexer.Lexer
		errors []string

		curToken  lexer.Token
		peekToken lexer.Token

		// Protecciones contra memory leak
		recursionDepth    int
		maxRecursionDepth int
		maxErrors         int

		// Funciones de parsing para prefijos y sufijos.
		prefixParseFns map[lexer.TokenType]prefixParseFn
		infixParseFns  map[lexer.TokenType]infixParseFn
	}

	// prefixParseFn es el tipo para funciones que parsean expresiones prefijo.
	type prefixParseFn func() ast.Expression

	// infixParseFn es el tipo para funciones que parsean expresiones infijo.
	type infixParseFn func(ast.Expression) ast.Expression

	// New crea un nuevo Parser.
	func New(l *lexer.Lexer) *Parser {
		p := &Parser{
			l:                 l,
			errors:            []string{},
			maxRecursionDepth: 1000,
			maxErrors:         100,
		}
	
		p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
		p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	
		// LITERALES Y IDENTIFICADORES
		p.registerPrefix(lexer.IDENTIFIER, p.parseIdentifier)
		p.registerPrefix(lexer.NUMBER, p.parseNumberLiteral)
		p.registerPrefix(lexer.STRING, p.parseStringLiteral)
		p.registerPrefix(lexer.TRUE, p.parseBooleanLiteral)
		p.registerPrefix(lexer.FALSE, p.parseBooleanLiteral)
		p.registerPrefix(lexer.NIL, p.parseNullLiteral)
	
		// OPERADORES PREFIJO
		p.registerPrefix(lexer.BANG, p.parsePrefixExpression)
		p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
		p.registerPrefix(lexer.PLUS, p.parsePrefixExpression) // +num
	
		// AGRUPACIÓN Y ESTRUCTURAS
		p.registerPrefix(lexer.LEFT_PAREN, p.parseGroupedExpression)
		p.registerPrefix(lexer.LEFT_BRACE, p.parseHashLiteral) // Para objetos {}
		p.registerPrefix(lexer.LEFT_BRACKET, p.parseListLiteral) // Para arrays []
	
		// PALABRAS CLAVE QUE PUEDEN SER EXPRESIONES
		p.registerPrefix(lexer.THIS, p.parseThisExpression)
		p.registerPrefix(lexer.SUPER, p.parseSuperExpression)
		// FUNC is handled as statement, not expression
		p.registerPrefix(lexer.IMPORT, p.parseImportExpression) // Import como expresión
		p.registerPrefix(lexer.ELIF, func() ast.Expression {
			// ELIF no debería ser una expresión, devolver error controlado
			p.addError("elif must be used after if statement")
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})

		// Dummy prefix parsers for tokens that should not appear in expressions
		p.registerPrefix(lexer.RETURN, func() ast.Expression {
			p.addError("return statement not allowed in expression")
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})
		p.registerPrefix(lexer.VAR, func() ast.Expression {
			p.addError("var statement not allowed in expression")
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})
		p.registerPrefix(lexer.IF, func() ast.Expression {
			p.addError("if statement not allowed in expression")
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})
		p.registerPrefix(lexer.FUNC, func() ast.Expression {
			p.addError("func statement not allowed in expression")
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})

		// Registrar NEWLINE como token que se puede saltar
		p.registerPrefix(lexer.NEWLINE, func() ast.Expression {
			// Saltar el NEWLINE y continuar
			p.nextToken()
			return p.parseExpression(LOWEST)
		})

		// Dummy prefix parsers for tokens that should not appear in expressions
		p.registerPrefix(lexer.COMMA, func() ast.Expression {
			p.addError("comma not allowed in expression")
			p.nextToken()
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})
		p.registerPrefix(lexer.DOT, func() ast.Expression {
			p.addError("dot not allowed in expression")
			p.nextToken()
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})
		p.registerPrefix(lexer.EQUAL, func() ast.Expression {
			p.addError("= not allowed in expression")
			p.nextToken()
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})
		p.registerPrefix(lexer.LESS, func() ast.Expression {
			p.addError("< not allowed in expression")
			p.nextToken()
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})
		p.registerPrefix(lexer.BANG_EQUAL, func() ast.Expression {
			p.addError("!= not allowed in expression")
			p.nextToken()
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})
		p.registerPrefix(lexer.AND, func() ast.Expression {
			p.addError("&& not allowed in expression")
			p.nextToken()
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})
		p.registerPrefix(lexer.RIGHT_PAREN, func() ast.Expression {
			p.addError(") not allowed in expression")
			p.nextToken()
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})
		p.registerPrefix(lexer.EOF, func() ast.Expression {
			p.addError("EOF not allowed in expression")
			p.nextToken()
			return &ast.Identifier{Token: p.curToken, Value: "ERROR"}
		})
	
		// OPERADORES INFIJO
		p.registerInfix(lexer.PLUS, p.parseInfixExpression)
		p.registerInfix(lexer.MINUS, p.parseInfixExpression)
		p.registerInfix(lexer.STAR, p.parseInfixExpression)
		p.registerInfix(lexer.SLASH, p.parseInfixExpression)
		p.registerInfix(lexer.PERCENT, p.parseInfixExpression)
	
		// COMPARACIÓN
		p.registerInfix(lexer.EQUAL_EQUAL, p.parseInfixExpression)
		p.registerInfix(lexer.BANG_EQUAL, p.parseInfixExpression)
		p.registerInfix(lexer.LESS, p.parseInfixExpression)
		p.registerInfix(lexer.LESS_EQUAL, p.parseInfixExpression)
		p.registerInfix(lexer.GREATER, p.parseInfixExpression)
		p.registerInfix(lexer.GREATER_EQUAL, p.parseInfixExpression)
	
		// LÓGICOS
		p.registerInfix(lexer.AND, p.parseInfixExpression)
		p.registerInfix(lexer.OR, p.parseInfixExpression)
	
		// ASIGNACIÓN
		p.registerInfix(lexer.EQUAL, p.parseInfixExpression)
	
		// ACCESO
		p.registerInfix(lexer.LEFT_PAREN, func(left ast.Expression) ast.Expression {
			return p.parseCallExpression(left)
		})
		p.registerInfix(lexer.LEFT_BRACKET, func(left ast.Expression) ast.Expression {
			return p.parseIndexExpression(left)
		})
		p.registerInfix(lexer.DOT, p.parseDotExpression)
	
		p.nextToken()
		p.nextToken()
	
		return p
	}

	// registerPrefix registra una función de parsing prefijo.
	func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
		p.prefixParseFns[tokenType] = fn
	}

	// registerInfix registra una función de parsing infijo.
	func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
		p.infixParseFns[tokenType] = fn
	}

	// ParseProgram es el punto de entrada para el parsing.
	func (p *Parser) ParseProgram() *ast.Program {
		return p.ParseProgramWithTimeout(30 * time.Second)
	}

	// ParseProgramWithTimeout - con debugging y protección extra
	func (p *Parser) ParseProgramWithTimeout(timeout time.Duration) *ast.Program {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		program := &ast.Program{}
		program.Statements = []ast.Statement{}

		iterations := 0 // Contador de seguridad

		// Saltar tokens NEWLINE iniciales
		for p.curToken.Type == lexer.NEWLINE && p.curToken.Type != lexer.EOF {
			p.nextToken()
		}

		for p.curToken.Type != lexer.EOF {
			iterations++
			if iterations > 1000 { // Límite de seguridad
				p.addError("too many iterations - possible infinite loop")
				break
			}

			select {
			case <-ctx.Done():
				p.addError("parsing timeout")
				return program
			default:
			}

			// Guardar token actual para detectar si avanzamos
			currentTokenType := p.curToken.Type

			stmt := p.parseStatement()
			if stmt != nil {
				program.Statements = append(program.Statements, stmt)
			}

			// CRÍTICO: Siempre avanzar si no cambió el token
			if p.curToken.Type == currentTokenType {
				p.nextToken()
			}

			if len(program.Statements) > 1000 {
				p.addError("too many statements")
				break
			}
		}

		return program
	}

	// Errors devuelve la lista de errores encontrados durante el parsing.
	func (p *Parser) Errors() []string {
		return p.errors
	}

	// Funciones helper para control de recursión
	func (p *Parser) enterRecursion() error {
		p.recursionDepth++
		if p.recursionDepth > p.maxRecursionDepth {
			return fmt.Errorf("máxima profundidad de recursión alcanzada (%d)", p.maxRecursionDepth)
		}
		return nil
	}

	func (p *Parser) exitRecursion() {
		p.recursionDepth--
	}

	// addError añade un error con protección contra overflow
	func (p *Parser) addError(msg string) {
		if len(p.errors) < p.maxErrors {
			p.errors = append(p.errors, msg)
		}
	}

	// skipNewlines salta tokens NEWLINE consecutivos
	func (p *Parser) skipNewlines() {
		for p.curToken.Type == lexer.NEWLINE {
			p.nextToken()
		}
	}

	// noPrefixParseFnError añade un error cuando no hay función de parsing prefijo
	func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
		msg := fmt.Sprintf("no prefix parse function for %s found", t)
		p.addError(msg)
	}

	// parseImportStatement analiza una declaración de import.
	func (p *Parser) parseImportStatement() *ast.ImportStatement {
		stmt := &ast.ImportStatement{Token: p.curToken}

		if !p.expectPeek(lexer.IDENTIFIER) {
			return nil
		}

		stmt.ModuleName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

		// Skip newlines after import statement
		for p.peekTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		return stmt
	}

	// parseImportExpression analiza import como expresión (para casos donde aparece en contexto de expresión)
	func (p *Parser) parseImportExpression() ast.Expression {
		stmt := &ast.ImportStatement{Token: p.curToken}

		if !p.expectPeek(lexer.IDENTIFIER) {
			return nil
		}

		stmt.ModuleName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

		return stmt
	}

	// parseBreakStatement analiza una sentencia 'break'.
	func (p *Parser) parseBreakStatement() *ast.BreakStatement {
		stmt := &ast.BreakStatement{Token: p.curToken}
		return stmt
	}

	// parseContinueStatement analiza una sentencia 'continue'.
	func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
		stmt := &ast.ContinueStatement{Token: p.curToken}
		return stmt
	}

	// parseElifStatement analiza una sentencia 'elif' - esto se trata como un if anidado
	func (p *Parser) parseElifStatement() ast.Statement {
		// 'elif' se convierte en 'if' para el parser
		return p.parseIfStatement()
	}

	// AGREGAR parseStatement mejorado que maneje ELIF correctamente
	func (p *Parser) parseStatement() ast.Statement {
		// Saltar NEWLINES al inicio
		for p.curToken.Type == lexer.NEWLINE {
			p.nextToken()
		}

		switch p.curToken.Type {
		case lexer.IMPORT:
			return p.parseImportStatement()
		case lexer.VAR:
			return p.parseVarStatement()
		case lexer.FUNC:
			return p.parseFuncStatement()
		case lexer.RETURN:
			return p.parseReturnStatement()
		case lexer.IF:
			return p.parseIfStatement()
		case lexer.ELIF:
			// ELIF solo es válido después de IF, tratar como error
			p.addError("elif without preceding if statement")
			return nil
		case lexer.WHILE:
			return p.parseWhileStatement()
		case lexer.FOR:
			return p.parseForStatement()
		case lexer.TRY:
			return p.parseTryStatement()
		case lexer.CLASS:
			return p.parseClassStatement()
		case lexer.BREAK:
			return p.parseBreakStatement()
		case lexer.CONTINUE:
			return p.parseContinueStatement()
		case lexer.THROW:
			return p.parseThrowStatement()
		case lexer.SEMICOLON, lexer.NEWLINE:
			return nil
		case lexer.RIGHT_BRACE:
			return nil
		default:
			return p.parseExpressionStatement()
		}
	}

	func (p *Parser) parseVarStatement() *ast.VarStatement {
		stmt := &ast.VarStatement{Token: p.curToken}

		p.skipNewlines()
		if !p.expectPeek(lexer.IDENTIFIER) {
			return nil
		}

		stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
	
		p.skipNewlines()
	
		// CRÍTICO: Manejar tipo opcional ": Float" o ": Array<String>"
		if p.peekTokenIs(lexer.COLON) {
			p.nextToken() // consumir ':'
			if !p.expectPeek(lexer.IDENTIFIER) {
				return nil
			}
			// Opcional: guardar tipo de variable
			// stmt.Type = p.curToken.Lexeme

			// Manejar tipos genéricos como Array<String>
			if p.peekTokenIs(lexer.LESS) {
				p.nextToken() // consume <
				if p.expectPeek(lexer.IDENTIFIER) {
					// Tipo parámetro, ignorar por ahora
				}
				p.expectPeek(lexer.GREATER) // consume >
			}
		}

		if p.peekTokenIs(lexer.EQUAL) {
			p.nextToken() // consume =
			p.nextToken() // move to value
			stmt.Value = p.parseExpression(LOWEST)
		}

		if p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}

		return stmt
	}

	func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
		stmt := &ast.ReturnStatement{Token: p.curToken}

		p.nextToken()
		if !p.curTokenIs(lexer.SEMICOLON) && !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.RIGHT_BRACE) {
			stmt.ReturnValue = p.parseExpression(LOWEST)
		}

		if p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}

		return stmt
	}

	func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
		// Verificar que el token actual puede ser una expresión
		if p.curToken.Type == lexer.RIGHT_PAREN ||
		   p.curToken.Type == lexer.ELSE ||
		   p.curToken.Type == lexer.CATCH {
			// Estos no son expresiones válidas, devolver nil
			return nil
		}

		stmt := &ast.ExpressionStatement{Token: p.curToken}
		stmt.Expression = p.parseExpression(LOWEST)

		// CRÍTICO: Si la expresión es nil, devolver nil en lugar del statement
		if stmt.Expression == nil {
			return nil
		}

		// Consumir ; si está presente
		if p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}

		return stmt
	}

// ARREGLAR parseFuncStatement para el error "expected LEFT_BRACE, got NUMBER"
func (p *Parser) parseFuncStatement() *ast.FuncStatement {
	stmt := &ast.FuncStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENTIFIER) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

	if !p.expectPeek(lexer.LEFT_PAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	// CRÍTICO: Manejar tipo de retorno opcional sin consumir tokens incorrectos
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // consume :
		if p.expectPeek(lexer.IDENTIFIER) {
			// stmt.ReturnType = p.curToken.Lexeme
		} else {
			return nil
		}
	} else if p.peekTokenIs(lexer.IDENTIFIER) {
		// Tipo sin dos puntos
		p.nextToken()
		// stmt.ReturnType = p.curToken.Lexeme
	}

	// Skip newlines antes de {
	p.skipNewlines()

	if !p.curTokenIs(lexer.LEFT_BRACE) {
		p.addError("expected '{' to start function body")
		return nil
	}

	p.nextToken() // consume {
	stmt.Body = p.parseBlockStatement()
	return stmt
}

	func (p *Parser) parseFunctionParameters() []*ast.Identifier {
		identifiers := []*ast.Identifier{}

		if p.peekTokenIs(lexer.RIGHT_PAREN) {
			p.nextToken() // consume )
			p.nextToken() // advance past )
			return identifiers
		}

		p.nextToken() // consume primer identificador

		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

		// Revisar si hay : Tipo
		if p.peekTokenIs(lexer.COLON) {
			p.nextToken() // :
			p.nextToken() // tipo
			// Por ahora ignorar el tipo
		}

		identifiers = append(identifiers, ident)

		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // ,
			p.nextToken() // siguiente ident

			ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

			if p.peekTokenIs(lexer.COLON) {
				p.nextToken() // :
				p.nextToken() // tipo
				// Ignorar tipo
			}

			identifiers = append(identifiers, ident)
		}

		p.skipNewlines()
		if !p.expectPeek(lexer.RIGHT_PAREN) {
			return nil
		}
		p.nextToken() // consume )

		return identifiers
	}



	// Helper functions
	func (p *Parser) nextToken() {
		p.curToken = p.peekToken
		p.peekToken = p.l.NextToken()
	}

	func (p *Parser) curTokenIs(t lexer.TokenType) bool {
		return p.curToken.Type == t
	}

	func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
		return p.peekToken.Type == t
	}

	func (p *Parser) expectPeek(t lexer.TokenType) bool {
		if p.peekTokenIs(t) {
			p.nextToken()
			return true
		}
		p.peekError(t)
		return false
	}

	func (p *Parser) peekError(t lexer.TokenType) {
		msg := fmt.Sprintf("expected next token to be %s, got %s instead",
			t, p.peekToken.Type)
		p.addError(msg)
	}

	// Parsing functions (minimal implementations)
	func (p *Parser) parseIdentifier() ast.Expression {
		expr := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
		p.nextToken() // consume the IDENTIFIER token
		return expr
	}

	func (p *Parser) parseNumberLiteral() ast.Expression {
		lit := &ast.NumberLiteral{Token: p.curToken}
		if p.curToken.Literal != nil {
			lit.Value = p.curToken.Literal
		} else {
			lit.Value = int64(0)
		}
		p.nextToken() // consume the NUMBER token
		return lit
	}

	func (p *Parser) parseStringLiteral() ast.Expression {
		lit := &ast.StringLiteral{Token: p.curToken}
		if p.curToken.Literal != nil {
			if val, ok := p.curToken.Literal.(string); ok {
				lit.Value = val
			} else {
				lit.Value = ""
			}
		} else {
			lit.Value = ""
		}
		p.nextToken() // consume the STRING token
		return lit
	}

	func (p *Parser) parseBooleanLiteral() ast.Expression {
		expr := &ast.BooleanLiteral{
			Token: p.curToken,
			Value: p.curToken.Type == lexer.TRUE,
		}
		p.nextToken() // consume the TRUE/FALSE token
		return expr
	}

	func (p *Parser) parseNullLiteral() ast.Expression {
		expr := &ast.NullLiteral{Token: p.curToken}
		p.nextToken() // consume the NIL token
		return expr
	}

	func (p *Parser) parsePrefixExpression() ast.Expression {
		token := p.curToken
		operator := p.curToken.Lexeme
		p.nextToken() // consume the operator
		right := p.parseExpression(PREFIX)
		return &ast.PrefixExpression{
			Token:    token,
			Operator: operator,
			Right:    right,
		}
	}

	func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
		exp := &ast.InfixExpression{
			Token:    p.curToken,
			Left:     left,
			Operator: p.curToken.Lexeme,
		}

		precedence := p.currentPrecedence()
		p.nextToken() // consume operator
		exp.Right = p.parseExpression(precedence)

		return exp
	}

	func (p *Parser) parseGroupedExpression() ast.Expression {
		p.nextToken()
		exp := p.parseExpression(LOWEST)
		p.skipNewlines()
		p.expectPeek(lexer.RIGHT_PAREN)
		p.nextToken() // consume )
		return exp
	}

	func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
		exp := &ast.CallExpression{
			Token:     p.curToken,
			Function:  left,
			Arguments: []ast.Expression{},
		}

		// Parse arguments
		if !p.peekTokenIs(lexer.RIGHT_PAREN) {
			p.nextToken() // move to first argument
			exp.Arguments = append(exp.Arguments, p.parseExpression(LOWEST))

			for p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // consume ,
				p.nextToken() // move to next argument
				exp.Arguments = append(exp.Arguments, p.parseExpression(LOWEST))
			}
		}

		p.skipNewlines()
		if !p.expectPeek(lexer.RIGHT_PAREN) {
			return nil
		}
		p.nextToken() // consume )

		return exp
	}

	func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
		exp := &ast.IndexExpression{
			Token: p.curToken,
			Left:  left,
		}
		p.nextToken() // consume [
		exp.Index = p.parseExpression(LOWEST)
		p.skipNewlines()
		if !p.expectPeek(lexer.RIGHT_BRACKET) {
			return nil
		}
		p.nextToken() // consume ]
		return exp
	}

	func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
		if !p.expectPeek(lexer.IDENTIFIER) {
			return nil
		}

		prop := &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}

		memberExp := &ast.MemberExpression{
			Token:    p.curToken,
			Object:   left,
			Property: prop,
		}

		// Si hay paréntesis, es una llamada a método como show.log()
		if p.peekTokenIs(lexer.LEFT_PAREN) {
			p.nextToken() // consumir '('
			return p.parseCallExpression(memberExp)
		}

		return memberExp
	}

	func (p *Parser) parseExpression(precedence int) ast.Expression {
		// MEJORAR: Saltar múltiples NEWLINES
		for p.curToken.Type == lexer.NEWLINE {
			p.nextToken()
		}

		prefix := p.prefixParseFns[p.curToken.Type]
		if prefix == nil {
			// No tratar NEWLINE como error
			if p.curToken.Type == lexer.NEWLINE {
				return nil
			}
			if p.curToken.Type == lexer.RIGHT_BRACKET {
				return nil
			}
			p.noPrefixParseFnError(p.curToken.Type)
			return nil
		}

		leftExp := prefix()
		if leftExp == nil {
			return nil
		}

		iterations := 0
		for precedence < p.peekPrecedence() && iterations < 50 {
			iterations++

			infix := p.infixParseFns[p.peekToken.Type]
			if infix == nil {
				return leftExp
			}

			p.nextToken()
			newLeftExp := infix(leftExp)

			if newLeftExp == nil {
				return leftExp
			}

			leftExp = newLeftExp
		}

		return leftExp
	}

	func (p *Parser) currentPrecedence() int {
		if p, ok := precedences[p.curToken.Type]; ok {
			return p
		}
		return LOWEST
	}

	func (p *Parser) peekPrecedence() int {
		if p, ok := precedences[p.peekToken.Type]; ok {
			return p
		}
		return LOWEST
	}

	func (p *Parser) parseBlockExpression() ast.Expression {
		return &ast.BlockExpression{Token: p.curToken}
	}

	func (p *Parser) parseBlockStatement() *ast.BlockStatement {
		block := &ast.BlockStatement{Token: p.curToken, Statements: []ast.Statement{}}

		if !p.curTokenIs(lexer.LEFT_BRACE) {
			p.addError("expected '{' to start block")
			return nil
		}
		p.nextToken() // consume {

		iterations := 0
		for !p.curTokenIs(lexer.RIGHT_BRACE) && !p.curTokenIs(lexer.EOF) {
			iterations++
			if iterations > 100 {
				p.addError("too many iterations in parseBlockStatement")
				break
			}

			// Skip newlines
			if p.curTokenIs(lexer.NEWLINE) {
				p.nextToken()
				continue
			}

			// CRÍTICO: Detener si encontramos tokens que no deberían estar aquí
			if p.curToken.Type == lexer.RIGHT_PAREN ||
			   p.curToken.Type == lexer.ELSE ||
			   p.curToken.Type == lexer.CATCH {
				// Estos tokens indican fin de bloque o tokens mal posicionados
				break
			}

			currentToken := p.curToken.Type
			stmt := p.parseStatement()

			if stmt != nil {
				block.Statements = append(block.Statements, stmt)
			}

			// Protección contra bucles infinitos
			if p.curToken.Type == currentToken {
				p.nextToken()
			}
		}

		if p.curTokenIs(lexer.RIGHT_BRACE) {
			p.nextToken() // consume }
		}

		return block
	}

	// ARREGLAR parseExpressionList completamente para COMMA y RIGHT_BRACKET
		func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
			args := []ast.Expression{}
	
			// Lista vacía - si el siguiente token es el de cierre
			if p.peekTokenIs(end) {
				p.nextToken() // consume closing token
				return args
			}
	
			// Avanzar al primer elemento
			p.nextToken()
	
			// Parsear primer elemento si no es el token de cierre
			if !p.curTokenIs(end) {
				expr := p.parseExpression(LOWEST)
				if expr != nil {
					args = append(args, expr)
				}
	
				// Parsear elementos adicionales separados por comas
				for p.peekTokenIs(lexer.COMMA) {
					p.nextToken() // consume comma
					p.nextToken() // move to next expression
	
					if p.curTokenIs(end) {
						// Coma trailing, terminar
						break
					}
	
					expr := p.parseExpression(LOWEST)
					if expr != nil {
						args = append(args, expr)
					}
				}
			}
	
			if !p.expectPeek(end) {
				return nil
			}
			p.nextToken() // consume end
	
			return args
		}

// MEJORAR parseListLiteral para manejar arrays correctamente
func (p *Parser) parseListLiteral() ast.Expression {
	lit := &ast.ListLiteral{Token: p.curToken, Elements: []ast.Expression{}}

	// Array vacío []
	if p.peekTokenIs(lexer.RIGHT_BRACKET) {
		p.nextToken() // consume ]
		return lit
	}

	// Usar parseExpressionList mejorado
	lit.Elements = p.parseExpressionList(lexer.RIGHT_BRACKET)
	p.nextToken() // consume ]
	return lit
}

// parseIfStatement con manejo de paréntesis mejorado:
func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	p.nextToken() // consume IF

	// Detectar si hay paréntesis
	hasParens := p.curTokenIs(lexer.LEFT_PAREN)
	if hasParens {
		p.nextToken() // consume (
	}

	stmt.Condition = p.parseExpression(LOWEST)

	if hasParens {
		p.skipNewlines()
		if !p.expectPeek(lexer.RIGHT_PAREN) {
			return nil
		}
	}

	// Skip newlines
	p.skipNewlines()

	if !p.curTokenIs(lexer.LEFT_BRACE) {
		p.addError("expected '{' after if condition")
		return nil
	}

	p.nextToken() // consume {
	stmt.Consequence = p.parseBlockStatement()

	// Manejar else o elif
	if p.peekTokenIs(lexer.ELSE) || p.peekTokenIs(lexer.ELIF) {
		p.nextToken() // consume current token (likely })
		p.nextToken() // consume ELSE/ELIF
		if p.curTokenIs(lexer.IF) {
			p.nextToken() // consume IF si es else if
		}
		if p.peekTokenIs(lexer.LEFT_BRACE) {
			// else normal
			if !p.expectPeek(lexer.LEFT_BRACE) {
				return nil
			}
			stmt.Alternative = p.parseBlockStatement()
		} else {
			// else if o elif - crear nuevo if statement
			// Crear un nuevo if statement para elif
			elifStmt := &ast.IfStatement{Token: p.curToken}
			elifStmt.Condition = p.parseExpression(LOWEST)
			// Skip newlines
			p.skipNewlines()
			if !p.curTokenIs(lexer.LEFT_BRACE) {
				p.addError("expected '{' after elif condition")
				return nil
			}
			p.nextToken() // consume {
			elifStmt.Consequence = p.parseBlockStatement()
			stmt.Alternative = &ast.BlockStatement{
				Token: p.curToken,
				Statements: []ast.Statement{elifStmt},
			}
		}
	}

	return stmt
}

	// parseForStatement analiza una sentencia 'for in'
	func (p *Parser) parseForStatement() ast.Statement {
		token := p.curToken // FOR token
	
		// Determinar tipo de for loop
		if p.peekTokenIs(lexer.IDENTIFIER) {
			// Podría ser for-in: for x in array
			p.nextToken() // consume FOR
			identifier := p.curToken
	
			if p.peekTokenIs(lexer.IN) {
				return p.parseForInStatement(token, identifier)
			} else {
				// for tradicional: for i = 0; i < 10; i++
				return p.parseTraditionalForStatement(token)
			}
		}
	
		// for con paréntesis: for (i = 0; i < 10; i++)
		return p.parseTraditionalForStatement(token)
	}
	
	func (p *Parser) parseForInStatement(forToken lexer.Token, identifier lexer.Token) *ast.ForInStatement {
		stmt := &ast.ForInStatement{Token: forToken}
		stmt.Identifier = &ast.Identifier{Token: identifier, Value: identifier.Lexeme}
	
		if !p.expectPeek(lexer.IN) {
			return nil
		}
	
		p.nextToken()
		stmt.Iterable = p.parseExpression(LOWEST)
	
		// Skip newlines
		p.skipNewlines()

		if !p.curTokenIs(lexer.LEFT_BRACE) {
			p.addError("expected '{' after for in")
			return nil
		}

		p.nextToken() // consume {
		stmt.Body = p.parseBlockStatement()
		return stmt
	}
	
	func (p *Parser) parseTraditionalForStatement(forToken lexer.Token) ast.Statement {
		// Por ahora, devolver un statement vacío
		// TODO: Implementar for tradicional
		return &ast.ExpressionStatement{
			Token: forToken,
			Expression: &ast.Identifier{Token: forToken, Value: "FOR_LOOP"},
		}
	}

	// parseClassStatement analiza una declaración de clase
	func (p *Parser) parseClassStatement() *ast.ClassStatement {
		stmt := &ast.ClassStatement{Token: p.curToken}
	
		if !p.expectPeek(lexer.IDENTIFIER) {
			return nil
		}
		stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
	
		// CRÍTICO: Avanzar después del nombre de la clase
		p.nextToken()
	
		// Skip newlines antes de {
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}
	
		// ARREGLO: Verificar que estamos en LEFT_BRACE antes de continuar
		if !p.curTokenIs(lexer.LEFT_BRACE) {
			p.addError(fmt.Sprintf("expected '{' after class name, got %s", p.curToken.Type))
			return nil
		}

		// Parse class body
		p.nextToken() // consume {
	
		for p.curToken.Type != lexer.RIGHT_BRACE && p.curToken.Type != lexer.EOF {
			if p.curTokenIs(lexer.NEWLINE) {
				p.nextToken()
				continue
			}
	
			switch p.curToken.Type {
			case lexer.VAR:
				attr := p.parseVarStatement()
				if attr != nil {
					if stmt.Attributes == nil {
						stmt.Attributes = []*ast.VarStatement{}
					}
					stmt.Attributes = append(stmt.Attributes, attr)
				}
			case lexer.FUNC:
				method := p.parseFunctionStatement()
				if method != nil {
					if stmt.Methods == nil {
						stmt.Methods = []*ast.FuncStatement{}
					}
					stmt.Methods = append(stmt.Methods, method)
					if method.Name != nil && method.Name.Value == "init" {
						stmt.InitMethod = method
					}
				}
			default:
				// Skip tokens desconocidos sin error
				p.nextToken()
			}
		}
	
		if p.curTokenIs(lexer.RIGHT_BRACE) {
			p.nextToken() // consume }
		}
	
		return stmt
	}

	func (p *Parser) parseTryStatement() *ast.TryStatement {
		stmt := &ast.TryStatement{Token: p.curToken}

		// Skip newlines after try
		p.skipNewlines()
	
		if !p.curTokenIs(lexer.LEFT_BRACE) {
			p.addError("expected '{' after try")
			return nil
		}
	
		p.nextToken() // consume {
		stmt.TryBlock = p.parseBlockStatement()

		// Skip newlines before catch
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		// CRÍTICO: Verificar que realmente es CATCH
		if p.curTokenIs(lexer.CATCH) {
			stmt.CatchClause = &ast.CatchClause{Token: p.curToken}

			// Manejar parámetro opcional
			if p.peekTokenIs(lexer.LEFT_PAREN) {
				p.nextToken() // consume CATCH
				p.nextToken() // consume (
				if p.curTokenIs(lexer.IDENTIFIER) {
					stmt.CatchClause.Parameter = &ast.Identifier{
						Token: p.curToken,
						Value: p.curToken.Lexeme,
					}
				}
				if !p.expectPeek(lexer.RIGHT_PAREN) {
					return nil
				}
			} else if p.peekTokenIs(lexer.IDENTIFIER) {
				p.nextToken() // consume CATCH
				stmt.CatchClause.Parameter = &ast.Identifier{
					Token: p.curToken,
					Value: p.curToken.Lexeme,
				}
			} else {
				p.nextToken() // solo consume CATCH
			}

			// Skip newlines before {
			p.skipNewlines()

			if !p.curTokenIs(lexer.LEFT_BRACE) {
				p.addError("expected '{' after catch")
				return nil
			}

			p.nextToken() // consume {

			stmt.CatchClause.CatchBlock = p.parseBlockStatement()
		}

		return stmt
	}

	// parseThrowStatement analiza una sentencia 'throw'
	func (p *Parser) parseThrowStatement() *ast.ThrowStatement {
		stmt := &ast.ThrowStatement{Token: p.curToken}

		p.nextToken() // consume THROW
		stmt.Exception = p.parseExpression(LOWEST)

		return stmt
	}

	// parseThisExpression analiza una expresión 'this'
	func (p *Parser) parseThisExpression() ast.Expression {
		return &ast.ThisExpression{Token: p.curToken}
	}
	
	func (p *Parser) parseSuperExpression() ast.Expression {
		return &ast.Identifier{Token: p.curToken, Value: "super"}
	}
	
	func (p *Parser) parseFunctionLiteral() ast.Expression {
		// Para funciones anónimas: func(a, b) { return a + b }
		return &ast.Identifier{Token: p.curToken, Value: "FUNCTION_LITERAL"}
	}
	
	func (p *Parser) parseHashLiteral() ast.Expression {
		hash := &ast.HashLiteral{Token: p.curToken, Pairs: make(map[ast.Expression]ast.Expression)}

		// Hash vacío {}
		if p.peekTokenIs(lexer.RIGHT_BRACE) {
			p.nextToken() // consume }
			return hash
		}

		p.nextToken() // consume {

		for !p.curTokenIs(lexer.RIGHT_BRACE) {
			key := p.parseExpression(LOWEST)
			if key == nil {
				return nil
			}

			if !p.expectPeek(lexer.COLON) {
				return nil
			}

			p.nextToken() // consume :
			value := p.parseExpression(LOWEST)
			if value == nil {
				return nil
			}

			hash.Pairs[key] = value

			if p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // consume ,
			} else if !p.peekTokenIs(lexer.RIGHT_BRACE) {
				p.addError("expected ',' or '}' in hash literal")
				return nil
			}
		}

		if !p.expectPeek(lexer.RIGHT_BRACE) {
			return nil
		}
		p.nextToken() // consume }

		return hash
	}


	// Precedence constants
	const (
		LOWEST int = iota
		ASSIGN
		EQUALS
		COMPARES
		SUM
		PRODUCT
		PREFIX
		CALL
		INDEX
	)

// AGREGAR manejo de precedencias para nuevos operadores:
var precedences = map[lexer.TokenType]int{
	lexer.EQUAL:         ASSIGN,
	lexer.OR:            EQUALS,
	lexer.AND:           COMPARES,
	lexer.EQUAL_EQUAL:   EQUALS,
	lexer.BANG_EQUAL:    EQUALS,
	lexer.LESS:          COMPARES,
	lexer.LESS_EQUAL:    COMPARES,
	lexer.GREATER:       COMPARES,
	lexer.GREATER_EQUAL: COMPARES,
	lexer.PLUS:          SUM,
	lexer.MINUS:         SUM,
	lexer.STAR:          PRODUCT,
	lexer.SLASH:         PRODUCT,
	lexer.PERCENT:       PRODUCT,
	lexer.LEFT_PAREN:    CALL,
	lexer.LEFT_BRACKET:  INDEX,
	lexer.DOT:           CALL,
}

	// parseFunctionStatement analiza una declaración de función
	func (p *Parser) parseFunctionStatement() *ast.FuncStatement {
		stmt := &ast.FuncStatement{Token: p.curToken}

		p.nextToken() // consume FUNC

		// Parse function name
		if !p.curTokenIs(lexer.IDENTIFIER) {
			p.addError("expected function name after 'func'")
			return nil
		}
		stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Lexeme}
		p.nextToken() // consume function name

		// Parse parameters
		if !p.curTokenIs(lexer.LEFT_PAREN) {
			p.addError("expected '(' after function name")
			return nil
		}
		p.nextToken() // consume (
		stmt.Parameters = []*ast.Identifier{}

		for !p.curTokenIs(lexer.RIGHT_PAREN) && !p.curTokenIs(lexer.EOF) {
			if p.curTokenIs(lexer.IDENTIFIER) {
				param := &ast.Identifier{
					Token: p.curToken,
					Value: p.curToken.Lexeme,
				}
				stmt.Parameters = append(stmt.Parameters, param)
				p.nextToken()

				// Skip comma if present
				if p.curTokenIs(lexer.COMMA) {
					p.nextToken()
				}
			} else if p.curTokenIs(lexer.RIGHT_PAREN) {
				// Empty parameter list, just break
				break
			} else {
				p.addError("expected parameter name")
				break
			}
		}

		if !p.curTokenIs(lexer.RIGHT_PAREN) {
			p.addError("expected ')' after function parameters")
			return nil
		}
		p.nextToken() // consume )

		// Tipo de retorno opcional
		if p.peekTokenIs(lexer.COLON) {
			p.nextToken() // consume :
			if !p.curTokenIs(lexer.IDENTIFIER) {
				p.addError("expected return type identifier, got " + string(p.curToken.Type))
				return nil
			} else {
				stmt.ReturnType = p.curToken.Lexeme
				p.nextToken()
			}
		}

		// Ignorar NEWLINE antes de '{'
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		// Cuerpo de la función
		if !p.curTokenIs(lexer.LEFT_BRACE) {
			p.addError("expected '{' to start function body")
			return nil
		}

		stmt.Body = p.parseBlockStatement()
		return stmt
	}

	func (p *Parser) parseWhileStatement() *ast.WhileStatement {
		stmt := &ast.WhileStatement{Token: p.curToken}

		p.nextToken() // consume WHILE

		// Manejar paréntesis opcionales explícitamente
		hasParens := false
		if p.curTokenIs(lexer.LEFT_PAREN) {
			hasParens = true
			p.nextToken() // consume '('
		}

		stmt.Condition = p.parseExpression(LOWEST)
	
		// Si había paréntesis, consumir el ')'
		if hasParens {
			p.skipNewlines()
			if !p.expectPeek(lexer.RIGHT_PAREN) {
				return nil
			}
		}

		// Skip newlines before {
		p.skipNewlines()
	
		if !p.curTokenIs(lexer.LEFT_BRACE) {
			p.addError("expected '{' after while condition")
			return nil
		}
	
		p.nextToken() // consume {
		stmt.Body = p.parseBlockStatement()

		return stmt
	}

