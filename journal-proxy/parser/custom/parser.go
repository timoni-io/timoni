package custom

import (
	"fmt"
	"io"
	"journal-proxy/global"
	"journal-proxy/parser/custom/types"
	"lib/utils/conv"
	"lib/utils/math"
	"strconv"
	"strings"
	"text/scanner"
)

type Parser struct {
	scanner.Scanner
	tokens []types.Token
}

func NewParser(rd io.Reader) *Parser {
	tkns := NewLexer(rd).TokenList()
	for _, t := range tkns {
		if t.Type() == types.Error {
			fmt.Printf("couldn't setup parser. Lexer error %s @ %s", t.Value(), t.Position())
			return nil
		}
	}
	return &Parser{
		tokens: tkns,
	}
}

func (p Parser) Parse(entry *global.Entry) error {
	if entry.Message == "" {
		return nil
	}

	p.Init(strings.NewReader(entry.Message))

	for _, token := range p.tokens {
		switch token.Type() {
		case types.Tag:
			include := 1
			if p := token.Property(types.IncludeProp); p.Value != nil {
				include, _ = p.Value.(int)
			}

			value := p.readWords(include, token.Delimeter())

			// --- apply Properties ---
			for _, prop := range token.Properties() {
				pVal, _ := prop.Value.(string)
				switch prop.Type {
				case types.TrimProp:
					value = strings.Trim(value, pVal)
				case types.TrimRightProp:
					value = strings.TrimRight(value, pVal)
				case types.TrimLeftProp:
					value = strings.TrimLeft(value, pVal)
				}
			}
			// --- --- --- ---

			// --- Assign ---
			switch field := conv.String(token.Value()); field {
			case "level":
				entry.Level = value
			case "message", "msg":
				entry.Message = value
			default:
				// TODO: Maybe add parser parameters which forces numbers
				if math.IsNumeric(value) {
					if entry.TagsNumber == nil {
						entry.TagsNumber = map[string]float64{}
					}
					entry.TagsNumber[field], _ = strconv.ParseFloat(value, 64)
					break
				}

				if entry.TagsString == nil {
					entry.TagsString = map[string]string{}
				}
				entry.TagsString[field] = value
			}
			// --- --- --- ---

		case types.Ignore:
			ignoreStr := token.Value()
			for i := 0; i < len(ignoreStr); i++ {
				if ignoreStr[i] != byte(p.Peek()) {
					return fmt.Errorf("invalid character. in token '%s' expected '%c', got '%c'", ignoreStr, ignoreStr[i], p.Peek())
				}
				p.Next()
			}
		}
	}

	return nil
}

func (p *Parser) readWords(i int, delim rune) string {
	sb := strings.Builder{}
	for {
		t := p.Peek()
		if t == scanner.EOF {
			break
		}
		if t == delim {
			i--
			if i == 0 {
				break
			}
		}

		sb.WriteRune(p.Next())
	}
	return sb.String()
}
