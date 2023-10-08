package bitmap

import "golang.org/x/exp/constraints"

func SetBit[T constraints.Unsigned, P ~uint8](n *T, pos P){
	*n |= (1 << pos)
}

func ClearBit[T constraints.Unsigned, P ~uint8](n *T, pos P) {
	*n &^= (1 << pos)
}

func Join[T constraints.Unsigned](a *T, b T) {
	*a |= b
}

func GetBit[T constraints.Unsigned, P ~uint8](n T, pos P) bool {
	return (n&(1<<pos) > 0)
}

func SetAll[T constraints.Unsigned](n *T) {
	*n = ^T(0)
}

func ClearAll[T constraints.Unsigned](n *T) {
	*n = 0
}
