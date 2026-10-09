package world

type Enum interface {
	~int
	String() string
}

// Returns the enum value whose String() matches enumString.
// enumCount is the enum's _COUNT sentinel; values 0..enumCount-1 are checked.
func EnumFromString[T Enum](enumString string, enumCount T) (T, bool) {
	for value := range enumCount {
		if value.String() == enumString {
			return value, true
		}
	}

	return 0, false
}
