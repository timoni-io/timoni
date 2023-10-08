package scanner

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"text/scanner"
)

var (
	// Read Line
	fnRL = func(r rune) bool { return r == '\n' }
	// Read Word
	fnRW = func(r rune) bool { return r == ' ' || r == '\n' }
)

type Position struct {
	Line, Column int
}

func (p Position) String() string {
	return fmt.Sprintf("(line: %d, column: %d)", p.Line, p.Column)
}

type Scanner struct {
	bufio.Scanner
	Position Position
}

func NewScanner(rd io.Reader) *Scanner {
	s := bufio.NewScanner(rd)
	s.Split(bufio.ScanRunes)
	s.Scan()
	return &Scanner{
		Scanner:  *s,
		Position: Position{Line: 1, Column: 1},
	}
}

func (sc *Scanner) End() bool {
	return sc.Peek() == scanner.EOF
}

func (sc *Scanner) Peek() rune {
	txt := sc.Scanner.Text()
	if txt == "" {
		return -1
	}
	return rune(txt[0])
}

func (sc *Scanner) Next() rune {
	r := sc.Peek()
	if r != -1 {
		sc.Scanner.Scan()
	}

	if r == '\n' {
		sc.Position.Line += 1
		sc.Position.Column = 1
	} else {
		sc.Position.Column += 1
	}

	return r
}

func (sc Scanner) PeekEscaped() rune {
	r := sc.Next()
	if r == '\\' {
		r = sc.Peek()
	}
	return r
}

func (sc *Scanner) NextEscaped() rune {
	r := sc.Next()
	if r == '\\' {
		r = sc.Next()
	}
	return r
}

// Scanner holds found rune
func (sc *Scanner) ReadUntil(until func(rune) bool) string {
	sb := strings.Builder{}

	for t := sc.PeekEscaped(); t != scanner.EOF; t = sc.PeekEscaped() {
		if until(t) {
			break
		}

		sb.WriteRune(t)
		sc.NextEscaped()
	}
	return sb.String()
}

// Scanner holds next rune after found string
func (sc *Scanner) ReadUntilStr(until ...string) string {
	sb := strings.Builder{}
	for t := sc.NextEscaped(); t != scanner.EOF; t = sc.NextEscaped() {
		for _, x := range until {
			if strings.HasSuffix(sb.String(), x) {
				return strings.TrimSuffix(sb.String(), x)
			}
		}
		sb.WriteRune(t)
	}
	return sb.String()
}

func (sc *Scanner) ReadEnd() string {
	sb := strings.Builder{}

	for t := sc.Next(); t != scanner.EOF; t = sc.Next() {
		sb.WriteRune(t)
	}
	return sb.String()
}

func (sc *Scanner) ReadLine() string {
	data := sc.ReadUntil(fnRL)
	// Eat \n
	sc.Next()
	return data
}

func (sc *Scanner) ReadWord() string {
	data := sc.ReadUntil(fnRW)
	// Eat ' '
	sc.Next()
	return data
}
