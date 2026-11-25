// © 2019-present nextmv.io inc

package common

// Intersect returns the intersection of two slices.
func Intersect[T comparable](a []T, b []T) []T {
	set := make([]T, 0)

	if len(a) == 0 || len(b) == 0 {
		return set
	}

	ref := a
	other := b

	if len(a) > len(b) {
		ref = b
		other = a
	}

	hash := make(map[T]struct{}, len(ref))

	for _, v := range ref {
		hash[v] = struct{}{}
	}

	for _, v := range other {
		if _, ok := hash[v]; ok {
			set = append(set, v)
		}
	}

	return set
}
