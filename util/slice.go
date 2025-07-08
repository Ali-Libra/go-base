package util

func SliceContains[T string | int | int64 | uint32 | uint64](
	arr []T, target T) bool {
	for _, s := range arr {
		if s == target {
			return true
		}
	}
	return false
}
