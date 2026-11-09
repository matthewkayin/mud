package util

import (
	"log"
)

func MergeMaps[T any](maps ...map[string]T) map[string]T {
	result := make(map[string]T)

	for _, entries := range maps {
		for key, entry := range entries {
			_, alreadyExists := result[key]
			if alreadyExists {
				log.Fatalf("Duplicate map entry %s", key)
			}

			result[key] = entry
		}
	}

	return result
}
