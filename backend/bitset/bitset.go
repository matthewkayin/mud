package bitset

func New(size int) []byte {
	sizeInBytes := size / 8
	if size % 8 != 0 {
		sizeInBytes++
	}
	bitset := make([]byte, sizeInBytes)
	clear(bitset)

	return bitset
}

func Check(bitset []byte, index int) bool {
	bucket := index / 8
	var bit byte = 1 << (index % 8)

	return (bitset[bucket] & bit) == bit
}

func Set(bitset []byte, index int, value bool) {
	bucket := index / 8
	var bit byte = 1 << (index % 8)

	if value {
		bitset[bucket] |= bit
	} else {
		bitset[bucket] &= ^bit
	}
}
