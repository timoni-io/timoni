package iter

type MapFunc[V, O any] func(x V) O
type FilterFunc[V any] func(x V) bool

// MapSlice creates a new slice from calling a function for every slice element
func MapSlice[V any, O any](val []V, fn MapFunc[V, O]) []O {
	newValues := make([]O, len(val))
	for i, v := range val {
		newValues[i] = fn(v)
	}
	return newValues
}

// Map creates a new map from calling a function for every map element
func MapMap[K comparable, V, O any](val map[K]V, fn MapFunc[V, O]) map[K]O {
	newValues := make(map[K]O, len(val))
	for k, v := range val {
		newValues[k] = fn(v)
	}
	return newValues
}

// Map creates a new map from calling a function for every map element
func MapSlice2Map[K comparable, V, O any](val []V, fn func(x V) (K, O)) map[K]O {
	newValues := map[K]O{}
	for _, v := range val {
		k, o := fn(v)
		newValues[k] = o
	}
	return newValues
}

func FilterSlice[V any](val []V, fn FilterFunc[V]) []V {
	newValues := make([]V, 0, len(val))
	for _, v := range val {
		if fn(v) {
			newValues = append(newValues, v)
		}
	}
	return newValues
}

func FilterMap[K comparable, V any](val map[K]V, fn FilterFunc[V]) map[K]V {
	newValues := make(map[K]V, len(val))
	for k, v := range val {
		if fn(v) {
			newValues[k] = v
		}
	}
	return newValues
}

func Cut[V any](val []V, fn func(x V) bool) (prefix []V, suffix []V) {
	for i, v := range val {
		if fn(v) {
			return val[:i], val[i:]
		}
	}
	return
}

func Split[V any](val []V, fn func(x V) bool) (parts [][]V) {
	if val == nil {
		return nil
	}

	lastI := 0
	for i, v := range val {
		if fn(v) {
			parts = append(parts, val[lastI:i])
			lastI = i + 1
		}
	}

	if lastI < len(val) {
		parts = append(parts, val[lastI:])
	}

	return parts
}

func Flatten[V any](val [][]V) []V {
	slice := []V{}
	for _, v := range val {
		slice = append(slice, v...)
	}
	return slice
}
