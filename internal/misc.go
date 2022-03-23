package internal

type Comparable interface {
	bool | uint | uint32 | uint64 | int | int32 | int64 | string
}

func indexOf[T Comparable](list []T, e T) int {
	for i := range list {
		if list[i] == e {
			return i
		}
	}

	return -1
}
