package custom

import (
	"fmt"
	"io"
	"journal-proxy/parser/custom/types"
	"lib/utils/maps"
	mys "lib/utils/scanner"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"
)

type Lexer struct {
	mys.Scanner
}

func NewLexer(rd io.Reader) *Lexer {
	return &Lexer{
		Scanner: *mys.NewScanner(rd),
	}
}

// Run lexer and return all tokens
func (lex *Lexer) TokenList() []types.Token {
	var list []types.Token
	for token := lex.NextToken(); token != nil; token = lex.NextToken() {
		list = append(list, token)
	}
	return list
}

func (lex *Lexer) NextToken() types.Token {
	t := lex.Peek()
	if t == scanner.EOF {
		return nil
	}

	// tag token
	if t == '{' {
		token := lex.tag()
		if token != nil {
			return token
		}
	}

	// ignore token
	token := lex.ignore()
	if token != nil {
		return token
	}
	return nil
}

// scan tag token
func (lex *Lexer) tag() types.Token {
	t := lex.Peek()
	if t == scanner.EOF {
		return nil
	}
	lex.Next()

	if t != '{' {
		return types.TokenObject{Typ: types.Error, Pos: lex.Position, Val: fmt.Sprintf("expected '{' got '%c'", t)}
	}

	// eat '{'
	t = lex.Next()
	if t != '{' {
		return types.TokenObject{Typ: types.Error, Pos: lex.Position, Val: fmt.Sprintf("expected '{' got '%c'", t)}
	}

	// Scan tag value
	val := lex.ReadUntil(func(r rune) bool {
		return !unicode.IsLetter(r) && r != '_'
	})

	// read properties
	props, err := lex.properties()
	if err != nil {
		return types.TokenObject{Typ: types.Error, Pos: lex.Position, Val: err.Error()}
	}
	propKeys := make([]types.PropertyType, len(props))
	for i, prop := range props {
		propKeys[i] = prop.Type
	}

	// eat '}'
	t = lex.Next()
	if t != '}' {
		return types.TokenObject{Typ: types.Error, Pos: lex.Position, Val: fmt.Sprintf("expected '}' got '%c'", t)}
	}

	// eat '}'
	t = lex.Next()
	if t != '}' {
		return types.TokenObject{Typ: types.Error, Pos: lex.Position, Val: fmt.Sprintf("expected '}' got '%c'", t)}
	}

	// get next char as delimeter
	delim := lex.PeekEscaped()

	return types.TokenObject{Typ: types.Tag, Val: val, Pos: lex.Position, Delim: delim, Props: maps.NewWeightedMapFromSlice(propKeys, props)}
}

// scan ignore token
func (lex *Lexer) ignore() types.Token {
	t := lex.Peek()
	if t == scanner.EOF {
		return nil
	}

	return types.TokenObject{Typ: types.Ignore, Val: lex.ReadUntil(func(r rune) bool { return r == '{' }), Pos: lex.Position}
}

func (lex *Lexer) properties() ([]types.TokenProperty, error) {
	end := func(r rune) bool {
		return r == '|' || r == '}'
	}

	notWord := func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsSpace(r)
	}

	// skip whitespaces
	if space := strings.TrimSpace(lex.ReadUntil(end)); space != "" {
		// string found
		return nil, fmt.Errorf("expected '}' or '|', got '%s'", space)
	}

	if lex.Peek() != '|' {
		// No properties
		return nil, nil
	}

	// eat '|'
	lex.Next()

	var properties []types.TokenProperty

	for prop := lex.ReadUntil(notWord); ; prop = lex.ReadUntil(notWord) {
		prop := types.PropertyType(strings.TrimSpace(prop))

		val := lex.ReadUntil(end)
		val = strings.Trim(val, "' \t\n")

		switch prop {
		case types.IncludeProp:
			num, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
			properties = append(properties, types.TokenProperty{Type: prop, Value: num})

		case types.TrimProp, types.TrimLeftProp, types.TrimRightProp:
			properties = append(properties, types.TokenProperty{Type: prop, Value: val})
		}

		t := lex.Peek()
		if t == scanner.EOF || t == '}' {
			break
		}
	}

	return properties, nil
}

func (lex *Lexer) ReadUntil(until func(rune) bool) string {
	sb := strings.Builder{}

	for t := lex.PeekEscaped(); t != scanner.EOF; t = lex.PeekEscaped() {
		if until(t) {
			break
		}

		sb.WriteRune(t)
		lex.NextEscaped()
	}
	return sb.String()
}
