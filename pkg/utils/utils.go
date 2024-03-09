package utils

import (
	"bufio"
	"os"

	log "github.com/sirupsen/logrus"
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

func LogInit(logLevel log.Level) {
	// Log as JSON instead of the default ASCII formatter.
	if logLevel >= log.InfoLevel {
		log.Warn("Log level set to ", logLevel)
		log.SetReportCaller(true)
	}

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(logLevel)
}
