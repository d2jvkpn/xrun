package internal

type Comparable interface {
	string | int | int32 | int64
}

func indexOf[T Comparable](list []T, e T) (idx int) {
	idx = -1

	for i := range list {
		if list[i] == e {
			idx = i
			break
		}
	}

	return
}
