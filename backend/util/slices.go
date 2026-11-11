package util

func RemoveAtUnordered[T any](slice []T, index int) []T {
	lastIndex := len(slice) - 1
	slice[index] = slice[lastIndex]
	return slice[:lastIndex]
}
