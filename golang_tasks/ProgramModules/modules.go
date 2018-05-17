package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode"
)

//TokenType тип представляющий константы типа токена
type TokenType int

//Констаты типов токена
const (
	EmptyValue      TokenType = iota
	NUMBER          TokenType = iota
	IDENT           TokenType = iota
	OPERATION       TokenType = iota
	COMPARE         TokenType = iota
	TERNAR_QUESTION TokenType = iota
	TERNAR_COLON    TokenType = iota
	LBRACKET        TokenType = iota
	RBRACKET        TokenType = iota
	COMMA           TokenType = iota
	SEMICOLON       TokenType = iota
	FUNCTION_BODY   TokenType = iota
)

//Token тип представлющий токен при разборе формулы
type Token struct {
	Content string
	Type    TokenType
}

//Tokenizer хранит данные необходимые для выделения токенов
type Tokenizer struct {
	reader       *bufio.Reader
	currentToken *Token
}

//NewTokenizer инициализирует токенайзер строкой
func NewTokenizer(in io.Reader) *Tokenizer {
	return &Tokenizer{
		reader: bufio.NewReader(in),
	}
}

func (t *Tokenizer) readLexem(stopPred func(rune) bool) string {
	buff := bytes.Buffer{}
	for {
		r, _, err := t.reader.ReadRune()
		if err != nil {
			break
		}
		if stopPred(r) {
			t.reader.UnreadRune()
			break
		}

		buff.WriteRune(r)
	}

	return buff.String()
}

func (t *Tokenizer) next() (Token, error) {
	//Будет постоянно возвращать EOF, когда текст программы закончился
	if _, err := t.reader.Peek(1); err != nil {
		return Token{}, io.EOF
	}

	//Пропускаем пробелы
	for {
		r, _, err := t.reader.ReadRune()
		if err != nil {
			return Token{}, io.EOF
		}

		if !unicode.IsSpace(r) {
			t.reader.UnreadRune()
			break
		}
	}

	if x, _, _ := t.reader.ReadRune(); unicode.IsDigit(x) {
		//Числовые константы
		t.reader.UnreadRune()
		return Token{
			Content: t.readLexem(func(r rune) bool {
				return !unicode.IsDigit(r)
			}),
			Type: NUMBER,
		}, nil
	} else if unicode.IsLetter(x) {
		t.reader.UnreadRune()
		return Token{
			Content: t.readLexem(func(r rune) bool {
				return !(unicode.IsDigit(r) || unicode.IsLetter(r))
			}),
			Type: IDENT,
		}, nil
	} else if x == '+' || x == '-' || x == '/' || x == '*' {
		return Token{Content: string(x), Type: OPERATION}, nil
	} else if x == '=' {
		return Token{Content: string(x), Type: COMPARE}, nil
	} else if x == '>' {
		if op, err := t.reader.Peek(1); err != nil || op[0] != '=' {
			return Token{Content: string(x), Type: COMPARE}, nil
		}

		t.reader.ReadRune()
		return Token{Content: ">=", Type: COMPARE}, nil
	} else if x == '<' {
		op, err := t.reader.Peek(1)
		if err != nil ||
			(op[0] != '=' && op[0] != '>') {
			return Token{Content: string(x), Type: COMPARE}, nil
		}

		t.reader.ReadRune()
		buff := bytes.Buffer{}
		buff.WriteRune(x)
		buff.WriteByte(op[0])
		return Token{Content: buff.String(), Type: COMPARE}, nil
	} else if x == '?' {
		return Token{Content: string(x), Type: TERNAR_QUESTION}, nil
	} else if x == ':' {
		if op, err := t.reader.Peek(1); err != nil || op[0] != '=' {
			return Token{Content: string(x), Type: TERNAR_COLON}, nil
		}

		t.reader.ReadRune()
		return Token{Content: ":=", Type: FUNCTION_BODY}, nil
	} else if x == '(' {
		return Token{Content: string(x), Type: LBRACKET}, nil
	} else if x == ')' {
		return Token{Content: string(x), Type: RBRACKET}, nil
	} else if x == ',' {
		return Token{Content: string(x), Type: COMMA}, nil
	} else if x == ';' {
		return Token{Content: string(x), Type: SEMICOLON}, nil
	} else {
		return Token{}, fmt.Errorf("next(): syntax error: wrong token")
	}
}

//Peek возвращает следующий токен, но оставляет его в потоке.
func (t *Tokenizer) Peek() (Token, error) {
	if t.currentToken == nil {
		token, err := t.next()
		if err != nil {
			return Token{}, err
		}

		t.currentToken = &token
	}

	return *t.currentToken, nil
}

//Read забирает из потока и возвращает токен
func (t *Tokenizer) Read() (Token, error) {
	if t.currentToken == nil {
		token, err := t.next()
		if err != nil {
			return Token{}, err
		}

		return token, nil
	} else {
		token := t.currentToken
		t.currentToken = nil
		return *token, nil
	}
}

//Function представляет функцию
type Function struct {
	Ident      string
	FormalArgs map[string]bool
	Argc       int

	CallsOtherFunctions map[string]int

	TimeIn   int
	ModuleID int
	Low      int
}

// Нисходящий рекурсивный парсер. (Грамматика в задании)
func parseProgram(functions map[string]*Function, t *Tokenizer) error {
	if err := parseFunction(functions, t); err != nil {
		return err
	}

	if _, err := t.Peek(); err != nil {
		return nil
	}

	return parseProgram(functions, t)
}

func parseFunction(functions map[string]*Function, t *Tokenizer) error {
	token, err := t.Read()
	if err != nil {
		return fmt.Errorf("parseFunction(): syntax error: unexpected EOF")
	}

	if token.Type == IDENT {
		if _, ok := functions[token.Content]; ok {
			return fmt.Errorf("parseFunction(): syntax error: duplicate function: %s",
				token.Content)
		}

		function := Function{
			Ident:               token.Content,
			FormalArgs:          make(map[string]bool),
			CallsOtherFunctions: make(map[string]int, 0)}
		token, err := t.Read()
		if err != nil {
			return fmt.Errorf("parseFunction(): syntax error: unknown error: %v",
				err)
		}

		if token.Type != LBRACKET {
			return fmt.Errorf("parseFunction(): syntax error: unexpected token: %s",
				token.Content)
		}

		if err := parseFormalArgsList(&function, t); err != nil {
			return err
		}

		token, err = t.Read()
		if err != nil {
			return fmt.Errorf("parseFunction(): syntax error: unexpected EOF")
		}

		if token.Type != RBRACKET {
			return fmt.Errorf("parseFunction(): syntax error: unexpected token: %s",
				token.Content)
		}

		function.Argc = len(function.FormalArgs)
		functions[function.Ident] = &function

		token, err = t.Read()
		if err != nil {
			return fmt.Errorf("parseFunction(): syntax error: unknown error: %v",
				err)
		}

		if token.Type != FUNCTION_BODY {
			return fmt.Errorf("parseFunction(): syntax error: unexpected token: %s",
				token.Content)
		}

		if err := parseExpr(&function, t); err != nil {
			return err
		}

		token, err = t.Read()
		if err != nil {
			return fmt.Errorf("parseFunction(): syntax error: unknown error: %v",
				err)
		}

		if token.Type != SEMICOLON {
			return fmt.Errorf("parseFunction(): syntax error: unexpected token: %s",
				token.Content)
		}
	} else {
		return fmt.Errorf("parseFunction(): syntax error: unexpected token: %s",
			token.Content)
	}

	return nil
}

func parseFormalArgsList(function *Function, t *Tokenizer) error {
	tok, err := t.Peek()
	if err != nil {
		return nil
	}

	if tok.Type != IDENT {
		return nil
	}

	return parseIdentList(function, t)
}

func parseIdentList(function *Function, t *Tokenizer) error {
	token, err := t.Read()
	if err != nil {
		return fmt.Errorf("parseIdentList(): syntax error: unexpected EOF")
	}

	if token.Type != IDENT {
		return fmt.Errorf("parseIdentList(): syntax error: unexpected token: %s",
			token.Content)
	}

	if function.FormalArgs[token.Content] {
		return fmt.Errorf("parseIdentList(): syntax error: duplicate arg: %s",
			token.Content)
	}

	function.FormalArgs[token.Content] = true

	token, err = t.Peek()
	if err != nil {
		return nil
	}

	if token.Type == COMMA {
		t.Read()
		return parseIdentList(function, t)
	}

	return nil
}

func parseExpr(function *Function, t *Tokenizer) error {
	if err := parseComparisonExpr(function, t); err != nil {
		return err
	}

	token, err := t.Peek()
	if err != nil {
		return nil
	}

	if token.Type == TERNAR_QUESTION {
		t.Read()
		if err := parseComparisonExpr(function, t); err != nil {
			return err
		}

		token, err = t.Read()
		if err != nil {
			return err
		}

		if token.Type != TERNAR_COLON {
			return fmt.Errorf("parseExpr(): syntax error: unexpected token: %s",
				token.Content)
		}

		return parseExpr(function, t)
	}

	return nil
}

func parseComparisonExpr(function *Function, t *Tokenizer) error {
	if err := parseArithExpt(function, t); err != nil {
		return err
	}

	token, err := t.Peek()
	if err != nil {
		return nil
	}

	if token.Type == COMPARE {
		t.Read()
		return parseArithExpt(function, t)
	}

	return nil
}

func parseArithExpt(function *Function, t *Tokenizer) error {
	if err := parseTerm(function, t); err != nil {
		return err
	}

	token, err := t.Peek()
	if err != nil {
		return nil
	}

	if token.Type == OPERATION &&
		(token.Content == "+" || token.Content == "-") {
		t.Read()
		return parseArithExpt(function, t)
	}

	return nil
}

func parseTerm(function *Function, t *Tokenizer) error {
	if err := parseFactor(function, t); err != nil {
		return err
	}

	token, err := t.Peek()
	if err != nil {
		return nil
	}

	if token.Type == OPERATION &&
		(token.Content == "*" || token.Content == "/") {
		t.Read()
		return parseTerm(function, t)
	}

	return nil
}

func parseFactor(function *Function, t *Tokenizer) error {
	token, err := t.Read()
	if err != nil {
		return fmt.Errorf("parseFactor(): syntax error: unexpected EOF")
	}

	if token.Type == OPERATION && token.Content == "-" {
		return parseFactor(function, t)
	} else if token.Type == IDENT {
		tok, err := t.Peek()
		if err != nil {
			return nil
		}

		if tok.Type == LBRACKET { // Вызов функции
			t.Read()
			argc, err := parseActualArgsList(function, t)
			if err != nil {
				return err
			}

			function.CallsOtherFunctions[token.Content] = argc

			tok, err := t.Read()
			if err != nil {
				return fmt.Errorf("parseFactor(): syntax error: unexpected EOF")
			}

			if tok.Type != RBRACKET {
				return fmt.Errorf("parseFactor(): syntax error: unexpected token: %s",
					token.Content)
			}
		} else { // Просто переменная
			if !function.FormalArgs[token.Content] {
				return fmt.Errorf("parseFactor(): syntax error: can not resolve: %s",
					token.Content)
			}
		}
	} else if token.Type == LBRACKET {
		if err := parseExpr(function, t); err != nil {
			return err
		}

		token, err := t.Read()
		if err != nil {
			return fmt.Errorf("parseFactor(): syntax error: unexpected EOF")
		}

		if token.Type != RBRACKET {
			return fmt.Errorf("parseFactor(): syntax error: unexpected token: %s",
				token.Content)
		}
	} else if token.Type == NUMBER {
		//Представим, что делаем что-то полезное с NUMBER
	} else {
		return fmt.Errorf("parseFactor(): syntax error: unexpected token: %s",
			token.Content)
	}

	return nil
}

func parseActualArgsList(function *Function, t *Tokenizer) (int, error) {
	tok, err := t.Peek()
	if err != nil {
		return 0, nil
	}

	if tok.Type == RBRACKET {
		return 0, nil
	}

	return parseExprList(function, t)
}

func parseExprList(function *Function, t *Tokenizer) (int, error) {
	if err := parseExpr(function, t); err != nil {
		return 0, err
	}

	token, err := t.Peek()
	if err != nil {
		return 0, nil
	}

	if token.Type == COMMA {
		t.Read()
		argc, err := parseExprList(function, t)
		return argc + 1, err
	}

	return 1, nil
}

//Алгоритм Тарьяна поиска компонент сильной связности.
func dfs(currentFunction *Function, functions map[string]*Function, stack *[]*Function,
	time *int, moduleID *int) error {
	currentFunction.TimeIn = *time
	currentFunction.Low = *time
	*stack = append(*stack, currentFunction)
	(*time)++

	for call, argc := range currentFunction.CallsOtherFunctions {
		callFunction, ok := functions[call]
		if !ok {
			return fmt.Errorf("dfs(): unknown function %s", call)
		}

		if callFunction.Argc != argc {
			return fmt.Errorf("dfs(): syntax error: %s expected %d args, have %d values",
				call, callFunction.Argc, argc)
		}

		if callFunction.TimeIn == 0 {
			if err := dfs(callFunction, functions, stack, time, moduleID); err != nil {
				return err
			}
		}

		if callFunction.ModuleID == 0 &&
			currentFunction.Low > callFunction.Low {
			currentFunction.Low = callFunction.Low
		}
	}

	if currentFunction.TimeIn == currentFunction.Low {
		for {
			stackLen := len(*stack)
			nextFunc := (*stack)[stackLen-1]
			*stack = (*stack)[:stackLen-1]
			nextFunc.ModuleID = *moduleID

			if nextFunc == currentFunction {
				break
			}
		}

		(*moduleID)++
	}

	return nil
}

func findRecursiveDependencies(functions map[string]*Function) (int, error) {
	stack := make([]*Function, 0)

	time := 1
	moduleID := 1
	for _, function := range functions {
		if function.TimeIn == 0 {
			if err := dfs(function, functions, &stack, &time, &moduleID); err != nil {
				return 0, err
			}
		}
	}

	return moduleID - 1, nil
}

func main() {
	tokenizer := NewTokenizer(os.Stdin)
	functions := make(map[string]*Function)
	if err := parseProgram(functions, tokenizer); err != nil {
		fmt.Println("error")
		return
	}

	result, err := findRecursiveDependencies(functions)
	if err != nil {
		fmt.Println("error")
		return
	}

	fmt.Println(result)
}
