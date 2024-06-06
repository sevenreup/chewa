package parser

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/sevenreup/chewa/src/ast"
	"github.com/sevenreup/chewa/src/lexer"
	"github.com/sevenreup/chewa/src/token"
)

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

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if !integ.Value.Equal(decimal.NewFromInt(value)) {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}
	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value,
			integ.TokenLiteral())
		return false
	}
	return true
}

func testStringLiteral(t *testing.T, il ast.Expression, value string) bool {
	stringValue, ok := il.(*ast.StringLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if stringValue.Value != value {
		t.Errorf("stringValue.Value not %s. got=%s", value, stringValue.Value)
		return false
	}
	if stringValue.TokenLiteral() != value {
		t.Errorf("stringValue.TokenLiteral not %s. got=%s", value,
			stringValue.TokenLiteral())
		return false
	}
	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	switch ident := exp.(type) {
	case *ast.Identifier:
		{
			if ident.Value != value {
				t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
				return false
			}
			if ident.TokenLiteral() != value {
				t.Errorf("ident.TokenLiteral not %s. got=%s", value,
					ident.TokenLiteral())
				return false
			}
			return true
		}
	case *ast.IndexExpression:
		{
			return testIdentifier(t, ident.Left, value)
		}
	case *ast.CallExpression:
		{
			return testIdentifier(t, ident.Function, value)
		}
	}
	t.Errorf("exp not *ast.Identifier. got=%T", exp)
	return false

}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
	isIdentifier bool,
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		if isIdentifier {
			return testIdentifier(t, exp, v)
		}
		return testStringLiteral(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{},
	operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.OperatorExpression. got=%T(%s)", exp, exp)
		return false
	}
	if !testLiteralExpression(t, opExp.Left, left, true) {
		return false
	}
	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}
	if !testLiteralExpression(t, opExp.Right, right, true) {
		return false
	}
	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}
	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}
	if bo.TokenLiteral() != token.BooleanToString(value) {
		t.Errorf("bo.TokenLiteral not %t. got=%s",
			value, bo.TokenLiteral())
		return false
	}
	return true
}

func TestDeclerationAndAssignmentStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedType       string
		expectedValue      interface{}
	}{
		{"nambala x = 5;", "x", "nambala", 5},
		{"x = 5;", "x", "", 5},
		{"nambala y = zoona;", "y", "nambala", true},
		{"nambala foobar = y;", "foobar", "nambala", "y"},
		{"foobar[0] = 2;", "foobar", "", "y"},
		{"nambala result = linearSearch(arr, arr.length(), x);", "result", "", "linearSearch"},
		{"result = linearSearch(arr, arr.length(), x);", "result", "", "linearSearch"},
	}
	for _, tt := range tests {
		l := lexer.New([]byte(tt.input))
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}
		stmt := program.Statements[0]
		if !testDeclarationOrAssignmentStatement(t, stmt, tt.expectedIdentifier, tt.expectedType, tt.expectedValue) {
			return
		}
	}
}

func testDeclarationOrAssignmentStatement(t *testing.T, s ast.Statement, name string, varType string, value interface{}) bool {
	switch statement := s.(type) {
	case *ast.AssigmentStatement:
		{
			switch identifier := statement.Identifier.(type) {
			case *ast.Identifier:
				{
					if identifier.Value != name {
						t.Errorf("AssigmentStatement.Name.Value not '%s'. got=%s", name, identifier.Value)
						return false
					}
					if identifier.TokenLiteral() != name {
						t.Errorf("s.Name not '%s'. got=%s", name, identifier.Value)
						return false
					}
					return testLiteralExpression(t, statement.Value, value, true)
				}
			case *ast.IndexExpression:
				{
					if identifier.Left.String() != name {
						t.Errorf("AssigmentStatement.Name.Value not '%s'. got=%s", name, identifier.Left.String())
						return false
					}
				}
			}
			return false
		}
	case *ast.VariableDeclarationStatement:
		{
			if statement.Type.Literal != varType {
				t.Errorf("s.TokenLiteral not '%s'. got=%q", varType, statement.Type.Literal)
				return false
			}
			if statement.Identifier.Value != name {
				t.Errorf("VariableDeclarationStatement.Name.Value not '%s'. got=%s", name, statement.Identifier.Value)
				return false
			}
			if statement.Identifier.TokenLiteral() != name {
				t.Errorf("s.Name not '%s'. got=%s", name, statement.Identifier.Value)
				return false
			}
			switch exp := statement.Value.(type) {
			case *ast.CallExpression:
				{
					return true
				}
			default:
				{
					return testLiteralExpression(t, exp, value, true)
				}
			}
		}
	case *ast.ExpressionStatement:
		{
			s, ok := statement.Expression.(*ast.AssigmentStatement)
			if !ok {
				t.Errorf("s not *ast.AssigmentStatement. got=%T", s)
				return false
			}
			return testDeclarationOrAssignmentStatement(t, s, name, varType, value)
		}
	}
	t.Errorf("s not *ast.AssigmentStatement or *ast.VariableDeclarationStatement. got=%T", s)
	return false
}

func TestReturnStatements(t *testing.T) {
	input := `
	bweza 5;
	bweza 10;
	bweza 993322;
	`
	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}
	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.returnStatement. got=%T", stmt)
			continue
		}
		if returnStmt.TokenLiteral() != "bweza" {
			t.Errorf("returnStmt.TokenLiteral not 'bweza', got %q",
				returnStmt.TokenLiteral())
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"
	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}
	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}
	if ident.Value != "foobar" {
		t.Errorf("ident.Value not %s. got=%s", "foobar", ident.Value)
	}
	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral not %s. got=%s", "foobar",
			ident.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"
	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}
	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}
	if !literal.Value.Equal(decimal.NewFromInt(5)) {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}
	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral not %s. got=%s", "5",
			literal.TokenLiteral())
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`
	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value)
	}
}

func TestPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!zoona;", "!", true},
		{"!bodza;", "!", false},
	}
	for _, tt := range prefixTests {
		l := lexer.New([]byte(tt.input))
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}
		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PrefixExpression. got=%T", stmt.Expression)
		}
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s",
				tt.operator, exp.Operator)
		}
		if !testLiteralExpression(t, exp.Right, tt.value, true) {
			return
		}
	}
}

func TestInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5- 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"zoona == zoona", true, "==", true},
		{"zoona != bodza", true, "!=", false},
		{"bodza == bodza", false, "==", false},
		{"zoona && zoona", true, "&&", true},
		{"zoona && bodza", true, "&&", false},
		{"bodza && bodza", false, "&&", false},
		{"zoona || zoona", true, "||", true},
		{"zoona || bodza", true, "||", false},
		{"bodza || bodza", false, "||", false},
	}
	for _, tt := range infixTests {
		l := lexer.New([]byte(tt.input))
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}
		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("exp is not ast.InfixExpression. got=%T", stmt.Expression)
		}
		if !testLiteralExpression(t, exp.Left, tt.leftValue, true) {
			return
		}
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s",
				tt.operator, exp.Operator)
		}
		if !testLiteralExpression(t, exp.Right, tt.rightValue, true) {
			return
		}
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b- c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e- f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4;-5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"zoona",
			"zoona",
		},
		{
			"bodza",
			"bodza",
		},
		{
			"3 > 5 == bodza",
			"((3 > 5) == bodza)",
		},
		{
			"3 < 5 == zoona",
			"((3 < 5) == zoona)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(zoona == zoona)",
			"(!(zoona == zoona))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
		{
			"lo <= hi && x >= arr[lo] && x <= arr[hi]",
			"(((lo <= hi) && (x >= (arr[lo]))) && (x <= (arr[hi])))",
		},
	}

	for _, tt := range tests {
		l := lexer.New([]byte(tt.input))
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func TestBooleanExpression(t *testing.T) {
	input := "zoona;"
	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}
	ident, ok := stmt.Expression.(*ast.Boolean)
	if !ok {
		t.Fatalf("exp not *ast.Boolean. got=%T", stmt.Expression)
	}

	testLiteralExpression(t, ident, true, true)
}

func TestIfExpression(t *testing.T) {
	tests := []struct {
		input             string
		expectedOperation string
		identifier        string
	}{
		{input: `ngati (x < y) { x }`, expectedOperation: "<", identifier: "x"},
		{input: `ngati (x <= y) { x }`, expectedOperation: "<=", identifier: "x"},
		{input: `ngati (x >= y) { x }`, expectedOperation: ">=", identifier: "x"},
		{input: `ngati (x == y) { x }`, expectedOperation: "==", identifier: "x"},
		{input: `ngati (x[0] == y) { x }`, expectedOperation: "=="},
		{input: `ngati (x[i] == y) { x }`, expectedOperation: "=="},
	}
	for _, tt := range tests {
		l := lexer.New([]byte(tt.input))
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Body does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}
		exp, ok := stmt.Expression.(*ast.IfExpression)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T",
				stmt.Expression)
		}
		if !testInfixExpression(t, exp.Condition, "x", tt.expectedOperation, "y") {
			return
		}
		if len(exp.Consequence.Statements) != 1 {
			t.Errorf("consequence is not 1 statements. got=%d\n",
				len(exp.Consequence.Statements))
		}
		consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
				exp.Consequence.Statements[0])
		}
		if !testIdentifier(t, consequence.Expression, "x") {
			return
		}
		if exp.Alternative != nil {
			t.Errorf("exp.Alternative.Statements was not nil. got=%+v", exp.Alternative)
		}
	}

}

func TestIfElseExpression(t *testing.T) {
	input := `ngati (x < y) { x } kapena { y }`

	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d\n",
			len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if len(exp.Alternative.Statements) != 1 {
		t.Errorf("exp.Alternative.Statements does not contain 1 statements. got=%d\n",
			len(exp.Alternative.Statements))
	}

	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Alternative.Statements[0])
	}

	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestFunctionLiteral(t *testing.T) {
	input := `ndondomeko phatikiza(x, y) { x + y; }`
	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}
	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T",
			stmt.Expression)
	}
	if len(function.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d\n",
			len(function.Parameters))
	}
	if function.Name.TokenLiteral() != "phatikiza" {
		t.Fatalf("function name wrong. expected phatikiza got=%s\n", function.Name.TokenLiteral())
	}
	testLiteralExpression(t, function.Parameters[0], "x", true)
	testLiteralExpression(t, function.Parameters[1], "y", true)
	if len(function.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements has not 1 statements. got=%d\n",
			len(function.Body.Statements))
	}
	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body stmt is not ast.ExpressionStatement. got=%T",
			function.Body.Statements[0])
	}
	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionParameter(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "ndondomeko palibe() {};", expectedParams: []string{}},
		{input: "ndondomeko palibe(x) {};", expectedParams: []string{"x"}},
		{input: "ndondomeko palibe(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}
	for _, tt := range tests {
		l := lexer.New([]byte(tt.input))
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		stmt := program.Statements[0].(*ast.ExpressionStatement)
		function := stmt.Expression.(*ast.FunctionLiteral)
		if len(function.Parameters) != len(tt.expectedParams) {
			t.Errorf("length parameters wrong. want %d, got=%d\n",
				len(tt.expectedParams), len(function.Parameters))
		}
		for i, ident := range tt.expectedParams {
			testLiteralExpression(t, function.Parameters[i], ident, true)
		}
	}
}

func TestCallExpression(t *testing.T) {
	type CallValue struct {
		value      interface{}
		otherValue interface{}
		operator   string
	}
	tests := []struct {
		input          string
		identifier     string
		expectedParams []CallValue
	}{
		{input: "add(1, 2 * 3, 4 + 5);", identifier: "add", expectedParams: []CallValue{
			{value: 1, operator: ""},
			{value: 2, operator: "*", otherValue: 3},
			{value: 4, operator: "+", otherValue: 5},
		}},
		{input: "palibe(x.length(), y[0]);", identifier: "palibe", expectedParams: []CallValue{
			{value: "x", operator: "length", otherValue: 0},
			{value: "y", otherValue: 0},
		}},
		{input: "palibe(y[0]);", identifier: "palibe", expectedParams: []CallValue{
			{value: "y", otherValue: 0},
		}},
		{input: "palibe(x.length());", identifier: "palibe", expectedParams: []CallValue{
			{value: "x", operator: "length", otherValue: 0},
		}},
		{input: "palibe(arr, arr.length(), x);", identifier: "palibe", expectedParams: []CallValue{
			{value: "arr"},
			{value: "arr", operator: "length"},
			{value: "x"},
		}},
	}
	for _, tt := range tests {
		l := lexer.New([]byte(tt.input))
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("stmt is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}
		exp, ok := stmt.Expression.(*ast.CallExpression)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
				stmt.Expression)
		}
		if !testIdentifier(t, exp.Function, tt.identifier) {
			return
		}
		if len(exp.Arguments) != len(tt.expectedParams) {
			t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
		}
		for idx, arg := range exp.Arguments {
			call := tt.expectedParams[idx]
			switch arg := arg.(type) {
			case *ast.InfixExpression:
				{
					testInfixExpression(t, arg, call.value, call.operator, call.otherValue)
				}
			case *ast.IntegerLiteral:
				{
					testLiteralExpression(t, arg, call.value, true)
				}
			case *ast.Identifier:
				{
					testIdentifier(t, arg, call.value.(string))
				}
			case *ast.MethodExpression:
				{
					if !testIdentifier(t, arg.Left, call.value.(string)) {
						t.Fatalf("exp not *ast.Identifier. got=%T", arg.Left)
					}
					if !testIdentifier(t, arg.Method, call.operator) {
						t.Fatalf("exp not *ast.Identifier. got=%T", arg.Method)
					}
				}
			}
		}
	}

}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"
	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("exp not ast.ArrayLiteral. got=%T", stmt.Expression)
	}
	if len(array.Elements) != 3 {
		t.Fatalf("len(array.Elements) not 3. got=%d", len(array.Elements))
	}
	testIntegerLiteral(t, array.Elements[0], 1)
	testInfixExpression(t, array.Elements[1], 2, "*", 2)
	testInfixExpression(t, array.Elements[2], 3, "+", 3)
}

func TestArrayIndexExpressions(t *testing.T) {
	input := "myArray[1 + 1]"
	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	indexExp, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("exp not *ast.IndexExpression. got=%T", stmt.Expression)
	}
	if !testIdentifier(t, indexExp.Left, "myArray") {
		return
	}
	if !testInfixExpression(t, indexExp.Index, 1, "+", 1) {
		return
	}
}

func TestMethodCall(t *testing.T) {
	tests := []struct {
		input              string
		expectedArguments  []interface{}
		expectedIdentifier string
		expectedMethodName string
	}{
		{input: "myArray.length();", expectedArguments: []interface{}{}, expectedIdentifier: "myArray", expectedMethodName: "length"},
		{input: "myString.substring(0,2);", expectedArguments: []interface{}{0, 2}, expectedIdentifier: "myString", expectedMethodName: "substring"},
		{input: "myClass.doStuff(\"String\",2);", expectedArguments: []interface{}{"String", 2}, expectedIdentifier: "myClass", expectedMethodName: "doStuff"},
	}
	for _, tt := range tests {
		l := lexer.New([]byte(tt.input))
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		stmt := program.Statements[0].(*ast.ExpressionStatement)
		method := stmt.Expression.(*ast.MethodExpression)
		if len(method.Arguments) != len(tt.expectedArguments) {
			t.Errorf("length parameters wrong. want %d, got=%d\n",
				len(tt.expectedArguments), len(method.Arguments))
		}
		for i, ident := range tt.expectedArguments {
			testLiteralExpression(t, method.Arguments[i], ident, false)
		}
		if !testIdentifier(t, method.Left, tt.expectedIdentifier) {
			return
		}
		if !testIdentifier(t, method.Method, tt.expectedMethodName) {
			return
		}
	}
}

// TODO: Add method calls tests with infix operation

func TestForExpression(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		incrementIsAssign  bool
	}{
		{
			input:              `za (nambala x = 0; x < 10; x = x + 1) { true }`,
			expectedIdentifier: "x",
			incrementIsAssign:  true,
		},
		{
			input:              `za (nambala x = 0; x < 10; x++) { true }`,
			expectedIdentifier: "x",
			incrementIsAssign:  false,
		},
		{
			input:              `nambala x = 0; za (x = 0; x < 10; x++) { true }`,
			expectedIdentifier: "x",
			incrementIsAssign:  false,
		},
	}
	for _, tt := range tests {
		l := lexer.New([]byte(tt.input))
		p := New(l)
		program := p.ParseProgram()

		checkParserErrors(t, p)

		loopIndex := 0

		if len(program.Statements) != 1 {
			loopIndex = 1
		}

		statement, ok := program.Statements[loopIndex].(*ast.ExpressionStatement)

		if !ok {
			t.Fatalf("program.Statements[%T] is not ast.Expression. got=%T", loopIndex, program.Statements[0])
		}

		expression, ok := statement.Expression.(*ast.ForExpression)

		if !ok {
			t.Fatalf("statement.Expression is not ast.For. got=%T", statement.Expression)
		}

		if !testIdentifier(t, expression.Identifier, tt.expectedIdentifier) {
			return
		}

		if loopIndex == 1 {
			if _, ok = expression.Initializer.(*ast.AssigmentStatement); !ok {
				t.Fatalf("expression.Initializer is not ast.AssigmentStatement. got=%T", expression.Initializer)
			}
		} else {
			if _, ok = expression.Initializer.(*ast.VariableDeclarationStatement); !ok {
				t.Fatalf("expression.Initializer is not ast.VariableDeclarationStatement. got=%T", expression.Initializer)
			}
		}

		if tt.incrementIsAssign {
			if _, ok = expression.Increment.(*ast.AssigmentStatement); !ok {
				t.Fatalf("expression.Increment is not ast.Assign. got=%T", expression.Increment)
			}
		} else {
			if _, ok = expression.Increment.(*ast.PostfixExpression); !ok {
				t.Fatalf("expression.Increment is not ast.PostfixExpression. got=%T", expression.Increment)
			}
		}

		if _, ok = expression.Block.Statements[0].(ast.Expression); !ok {
			t.Fatalf("expression.Block.Statements[0] is not ast.Expression. got=%T", expression.Block.Statements[0])
		}
	}

}

func TestWhileExpressions(t *testing.T) {
	input := `pamene (x < 10) { x = x + 1; }`
	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	whileExp, ok := stmt.Expression.(*ast.WhileExpression)
	if !ok {
		t.Fatalf("exp not *ast.WhileExpression. got=%T", stmt.Expression)
	}
	if !testInfixExpression(t, whileExp.Condition, "x", "<", 10) {
		return
	}
	if len(whileExp.Consequence.Statements) != 1 {
		t.Fatalf("whileExp.Block.Statements does not contain 1 statements. got=%d\n",
			len(whileExp.Consequence.Statements))
	}

	if whileExp.Consequence == nil {
		t.Fatalf("body is not *ast.BlockStatement. got=%T", whileExp.Consequence)
	}
}

func TestMapExpressionsWithStringKeys(t *testing.T) {
	tests := []struct {
		input      string
		identifier string
		length     int
		values     map[string]int64
	}{
		{
			input:      `mgwirizano grades = {"one": 1, "two": 2, "three": 3};`,
			identifier: "grades",
			length:     3,
			values:     map[string]int64{"one": 1, "two": 2, "three": 3},
		},
		{
			input:      `mgwirizano grades = {"one": 1, "two": 2, "three": 3}`,
			identifier: "grades",
			length:     3,
			values:     map[string]int64{"one": 1, "two": 2, "three": 3},
		},
		{
			input:      `mgwirizano grades = {};`,
			identifier: "grades",
			length:     0,
			values:     map[string]int64{},
		},
		{
			input:      `mgwirizano grades = {}`,
			identifier: "grades",
			length:     0,
			values:     map[string]int64{},
		},
		{
			input:      `{}`,
			identifier: "",
			length:     0,
			values:     map[string]int64{},
		},
		{
			input:      `{"one": 1, "two": 2, "three": 3};`,
			identifier: "",
			length:     3,
			values:     map[string]int64{"one": 1, "two": 2, "three": 3},
		},
	}

	for _, tt := range tests {
		l := lexer.New([]byte(tt.input))
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		var mapLiteral *ast.MapExpression
		switch stmt := program.Statements[0].(type) {
		case *ast.VariableDeclarationStatement:
			{
				expression, ok := stmt.Value.(*ast.MapExpression)
				if !ok {
					t.Fatalf("exp not *ast.MapExpression. got=%T", stmt.Value)
				}
				mapLiteral = expression
				if stmt.Identifier.Value != tt.identifier {
					t.Fatalf("exp not %s. got=%s", tt.identifier, stmt.Identifier.Value)
				}
			}
		case *ast.ExpressionStatement:
			{
				expression, ok := stmt.Expression.(*ast.MapExpression)
				if !ok {
					t.Fatalf("exp not *ast.MapExpression. got=%T", stmt.Expression)
				}
				mapLiteral = expression
			}
		default:
			{
				t.Fatalf("exp not *ast.MapExpression or *ast.VariableDeclarationStatement. got=%T", stmt)
				return
			}

		}

		if len(mapLiteral.Pairs) != tt.length {
			t.Fatalf("map.Pairs has wrong length. got=%d", len(mapLiteral.Pairs))
		}

		for key, value := range mapLiteral.Pairs {
			identifier, ok := key.(*ast.StringLiteral)

			if !ok {
				t.Errorf("key is not ast.StringLiteral. got=%T", key)
			}

			expectedValue := tt.values[identifier.Value]

			testIntegerLiteral(t, value, expectedValue)
		}
	}
}

func TestMapLiteralsWithStringKeys(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)

	if !ok {
		t.Fatalf("program.Statements[0] is not ast.Expression. got=%T", program.Statements[0])
	}

	mapLiteral, ok := statement.Expression.(*ast.MapExpression)

	if !ok {
		t.Fatalf("statement is not ast.Map. got=%T", statement.Expression)
	}

	if len(mapLiteral.Pairs) != 3 {
		t.Fatalf("map.Pairs has wrong length. got=%d", len(mapLiteral.Pairs))
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for key, value := range mapLiteral.Pairs {
		literal, ok := key.(*ast.StringLiteral)

		if !ok {
			t.Errorf("key is not ast.String. got=%T", key)
		}

		expectedValue := expected[literal.Value]

		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestMapLiteralsWithBooleanKeys(t *testing.T) {
	input := `{zoona: 1, bodza: 2}`

	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)

	if !ok {
		t.Fatalf("program.Statements[0] is not ast.Expression. got=%T", program.Statements[0])
	}

	mapLiteral, ok := statement.Expression.(*ast.MapExpression)

	if !ok {
		t.Fatalf("statement is not ast.Map. got=%T", statement.Expression)
	}

	if len(mapLiteral.Pairs) != 2 {
		t.Fatalf("map.Pairs has wrong length. got=%d", len(mapLiteral.Pairs))
	}

	expected := map[bool]int64{
		true:  1,
		false: 2,
	}

	for key, value := range mapLiteral.Pairs {
		boolean, ok := key.(*ast.Boolean)

		if !ok {
			t.Errorf("key is not ast.Boolean. got=%T", key)
		}

		expectedValue := expected[boolean.Value]

		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestMapLiteralsWithIntegerKeys(t *testing.T) {
	input := `{1: 1, 2: 2, 3: 3}`

	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)

	if !ok {
		t.Fatalf("program.Statements[0] is not ast.Expression. got=%T", program.Statements[0])
	}

	mapLiteral, ok := statement.Expression.(*ast.MapExpression)

	if !ok {
		t.Fatalf("statement is not ast.Map. got=%T", statement.Expression)
	}

	if len(mapLiteral.Pairs) != 3 {
		t.Fatalf("map.Pairs has wrong length. got=%d", len(mapLiteral.Pairs))
	}

	expected := map[int64]int64{
		1: 1,
		2: 2,
		3: 3,
	}

	for key, value := range mapLiteral.Pairs {
		number, ok := key.(*ast.IntegerLiteral)

		if !ok {
			t.Errorf("key is not ast.Number. got=%T", key)
		}

		expectedValue := expected[number.Value.IntPart()]

		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestMapLiteralsWithVariableKeys(t *testing.T) {
	input := `{foo: 1, bar: 2, baz: 3}`

	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)

	if !ok {
		t.Fatalf("program.Statements[0] is not ast.Expression. got=%T", program.Statements[0])
	}

	mapLiteral, ok := statement.Expression.(*ast.MapExpression)

	if !ok {
		t.Fatalf("statement is not ast.Map. got=%T", statement.Expression)
	}

	if len(mapLiteral.Pairs) != 3 {
		t.Fatalf("map.Pairs has wrong length. got=%d", len(mapLiteral.Pairs))
	}

	expected := map[string]int64{
		"foo": 1,
		"bar": 2,
		"baz": 3,
	}

	for key, value := range mapLiteral.Pairs {
		identifier, ok := key.(*ast.Identifier)

		if !ok {
			t.Errorf("key is not ast.Identifier. got=%T", key)
		}

		expectedValue := expected[identifier.Value]

		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestParsingHashLiteralsWithExpressions(t *testing.T) {
	input := `{"one": 0 + 1, "two": 10- 8, "three": 15 / 5}`
	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.MapExpression)
	if !ok {
		t.Fatalf("exp is not ast.HashLiteral. got=%T", stmt.Expression)
	}
	if len(hash.Pairs) != 3 {
		t.Errorf("hash.Pairs has wrong length. got=%d", len(hash.Pairs))
	}
	tests := map[string]func(ast.Expression){
		"one": func(e ast.Expression) {
			testInfixExpression(t, e, 0, "+", 1)
		},
		"two": func(e ast.Expression) {
			testInfixExpression(t, e, 10, "-", 8)
		},
		"three": func(e ast.Expression) {
			testInfixExpression(t, e, 15, "/", 5)
		},
	}
	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not ast.StringLiteral. got=%T", key)
			continue
		}
		testFunc, ok := tests[literal.String()]
		if !ok {
			t.Errorf("No test function for key %q found", literal.String())
			continue
		}
		testFunc(value)
	}
}

func TestParsingClassExpressions(t *testing.T) {
	input := `
		kalasi Person { 
		}`
	l := lexer.New([]byte(input))
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt := program.Statements[0].(*ast.ClassStatement)
	if stmt.Name.Value != "Person" {
		t.Fatalf("stmt.Name.Value not 'Person'. got=%q", stmt.Name.Value)
	}
	if len(stmt.Body.Statements) != 1 {
		t.Fatalf("stmt.Body.Statements does not contain 1 statements. got=%d\n",
			len(stmt.Body.Statements))
	}
	bodyStmt, ok := stmt.Body.Statements[0].(*ast.VariableDeclarationStatement)
	if !ok {
		t.Fatalf("stmt.Body.Statements[0] is not ast.VariableDeclarationStatement. got=%T",
			stmt.Body.Statements[0])
	}
	if bodyStmt.Identifier.Value != "age" {
		t.Fatalf("bodyStmt.Identifier.Value not 'age'. got=%s", bodyStmt.Identifier.Value)
	}
}
