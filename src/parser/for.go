package parser

import (
	"github.com/sevenreup/chewa/src/ast"
	"github.com/sevenreup/chewa/src/token"
)

func (parser *Parser) parseForExpression() ast.Expression {
	expression := &ast.ForExpression{Token: parser.curToken}
	if !parser.expectPeek(token.OPENING_PAREN) {
		return nil
	}

	parser.nextToken()

	// parse assignment statement
	if token.LookupVariableType(parser.curToken.Type) == "" {
		return nil
	}

	assigmnent := parser.parseAssignmentStatement()

	expression.Initializer = assigmnent
	expression.Identifier = assigmnent.Identifier

	if expression.Initializer == nil {
		return nil
	}

	parser.nextToken()

	// parse condition statement

	expression.Condition = parser.parseExpression(LOWEST)
	if expression.Condition == nil {
		return nil
	}

	parser.nextToken()
	parser.nextToken()

	// parse increment statement
	expression.Increment = parser.forIncrement()

	if expression.Increment == nil {
		return nil
	}

	if !parser.expectPeek(token.CLOSING_PAREN) {
		return nil
	}

	if !parser.expectPeek(token.OPENING_BRACE) {
		return nil
	}

	expression.Block = parser.parseBlockStatement()

	return expression
}

func (parser *Parser) forIncrement() ast.Expression {
	if parser.curTokenIs(token.CLOSING_PAREN) {
		return nil
	}

	if parser.curTokenIs(token.SEMICOLON) {
		parser.nextToken()
		return nil
	}

	if parser.curTokenIs(token.IDENT) && parser.peekTokenIs(token.ASSIGN) {
		return parser.parseAssignmentStatement()
	}

	if parser.curTokenIs(token.IDENT) && (parser.peekTokenIs(token.PLUS_EQUAL) ||
		parser.peekTokenIs(token.MINUS_EQUAL) ||
		parser.peekTokenIs(token.SLASH_EQUAL) ||
		parser.peekTokenIs(token.STAR_EQUAL)) {
		identifier := parser.parseIdentifier()

		parser.nextToken()

		return parser.parseCompoundExpression(identifier)
	}

	if parser.curTokenIs(token.IDENT) && (parser.peekTokenIs(token.PLUS_PLUS) || parser.peekTokenIs(token.MINUS_MINUS)) {
		parser.nextToken()

		return parser.parsePostfixExpression()
	}

	return nil
}
