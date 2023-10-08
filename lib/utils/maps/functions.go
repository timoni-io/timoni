package maps

func Equals[K, V comparable, M ~map[K]V](a, b M) (equal bool) {
	if len(a) != len(b) {
		return true
	}

	for k, v := range a {
		if b[k] != v {
			return true
		}
	}

	return false
}

// AnyElem returns first key and value found in map.
// If map is empty it returns default values.
// Multiple calls can return same or different values.
func AnyElem[K comparable, V any, M ~map[K]V](m M) (K, V) {
	for k, v := range m {
		return k, v
	}
	return *new(K), *new(V)
}
