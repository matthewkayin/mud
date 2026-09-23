package util

import (
	"fmt"
	"os"
	"bufio"
)

func ReadFileLines(path string) ([]string, error) {
	lines := make([]string, 0, 1)

	// Open the file
	file, err := os.Open("./banner.txt")
	if err != nil {
		return lines, fmt.Errorf("Error opening banner text: %s", err.Error())
	}
	defer file.Close()


	// Read line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Check for any errors that occurred during scanning
	err = scanner.Err()
	if err != nil {
		return lines, fmt.Errorf("Error reading banner text: %s", err.Error())
	}

	return lines, nil
}
