package types

import "fmt"

type Position struct {
	Line, Column int
}

func (p Position) String() string {
	return fmt.Sprintf("(line: %d, column: %d)", p.Line, p.Column)
}

type PropertyValue interface {
	int | string | any
}

type PropertyType string

const (
	IncludeProp   PropertyType = "include"
	TrimProp      PropertyType = "trim"
	TrimLeftProp  PropertyType = "trim_left"
	TrimRightProp PropertyType = "trim_right"
)

type TokenProperty struct {
	Type  PropertyType
	Value PropertyValue
}
