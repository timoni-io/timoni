package types

import (
	"lib/utils/iter"
	"lib/utils/maps"
	"lib/utils/scanner"
)

type TokenType uint8

const (
	Error TokenType = iota
	Ignore
	Tag
)

type Token interface {
	Type() TokenType
	Value() string
	Position() scanner.Position
	Delimeter() rune
	Property(PropertyType) TokenProperty
	Properties() []TokenProperty
}

type TokenObject struct {
	Typ TokenType
	Val string
	Pos scanner.Position

	Delim rune
	Props *maps.WeightedMap[PropertyType, TokenProperty]
}

func (t TokenObject) Type() TokenType {
	return t.Typ
}

func (t TokenObject) Value() string {
	return t.Val
}

func (t TokenObject) Position() scanner.Position {
	return t.Pos
}

func (t TokenObject) Delimeter() rune {
	return t.Delim
}

func (t TokenObject) Property(prop PropertyType) TokenProperty {
	return t.Props.Get(prop).Value
}

func (t TokenObject) Properties() []TokenProperty {
	return iter.MapSlice(t.Props.Values(), func(v maps.Weighted[TokenProperty]) TokenProperty { return v.Value })
}
