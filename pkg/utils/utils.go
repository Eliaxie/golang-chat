package utils

import (
	"bufio"
	"os"
)

// function to read array of strings from file
func ReadConnectionsFromFile(filename string) ([]string, error) {
	// read file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// read file line by line
	var connections []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		connections = append(connections, scanner.Text())
	}

	return connections, nil
}
