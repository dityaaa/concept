package source

func notNilAssign[T any](dest, val *T) *T {
	if val != nil {
		*dest = *val
	}
	return dest
}
